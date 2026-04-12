package cli

import (
	"fmt"
	"os"
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
		r.message = "No stqry.yaml found in current directory or parents"
		r.detail = fmt.Sprintf("Looked up from: %s", cwd)
	} else {
		r.status = statusPass
		r.message = "stqry.yaml found"
		r.detail = fmt.Sprintf("Loaded from directory above: %s", cwd)
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
		r.message = err.Error()
	} else {
		r.status = statusPass
		r.message = fmt.Sprintf("Site resolved → %s", site.APIURL)
		r.detail = fmt.Sprintf("API URL: %s", site.APIURL)
	}
	r.duration = time.Since(start)
	return r
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
