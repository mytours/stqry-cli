package cli

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mytours/stqry-cli/internal/config"
)

type checkStatus string

const (
	statusPass checkStatus = "pass"
	statusFail checkStatus = "fail"
	statusSkip checkStatus = "skip"
	statusInfo checkStatus = "info"
	statusWarn checkStatus = "warn"
)

type checkResult struct {
	group    string
	name     string
	status   checkStatus
	message  string
	detail   string
	duration time.Duration
}

func checkGlobalConfig(configPath string) checkResult {
	start := time.Now()
	r := checkResult{
		group: "Config",
		name:  "Global config exists",
	}
	if _, err := os.Stat(configPath); err != nil {
		r.status = statusFail
		r.message = "~/.config/stqry/config.yaml not found"
		r.detail = fmt.Sprintf("Looked for: %s", configPath)
	} else {
		r.status = statusPass
		r.message = "Global config found"
		r.detail = fmt.Sprintf("Path: %s", configPath)
	}
	r.duration = time.Since(start)
	return r
}

func checkDirectoryConfig(cwd string) checkResult {
	start := time.Now()
	r := checkResult{
		group: "Config",
		name:  "Directory config found",
	}
	dirCfg, err := config.FindDirectoryConfig(cwd)
	if err != nil || dirCfg == nil || (dirCfg.Site == "" && dirCfg.Token == "" && dirCfg.APIURL == "") {
		r.status = statusFail
		r.message = "No stqry.yaml found in current directory or parents"
		r.detail = fmt.Sprintf("Looked up from: %s", cwd)
	} else {
		r.status = statusPass
		r.message = "stqry.yaml found"
		r.detail = fmt.Sprintf("Searched from: %s", cwd)
	}
	r.duration = time.Since(start)
	return r
}

func checkSiteResolved(globalCfg *config.GlobalConfig, flagSite string, dirCfg *config.DirectoryConfig) checkResult {
	start := time.Now()
	r := checkResult{
		group: "Config",
		name:  "Site resolved",
	}
	site, err := config.ResolveSite(globalCfg, flagSite, dirCfg)
	if err != nil {
		r.status = statusFail
		r.message = "Site could not be resolved"
		r.detail = err.Error()
	} else {
		r.status = statusPass
		r.message = fmt.Sprintf("Site resolved → %s", site.APIURL)
		r.detail = fmt.Sprintf("API URL: %s", site.APIURL)
	}
	r.duration = time.Since(start)
	return r
}

func checkAPIReachable(baseURL string, httpClient *http.Client) checkResult {
	start := time.Now()
	r := checkResult{group: "API", name: "API reachable"}

	if _, err := url.ParseRequestURI(baseURL); err != nil {
		r.status = statusFail
		r.message = fmt.Sprintf("Malformed API URL: %s", baseURL)
		r.duration = time.Since(start)
		return r
	}

	resp, err := httpClient.Get(baseURL)
	r.duration = time.Since(start)
	if err != nil {
		r.status = statusFail
		r.message = fmt.Sprintf("Cannot reach %s: %v", baseURL, err)
		return r
	}
	resp.Body.Close()

	host := hostFromURL(baseURL)
	r.status = statusPass
	r.message = fmt.Sprintf("Reachable (%s)", host)
	r.detail = fmt.Sprintf("URL: %s → HTTP %d", baseURL, resp.StatusCode)
	return r
}

func checkTokenValid(baseURL, token string, httpClient *http.Client) checkResult {
	start := time.Now()
	r := checkResult{group: "API", name: "Token valid"}

	reqURL := strings.TrimRight(baseURL, "/") + "/api/v3/collections?per_page=1"
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		r.status = statusFail
		r.message = fmt.Sprintf("Building request: %v", err)
		r.duration = time.Since(start)
		return r
	}
	req.Header.Set("X-Api-Token", token)

	resp, err := httpClient.Do(req)
	r.duration = time.Since(start)
	if err != nil {
		r.status = statusFail
		r.message = fmt.Sprintf("Request failed: %v", err)
		return r
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		r.status = statusFail
		r.message = fmt.Sprintf("Token rejected (HTTP %d)", resp.StatusCode)
		r.detail = fmt.Sprintf("GET /api/v3/collections → HTTP %d", resp.StatusCode)
		return r
	}

	r.status = statusPass
	r.message = "Token accepted"
	r.detail = fmt.Sprintf("GET /api/v3/collections → HTTP %d", resp.StatusCode)
	return r
}

func checkRegion(apiURL string) checkResult {
	r := checkResult{group: "API", name: "Region", status: statusInfo}
	host := hostFromURL(apiURL)
	if strings.HasPrefix(host, "api-") {
		parts := strings.SplitN(host, ".", 2)
		region := strings.TrimPrefix(parts[0], "api-")
		r.message = fmt.Sprintf("Region: %s", region)
		r.detail = fmt.Sprintf("Full URL: %s", apiURL)
		return r
	}
	r.message = fmt.Sprintf("Region: %s", host)
	r.detail = fmt.Sprintf("Full URL: %s", apiURL)
	return r
}

// hostFromURL extracts just the host portion of a URL.
func hostFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Host
}

func doctorSymbol(s checkStatus) string {
	switch s {
	case statusPass:
		return "✓"
	case statusFail:
		return "✗"
	case statusSkip:
		return "-"
	case statusInfo:
		return "ℹ"
	case statusWarn:
		return "⚠"
	default:
		return "?"
	}
}
