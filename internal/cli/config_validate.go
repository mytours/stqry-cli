package cli

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mytours/stqry-cli/internal/config"
	"github.com/spf13/cobra"
)

// errValidateFailed is a sentinel returned when validation checks fail.
// SilenceErrors on the command prevents cobra from printing it.
var errValidateFailed = fmt.Errorf("config validation failed")

func newConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Check config file syntax and token validity",
		Example: `  # Validate the currently resolved site
  stqry config validate

  # Validate a specific named site
  stqry config validate --site mysite`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			httpClient := &http.Client{Timeout: 15 * time.Second}
			if runConfigValidate(flagSite, httpClient) {
				return errValidateFailed
			}
			return nil
		},
	}
	return cmd
}

// runConfigValidate executes the validation checks and returns true if any check failed.
// Checks are intentionally chained: each step depends on the previous one succeeding
// (e.g. you cannot check whether a named site resolves until the global config parses,
// and you cannot check the token until a site is resolved). On failure we print results
// accumulated so far and return early rather than proceeding with an invalid config.
func runConfigValidate(targetSite string, httpClient *http.Client) bool {
	var results []checkResult

	// Always check: global config parseable.
	configPath := config.DefaultGlobalConfigPath()
	globalCfg, globalErr := config.LoadGlobalConfig(configPath)
	if globalErr != nil {
		results = append(results, checkResult{
			group:   "Config",
			name:    "Global config parsed",
			status:  statusFail,
			message: globalErr.Error(),
		})
		printDoctorResults(os.Stdout, results, false)
		return true
	}
	results = append(results, checkResult{
		group:   "Config",
		name:    "Global config parsed",
		status:  statusPass,
		message: fmt.Sprintf("%d sites", len(globalCfg.Sites)),
	})

	var site *config.Site

	if targetSite != "" {
		// --site flag: skip directory config checks, target named site directly.
		s, ok := globalCfg.Sites[targetSite]
		if !ok {
			results = append(results, checkResult{
				group:   "Config",
				name:    fmt.Sprintf("Site %q found in global config", targetSite),
				status:  statusFail,
				message: fmt.Sprintf("site %q not found", targetSite),
			})
			printDoctorResults(os.Stdout, results, false)
			return true
		}
		results = append(results, checkResult{
			group:  "Config",
			name:   fmt.Sprintf("Site %q found in global config", targetSite),
			status: statusPass,
		})
		site = s
	} else {
		// No --site flag: validate full resolution chain.
		cwd, _ := os.Getwd()
		var dirCfg *config.DirectoryConfig
		var dirCfgPath string
		if cwd != "" {
			var dirCfgErr error
			dirCfg, dirCfgPath, dirCfgErr = config.FindDirectoryConfigWithPath(cwd)
			if dirCfgErr != nil {
				results = append(results, checkResult{
					group:   "Config",
					name:    "Directory config parsed",
					status:  statusFail,
					message: dirCfgErr.Error(),
				})
				printDoctorResults(os.Stdout, results, false)
				return true
			}
		}

		if dirCfgPath != "" {
			results = append(results, checkResult{
				group:   "Config",
				name:    "Directory config parsed",
				status:  statusPass,
				message: dirConfigDescription(dirCfg),
			})
			// If the directory config references a named site, verify it exists.
			if dirCfg != nil && dirCfg.Site != "" {
				if _, ok := globalCfg.Sites[dirCfg.Site]; ok {
					results = append(results, checkResult{
						group:  "Config",
						name:   fmt.Sprintf("Named site resolves (%s)", dirCfg.Site),
						status: statusPass,
					})
				} else {
					results = append(results, checkResult{
						group:   "Config",
						name:    fmt.Sprintf("Named site resolves (%s)", dirCfg.Site),
						status:  statusFail,
						message: fmt.Sprintf("site %q not found in global config", dirCfg.Site),
					})
					printDoctorResults(os.Stdout, results, false)
					return true
				}
			}
		}

		// Resolve the effective site.
		resolved, _, resolveErr := config.ResolveSiteWithSource(globalCfg, "", dirCfg)
		if resolveErr != nil {
			results = append(results, checkResult{
				group:   "Config",
				name:    "Site resolved",
				status:  statusFail,
				message: resolveErr.Error(),
			})
			printDoctorResults(os.Stdout, results, false)
			return true
		}
		results = append(results, checkResult{
			group:   "Config",
			name:    "Site resolved",
			status:  statusPass,
			message: resolved.APIURL,
		})
		site = resolved
	}

	// Live token check.
	results = append(results, checkTokenValid(site.APIURL, site.Token, httpClient))

	printDoctorResults(os.Stdout, results, false)

	for _, r := range results {
		if r.status == statusFail {
			return true
		}
	}
	return false
}
