// Check functions in this file are intentionally parallel to the exported versions
// in internal/doctor/doctor.go (used by the MCP tool). If you fix a bug here,
// mirror the fix there too.
package cli

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

	"github.com/mytours/stqry-cli/internal/buildinfo"
	"github.com/mytours/stqry-cli/internal/config"
	"github.com/mytours/stqry-cli/internal/skills"
	"github.com/spf13/cobra"
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
		r.message = ""
		r.detail = fmt.Sprintf("Looked for: %s", configPath)
	} else {
		r.status = statusPass
		r.message = configPath
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
		r.message = ""
		r.detail = fmt.Sprintf("Looked up from: %s", cwd)
	} else {
		r.status = statusPass
		r.message = ""
		r.detail = fmt.Sprintf("Searched from: %s", cwd)
	}
	r.duration = time.Since(start)
	return r
}

// checkSiteResolved returns the check result and the resolved site (nil on failure).
func checkSiteResolved(globalCfg *config.GlobalConfig, flagSite string, dirCfg *config.DirectoryConfig) (checkResult, *config.Site) {
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
		r.message = site.APIURL
		r.detail = fmt.Sprintf("API URL: %s", site.APIURL)
	}
	r.duration = time.Since(start)
	return r, site
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
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	host := hostFromURL(baseURL)
	r.status = statusPass
	r.message = host
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
	r.message = ""
	r.detail = fmt.Sprintf("GET /api/v3/collections → HTTP %d", resp.StatusCode)
	return r
}

func checkRegion(apiURL string) checkResult {
	r := checkResult{group: "API", name: "Region", status: statusInfo}
	host := hostFromURL(apiURL)
	if strings.HasPrefix(host, "api-") {
		parts := strings.SplitN(host, ".", 2)
		region := strings.TrimPrefix(parts[0], "api-")
		r.message = region
		r.detail = fmt.Sprintf("Full URL: %s", apiURL)
		return r
	}
	r.message = host
	r.detail = fmt.Sprintf("Full URL: %s", apiURL)
	return r
}

const defaultGitHubReleasesURL = "https://api.github.com/repos/mytours/stqry-cli/releases/latest"

func checkCLIVersion(currentVersion string, releasesURL string, httpClient *http.Client) checkResult {
	start := time.Now()
	r := checkResult{group: "Version", name: "CLI version"}

	if currentVersion == "dev" {
		r.status = statusInfo
		r.message = "Running development build, skipping version check"
		r.duration = time.Since(start)
		return r
	}

	if releasesURL == "" {
		releasesURL = defaultGitHubReleasesURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := httpClient.Get(releasesURL)
	r.duration = time.Since(start)
	if err != nil {
		r.status = statusWarn
		r.message = "Could not check version (GitHub unreachable)"
		r.detail = err.Error()
		return r
	}
	defer resp.Body.Close()

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		r.status = statusWarn
		r.message = "Could not parse GitHub release response"
		r.detail = err.Error()
		return r
	}
	if payload.TagName == "" {
		r.status = statusWarn
		r.message = "Could not parse GitHub release response"
		return r
	}

	// GitHub tags use a "v" prefix (e.g. "v0.6.2"); strip it to match
	// the version baked into the binary by GoReleaser (e.g. "0.6.2").
	latestVersion := strings.TrimPrefix(payload.TagName, "v")
	if latestVersion == currentVersion {
		r.status = statusPass
		r.message = fmt.Sprintf("CLI is up to date (%s)", currentVersion)
		r.detail = fmt.Sprintf("Current: %s = Latest: %s", currentVersion, latestVersion)
	} else {
		r.status = statusWarn
		r.message = fmt.Sprintf("Update available: %s → %s", currentVersion, latestVersion)
		r.detail = fmt.Sprintf("Current: %s → Latest: %s\nRun: brew upgrade stqry (or download from GitHub releases)", currentVersion, latestVersion)
	}
	return r
}

func printDoctorResults(w io.Writer, results []checkResult, verbose bool) {
	var currentGroup string
	for _, r := range results {
		if r.group != currentGroup {
			if currentGroup != "" {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "%s\n", r.group)
			currentGroup = r.group
		}

		sym := doctorSymbol(r.status)
		if verbose {
			fmt.Fprintf(w, "  %s %s (%s)\n", sym, r.name, r.duration.Round(time.Millisecond))
			if r.detail != "" {
				for _, line := range strings.Split(r.detail, "\n") {
					fmt.Fprintf(w, "    %s\n", line)
				}
			}
		} else {
			fmt.Fprintf(w, "  %s %s", sym, r.name)
			if r.message != "" {
				fmt.Fprintf(w, " (%s)", r.message)
			}
			fmt.Fprintln(w)
		}
	}
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

func checkInstalledSkills() []checkResult {
	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()

	var localSkillDir string
	if cwd != "" {
		localSkillDir = filepath.Join(cwd, ".claude", "commands")
	}

	type loc struct {
		dir   string
		label string
	}
	locations := []loc{
		{dir: localSkillDir, label: "Claude Code (local)"},
		{dir: filepath.Join(home, ".claude", "commands"), label: "Claude Code (global)"},
	}

	skillNames, err := skills.EmbeddedSkillNames()
	if err != nil {
		return []checkResult{{group: "Skills", name: "Embedded skills", status: statusFail, message: err.Error()}}
	}

	var results []checkResult
	for _, l := range locations {
		for _, filename := range skillNames {
			results = append(results, checkOneInstalledSkill(l.dir, l.label, filename))
		}
	}
	return results
}

func checkOneInstalledSkill(dir, label, filename string) checkResult {
	skillName := strings.TrimSuffix(filename, ".md")
	r := checkResult{
		group: "Skills",
		name:  skillName + " — " + label,
	}

	if dir == "" {
		r.status = statusSkip
		r.message = "location not available on this platform"
		return r
	}

	installedPath := filepath.Join(dir, filename)

	data, err := os.ReadFile(installedPath)
	if err != nil {
		r.status = statusSkip
		r.message = "not installed"
		return r
	}

	installedHash, ok := skills.ExtractHashFromFrontmatter(data)
	if !ok {
		r.status = statusWarn
		r.message = "outdated (no version metadata)"
		r.detail = "Run: stqry setup claude (or --global)"
		return r
	}

	embeddedData, err := skills.SkillFiles.ReadFile(filename)
	if err != nil {
		r.status = statusFail
		r.message = "embedded skill missing: " + filename
		return r
	}

	if installedHash != skills.HashContent(embeddedData) {
		r.status = statusWarn
		r.message = "outdated (skill content has changed)"
		r.detail = "Run: stqry setup claude (or --global)"
		return r
	}

	r.status = statusPass
	r.message = "up to date"
	return r
}

// errDoctorFailed is a sentinel returned by the doctor command when checks fail.
// It is silenced (not printed) by cobra; the caller's Execute() returns it so
// the process can exit with code 1 without printing a redundant error line.
var errDoctorFailed = fmt.Errorf("doctor checks failed")

func newDoctorCmd() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check config, API connectivity, and CLI version",
		Long:  "doctor runs a series of diagnostic checks and reports pass/fail/skip for each.",
		Example: `  # Run all health checks
  stqry doctor`,
		// Override root PersistentPreRunE so doctor can run without a valid config.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		// SilenceErrors prevents cobra from printing "Error: doctor checks failed".
		// SilenceUsage prevents the usage block being shown on failure.
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if runDoctor(os.Stdout, verbose) {
				return errDoctorFailed
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detail for each check")
	return cmd
}

// runDoctor runs all checks, prints results, and returns true if any check failed.
func runDoctor(w io.Writer, verbose bool) bool {
	httpClient := &http.Client{Timeout: 15 * time.Second}

	// --- Config group ---
	configPath := config.DefaultGlobalConfigPath()
	var results []checkResult
	results = append(results, checkGlobalConfig(configPath))

	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		results = append(results, checkResult{
			group:  "Config",
			name:   "Directory config found",
			status: statusFail,
			detail: fmt.Sprintf("Could not determine working directory: %v", cwdErr),
		})
	} else {
		results = append(results, checkDirectoryConfig(cwd))
	}

	// Load config gracefully — no hard failure if absent.
	globalCfg, _ := config.LoadGlobalConfig(configPath)
	if globalCfg == nil {
		globalCfg = &config.GlobalConfig{Sites: make(map[string]*config.Site)}
	}
	var dirCfg *config.DirectoryConfig
	if cwd != "" {
		dirCfg, _ = config.FindDirectoryConfig(cwd)
	}

	siteResult, resolvedSite := checkSiteResolved(globalCfg, flagSite, dirCfg)
	results = append(results, siteResult)

	// --- API group (only if site resolved) ---
	if siteResult.status == statusPass {
		results = append(results, checkAPIReachable(resolvedSite.APIURL, httpClient))
		results = append(results, checkTokenValid(resolvedSite.APIURL, resolvedSite.Token, httpClient))
		results = append(results, checkRegion(resolvedSite.APIURL))
	} else {
		results = append(results,
			checkResult{group: "API", name: "API reachable", status: statusSkip, message: "No site resolved"},
			checkResult{group: "API", name: "Token valid", status: statusSkip, message: "No site resolved"},
			checkResult{group: "API", name: "Region", status: statusSkip, message: "No site resolved"},
		)
	}

	// --- Version group ---
	results = append(results, checkCLIVersion(buildinfo.Version, defaultGitHubReleasesURL, httpClient))

	// --- Skills group ---
	results = append(results, checkInstalledSkills()...)

	printDoctorResults(w, results, verbose)

	for _, r := range results {
		if r.status == statusFail {
			return true
		}
	}
	return false
}
