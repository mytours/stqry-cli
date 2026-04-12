package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// uploadClientTimeout is the total request budget for the S3 POST in the
// upload flow. It is deliberately much larger than Client.HTTPClient's 30 s
// default, because the timer covers body write + S3 commit + response
// headers, and a medium-size file on a slow connection can easily exceed 30 s.
const uploadClientTimeout = 10 * time.Minute

// PresignResponse holds the presigned S3 URL and form fields.
type PresignResponse struct {
	URL    string            `json:"url"`
	Fields map[string]string `json:"fields"`
}

// EnqueueResponse holds the job ID returned by the enqueue endpoint.
type EnqueueResponse struct {
	JobID string `json:"job_id"`
}

// ProcessStatusResponse holds the polling response from the process_status endpoint.
type ProcessStatusResponse struct {
	Status       string                 `json:"status"`
	PctComplete  int                    `json:"pct_complete"`
	Message      string                 `json:"message"`
	UploadedFile map[string]interface{} `json:"uploaded_file"`
}

// progressReader wraps an io.Reader and calls onProgress with bytes read so far.
type progressReader struct {
	r          io.Reader
	total      int64
	written    int64
	onProgress func(written, total int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.written += int64(n)
	if pr.onProgress != nil {
		pr.onProgress(pr.written, pr.total)
	}
	return n, err
}

// UploadFile performs the full multi-step upload flow:
//  1. Presign — POST /api/public/uploaded_files/presigned
//  2. S3 upload — multipart POST to presign URL (or s3BaseURL+"/upload" for tests)
//  3. Enqueue — POST /api/public/uploaded_files/process_enqueue
//  4. Poll — GET /api/public/uploaded_files/process_status/{job_id}
//
// Returns the uploaded_file map when complete.
// s3BaseURL is only used in tests; pass "" in production.
// onStatus, if non-nil, is called with each new server-side processing message
// during the polling phase. Callers are responsible for rendering these (e.g.
// printing to stdout), so that library code does not write to stdout directly.
func UploadFile(c *Client, filePath string, s3BaseURL string, onProgress func(written, total int64), onStatus func(string)) (map[string]interface{}, error) {
	basename := filepath.Base(filePath)

	// Step 1: Presign.
	var presign PresignResponse
	if err := c.Post("/api/public/uploaded_files/presigned", map[string]interface{}{"filename": basename}, &presign); err != nil {
		return nil, fmt.Errorf("presign: %w", err)
	}

	// Step 2: Build multipart body and upload to S3.
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}
	fileSize := fi.Size()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	// Add presign fields first (order matters for S3).
	for k, v := range presign.Fields {
		if err := mw.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("writing field %s: %w", k, err)
		}
	}

	// Add the file part.
	part, err := mw.CreateFormFile("file", basename)
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("copying file: %w", err)
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	uploadURL := presign.URL
	if s3BaseURL != "" {
		uploadURL = s3BaseURL + "/upload"
	}

	bodyBytes := buf.Bytes()
	totalSize := int64(len(bodyBytes))
	var bodyReader io.Reader = bytes.NewReader(bodyBytes)
	if onProgress != nil {
		bodyReader = &progressReader{r: bodyReader, total: totalSize, onProgress: func(written, total int64) {
			// Scale progress to reflect actual file bytes, not multipart overhead.
			pct := float64(written) / float64(total)
			onProgress(int64(pct*float64(fileSize)), fileSize)
		}}
	}

	uploadReq, err := http.NewRequest("POST", uploadURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating upload request: %w", err)
	}
	// Explicitly set Content-Length because http.NewRequest only auto-detects
	// it for *bytes.Reader / *bytes.Buffer / *strings.Reader, and wrapping the
	// reader in progressReader hides that concrete type. Without this the
	// request goes out with Transfer-Encoding: chunked, which real AWS S3
	// rejects with 411 Length Required.
	uploadReq.ContentLength = totalSize
	uploadReq.Header.Set("Content-Type", mw.FormDataContentType())

	// Use a dedicated HTTP client with a much longer total timeout for the S3
	// upload. c.HTTPClient has a 30 s Client.Timeout which is fine for normal
	// API calls but too tight for multi-megabyte uploads on slow connections —
	// the 30 s budget has to cover the entire body write plus reading S3's
	// response headers, and real users have hit "Client.Timeout exceeded while
	// awaiting headers" on ~6 MB files.
	// Inherit the transport from the parent client so that any custom proxy,
	// TLS, or middleware configuration is preserved for the S3 request too.
	uploadClient := &http.Client{Timeout: uploadClientTimeout, Transport: c.HTTPClient.Transport}
	uploadResp, err := uploadClient.Do(uploadReq)
	if err != nil {
		return nil, fmt.Errorf("uploading to S3: %w", err)
	}
	defer uploadResp.Body.Close()
	io.ReadAll(uploadResp.Body) // drain

	if uploadResp.StatusCode >= 400 {
		return nil, fmt.Errorf("S3 upload failed with status %d", uploadResp.StatusCode)
	}

	// Step 3: Enqueue. The Rails job passes s3_url straight to
	// S3Utils.download which calls fog files.get(remote_path) — remote_path
	// must be the S3 object key, not a full URL. See
	// app/sidekiq/uploaded_file_process.rb and
	// spec/sidekiq/uploaded_file_process_spec.rb in mytours-web.
	key := presign.Fields["key"]
	if key == "" {
		return nil, fmt.Errorf("presign response missing key field")
	}
	var enqueue EnqueueResponse
	if err := c.Post("/api/public/uploaded_files/process_enqueue", map[string]interface{}{"s3_url": key}, &enqueue); err != nil {
		return nil, fmt.Errorf("enqueue: %w", err)
	}

	// Step 4: Poll for completion. Surface progress messages so the user can
	// see what stage the server-side job is in.
	deadline := time.Now().Add(5 * time.Minute)
	var lastMessage string
	for time.Now().Before(deadline) {
		var status ProcessStatusResponse
		path := fmt.Sprintf("/api/public/uploaded_files/process_status/%s", enqueue.JobID)
		if err := c.Get(path, nil, &status); err != nil {
			return nil, fmt.Errorf("polling status: %w", err)
		}

		if status.Message != "" && status.Message != lastMessage {
			lastMessage = status.Message
			if onStatus != nil {
				onStatus(status.Message)
			}
		}

		switch status.Status {
		case "complete":
			return status.UploadedFile, nil
		case "failed", "error", "transcoder_error", "transcoder_invalid_file", "interrupted":
			msg := status.Message
			if msg == "" {
				msg = status.Status
			}
			return nil, fmt.Errorf("upload processing failed: %s", msg)
		}

		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("upload processing timed out after 5 minutes")
}

