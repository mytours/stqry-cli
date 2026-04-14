package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestCodesListCmd asserts that `stqry codes list` prints the coupon code
// returned by the API.
func TestCodesListCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/public/codes" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"codes": []interface{}{
				map[string]interface{}{
					"id":          "1",
					"coupon_code": "WELCOME10",
					"linked_type": "Collection",
					"linked_id":   42,
					"status":      "active",
				},
			},
			"meta": map[string]interface{}{
				"page": 1, "pages": 1, "per_page": 30, "count": 1,
			},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "codes", "list"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, "WELCOME10") {
		t.Errorf("expected output to contain %q, got:\n%s", "WELCOME10", out)
	}
}

// TestCodesGetCmd asserts that `stqry codes get 10` prints the coupon code
// returned by the API.
func TestCodesGetCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/public/codes/10" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": map[string]interface{}{
				"id":          "10",
				"coupon_code": "SUMMER25",
				"status":      "active",
			},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "codes", "get", "10"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, "SUMMER25") {
		t.Errorf("expected output to contain %q, got:\n%s", "SUMMER25", out)
	}
}

// TestCodesCreateMissingCouponCode asserts that `stqry codes create` without
// --coupon-code returns an error without hitting the API.
func TestCodesCreateMissingCouponCode(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "codes", "create", "--linked-type=Collection", "--linked-id=42", "--project-id=1"})
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --coupon-code, got nil")
	}
	if !contains(err.Error(), "--coupon-code") {
		t.Errorf("expected error to mention --coupon-code, got %q", err.Error())
	}
	if called {
		t.Error("API should not be called when --coupon-code is missing")
	}
}

// TestCodesCreateMissingLinkedType asserts that `stqry codes create` without
// --linked-type returns an error without hitting the API.
func TestCodesCreateMissingLinkedType(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "codes", "create", "--coupon-code=WELCOME10", "--linked-id=42", "--project-id=1"})
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --linked-type, got nil")
	}
	if !contains(err.Error(), "--linked-type") {
		t.Errorf("expected error to mention --linked-type, got %q", err.Error())
	}
	if called {
		t.Error("API should not be called when --linked-type is missing")
	}
}

// TestCodesCreateMissingLinkedID asserts that `stqry codes create` without
// --linked-id returns an error without hitting the API.
func TestCodesCreateMissingLinkedID(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "codes", "create", "--coupon-code=WELCOME10", "--linked-type=Collection", "--project-id=1"})
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --linked-id, got nil")
	}
	if !contains(err.Error(), "--linked-id") {
		t.Errorf("expected error to mention --linked-id, got %q", err.Error())
	}
	if called {
		t.Error("API should not be called when --linked-id is missing")
	}
}

// TestCodesCreateMissingProjectID asserts that `stqry codes create` without
// --project-id returns an error without hitting the API.
func TestCodesCreateMissingProjectID(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "codes", "create", "--coupon-code=WELCOME10", "--linked-type=Collection", "--linked-id=42"})
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --project-id, got nil")
	}
	if !contains(err.Error(), "--project-id") {
		t.Errorf("expected error to mention --project-id, got %q", err.Error())
	}
	if called {
		t.Error("API should not be called when --project-id is missing")
	}
}

// TestCodesCreateCmd asserts that `stqry codes create` with all required flags
// sends the correct fields in the POST request body.
func TestCodesCreateCmd(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/codes" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": map[string]interface{}{
				"id":          "1",
				"coupon_code": "WELCOME10",
				"linked_type": "Collection",
				"linked_id":   42,
				"project_id":  1,
				"status":      "active",
			},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "codes", "create",
		"--coupon-code=WELCOME10", "--linked-type=Collection", "--linked-id=42", "--project-id=1"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["coupon_code"] != "WELCOME10" {
		t.Errorf("expected coupon_code=%q in request body, got %v", "WELCOME10", captured["coupon_code"])
	}
	if captured["linked_type"] != "Collection" {
		t.Errorf("expected linked_type=%q in request body, got %v", "Collection", captured["linked_type"])
	}
}

// TestCodesUpdateNoFields asserts that `stqry codes update 10` with no flags
// returns an error containing "no fields".
func TestCodesUpdateNoFields(t *testing.T) {
	setupTestHome(t, "http://localhost:0")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "codes", "update", "10"})
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no update fields specified, got nil")
	}
	if !contains(err.Error(), "no fields") {
		t.Errorf("expected error to contain \"no fields\", got %q", err.Error())
	}
}

// TestCodesUpdateCmd asserts that `stqry codes update 10 --coupon-code=NEWCODE`
// sends the correct field in the PATCH request body.
func TestCodesUpdateCmd(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/api/public/codes/10" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": map[string]interface{}{
				"id":          "10",
				"coupon_code": "NEWCODE",
				"status":      "active",
			},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "codes", "update", "10", "--coupon-code=NEWCODE"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["coupon_code"] != "NEWCODE" {
		t.Errorf("expected coupon_code=%q in request body, got %v", "NEWCODE", captured["coupon_code"])
	}
}

// TestCodesDeleteCmd asserts that `stqry codes delete 10` sends a DELETE
// request to the correct endpoint and prints a confirmation.
func TestCodesDeleteCmd(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/api/public/codes/10" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		} else {
			deleted = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	origStdout := os.Stdout
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = wPipe

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "codes", "delete", "10"})
	cmd.SetErr(os.Stderr)
	execErr := cmd.Execute()

	wPipe.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(rPipe)
	rPipe.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !deleted {
		t.Error("expected DELETE request to have been made, but it was not")
	}
	if !contains(out, "Deleted") {
		t.Errorf("expected output to contain %q, got:\n%s", "Deleted", out)
	}
}
