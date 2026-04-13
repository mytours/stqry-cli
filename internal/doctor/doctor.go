// Package doctor contains the exported check logic for the stqry doctor command.
// The check functions here are intentionally parallel to the unexported versions in
// internal/cli/doctor.go (which are tested there). If you fix a bug in one copy,
// mirror the fix in the other.
package doctor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mytours/stqry-cli/internal/config"
	"github.com/mytours/stqry-cli/internal/skills"
)

type CheckStatus string

const (
	StatusPass CheckStatus = "pass"
	StatusFail CheckStatus = "fail"
	StatusSkip CheckStatus = "skip"
	StatusInfo CheckStatus = "info"
	StatusWarn CheckStatus = "warn"
)

type CheckResult struct {
	Group    string
	Name     string
	Status   CheckStatus
	Message  string
	Detail   string
	Duration time.Duration
}

type RunResult struct {
	Checks    []CheckResult
	AnyFailed bool
}

const DefaultGitHubReleasesURL = "https://api.github.com/repos/mytours/stqry-cli/releases/latest"

func Symbol(s CheckStatus) string {
	switch s {
	case StatusPass:
		return "✓"
	case StatusFail:
		return "✗"
	case StatusSkip:
		return "-"
	case StatusInfo:
		return "ℹ"
	case StatusWarn:
		return "⚠"
	default:
		return "?"
	}
}

func CheckGlobalConfig(configPath string) CheckResult {
	start := time.Now()
	r := CheckResult{Group: "Config", Name: "Global config exists"}
	if _, err := os.Stat(configPath); err != nil {
		r.Status = StatusFail
		r.Detail = fmt.Sprintf("Looked for: %s", configPath)
	} else {
		r.Status = StatusPass
		r.Message = configPath
		r.Detail = fmt.Sprintf("Path: %s", configPath)
	}
	r.Duration = time.Since(start)
	return r
}

func CheckDirectoryConfig(cwd string) CheckResult {
	start := time.Now()
	r := CheckResult{Group: "Config", Name: "Directory config found"}
	dirCfg, err := config.FindDirectoryConfig(cwd)
	if err != nil || dirCfg == nil || (dirCfg.Site == "" && dirCfg.Token == "" && dirCfg.APIURL == "") {
		r.Status = StatusFail
		r.Detail = fmt.Sprintf("Looked up from: %s", cwd)
	} else {
		r.Status = StatusPass
		r.Detail = fmt.Sprintf("Searched from: %s", cwd)
	}
	r.Duration = time.Since(start)
	return r
}

func CheckSiteResolved(globalCfg *config.GlobalConfig, flagSite string, dirCfg *config.DirectoryConfig) (CheckResult, *config.Site) {
	start := time.Now()
	r := CheckResult{Group: "Config", Name: "Site resolved"}
	site, err := config.ResolveSite(globalCfg, flagSite, dirCfg)
	if err != nil {
		r.Status = StatusFail
		r.Message = "Site could not be resolved"
		r.Detail = err.Error()
	} else {
		r.Status = StatusPass
		r.Message = site.APIURL
		r.Detail = fmt.Sprintf("API URL: %s", site.APIURL)
	}
	r.Duration = time.Since(start)
	return r, site
}

func CheckAPIReachable(baseURL string, httpClient *http.Client) CheckResult {
	start := time.Now()
	r := CheckResult{Group: "API", Name: "API reachable"}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		r.Status = StatusFail
		r.Message = fmt.Sprintf("Malformed API URL: %s", baseURL)
		r.Duration = time.Since(start)
		return r
	}
	resp, err := httpClient.Get(baseURL)
	r.Duration = time.Since(start)
	if err != nil {
		r.Status = StatusFail
		r.Message = fmt.Sprintf("Cannot reach %s: %v", baseURL, err)
		return r
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	host := hostFromURL(baseURL)
	r.Status = StatusPass
	r.Message = host
	r.Detail = fmt.Sprintf("URL: %s → HTTP %d", baseURL, resp.StatusCode)
	return r
}

func CheckTokenValid(baseURL, token string, httpClient *http.Client) CheckResult {
	start := time.Now()
	r := CheckResult{Group: "API", Name: "Token valid"}
	reqURL := strings.TrimRight(baseURL, "/") + "/api/v3/collections?per_page=1"
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		r.Status = StatusFail
		r.Message = fmt.Sprintf("Building request: %v", err)
		r.Duration = time.Since(start)
		return r
	}
	req.Header.Set("X-Api-Token", token)
	resp, err := httpClient.Do(req)
	r.Duration = time.Since(start)
	if err != nil {
		r.Status = StatusFail
		r.Message = fmt.Sprintf("Request failed: %v", err)
		return r
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		r.Status = StatusFail
		r.Message = fmt.Sprintf("Token rejected (HTTP %d)", resp.StatusCode)
		r.Detail = fmt.Sprintf("GET /api/v3/collections → HTTP %d", resp.StatusCode)
		return r
	}
	r.Status = StatusPass
	r.Detail = fmt.Sprintf("GET /api/v3/collections → HTTP %d", resp.StatusCode)
	return r
}

func CheckRegion(apiURL string) CheckResult {
	r := CheckResult{Group: "API", Name: "Region", Status: StatusInfo}
	host := hostFromURL(apiURL)
	if strings.HasPrefix(host, "api-") {
		parts := strings.SplitN(host, ".", 2)
		region := strings.TrimPrefix(parts[0], "api-")
		r.Message = region
		r.Detail = fmt.Sprintf("Full URL: %s", apiURL)
		return r
	}
	r.Message = host
	r.Detail = fmt.Sprintf("Full URL: %s", apiURL)
	return r
}

func CheckCLIVersion(currentVersion string, releasesURL string, httpClient *http.Client) CheckResult {
	start := time.Now()
	r := CheckResult{Group: "Version", Name: "CLI version"}
	if currentVersion == "dev" {
		r.Status = StatusInfo
		r.Message = "Running development build, skipping version check"
		r.Duration = time.Since(start)
		return r
	}
	if releasesURL == "" {
		releasesURL = DefaultGitHubReleasesURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := httpClient.Get(releasesURL)
	r.Duration = time.Since(start)
	if err != nil {
		r.Status = StatusWarn
		r.Message = "Could not check version (GitHub unreachable)"
		r.Detail = err.Error()
		return r
	}
	defer resp.Body.Close()
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		r.Status = StatusWarn
		r.Message = "Could not parse GitHub release response"
		r.Detail = err.Error()
		return r
	}
	if payload.TagName == "" {
		r.Status = StatusWarn
		r.Message = "Could not parse GitHub release response"
		return r
	}
	if payload.TagName == currentVersion {
		r.Status = StatusPass
		r.Message = fmt.Sprintf("CLI is up to date (%s)", currentVersion)
		r.Detail = fmt.Sprintf("Current: %s = Latest: %s", currentVersion, payload.TagName)
	} else {
		r.Status = StatusWarn
		r.Message = fmt.Sprintf("Update available: %s → %s", currentVersion, payload.TagName)
		r.Detail = fmt.Sprintf("Current: %s → Latest: %s\nRun: brew upgrade stqry (or download from GitHub releases)", currentVersion, payload.TagName)
	}
	return r
}

func PrintResults(w io.Writer, results []CheckResult, verbose bool) {
	var currentGroup string
	for _, r := range results {
		if r.Group != currentGroup {
			if currentGroup != "" {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "%s\n", r.Group)
			currentGroup = r.Group
		}
		sym := Symbol(r.Status)
		if verbose {
			fmt.Fprintf(w, "  %s %s (%s)\n", sym, r.Name, r.Duration.Round(time.Millisecond))
			if r.Detail != "" {
				for _, line := range strings.Split(r.Detail, "\n") {
					fmt.Fprintf(w, "    %s\n", line)
				}
			}
		} else {
			fmt.Fprintf(w, "  %s %s", sym, r.Name)
			if r.Message != "" {
				fmt.Fprintf(w, " (%s)", r.Message)
			}
			fmt.Fprintln(w)
		}
	}
}

// RunChecks executes all diagnostic checks and returns the results.
func RunChecks(currentVersion string) RunResult {
	httpClient := &http.Client{Timeout: 15 * time.Second}
	var results []CheckResult

	configPath := config.DefaultGlobalConfigPath()
	results = append(results, CheckGlobalConfig(configPath))

	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		results = append(results, CheckResult{
			Group:  "Config",
			Name:   "Directory config found",
			Status: StatusFail,
			Detail: fmt.Sprintf("Could not determine working directory: %v", cwdErr),
		})
	} else {
		results = append(results, CheckDirectoryConfig(cwd))
	}

	globalCfg, _ := config.LoadGlobalConfig(configPath)
	if globalCfg == nil {
		globalCfg = &config.GlobalConfig{Sites: make(map[string]*config.Site)}
	}
	var dirCfg *config.DirectoryConfig
	if cwd != "" {
		dirCfg, _ = config.FindDirectoryConfig(cwd)
	}

	siteResult, resolvedSite := CheckSiteResolved(globalCfg, "", dirCfg)
	results = append(results, siteResult)

	if siteResult.Status == StatusPass {
		results = append(results, CheckAPIReachable(resolvedSite.APIURL, httpClient))
		results = append(results, CheckTokenValid(resolvedSite.APIURL, resolvedSite.Token, httpClient))
		results = append(results, CheckRegion(resolvedSite.APIURL))
	} else {
		results = append(results,
			CheckResult{Group: "API", Name: "API reachable", Status: StatusSkip, Message: "No site resolved"},
			CheckResult{Group: "API", Name: "Token valid", Status: StatusSkip, Message: "No site resolved"},
			CheckResult{Group: "API", Name: "Region", Status: StatusSkip, Message: "No site resolved"},
		)
	}

	results = append(results, CheckCLIVersion(currentVersion, DefaultGitHubReleasesURL, httpClient))

	// Skills checks
	home, _ := os.UserHomeDir()
	skillLocations := []SkillLocation{
		{Dir: filepath.Join(cwd, ".claude", "commands"), Layout: SkillLayoutCode, Label: "Claude Code (local)"},
		{Dir: filepath.Join(home, ".claude", "commands"), Layout: SkillLayoutCode, Label: "Claude Code (global)"},
		{Dir: skills.DesktopSkillsDir(), Layout: SkillLayoutDesktop, Label: "Claude Desktop"},
	}
	results = append(results, CheckInstalledSkills(skillLocations)...)

	anyFailed := false
	for _, r := range results {
		if r.Status == StatusFail {
			anyFailed = true
			break
		}
	}
	return RunResult{Checks: results, AnyFailed: anyFailed}
}

func hostFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Host
}
