package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/doctor"
)

func registerDoctorTools(s *server.MCPServer, cliVersion string) {
	s.AddTool(
		mcpgo.NewTool("run_doctor",
			mcpgo.WithDescription(
				"Run STQRY CLI diagnostics: checks config validity, API connectivity, and CLI version. "+
					"Uses the site currently configured via stqry.yaml or global config — does not accept a site parameter. "+
					"Call this first when you suspect connectivity or authentication issues. "+
					"Check the top-level any_failed field for a quick pass/fail summary.",
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			result := doctor.RunChecks(cliVersion)

			type jsonCheck struct {
				Group      string `json:"group"`
				Name       string `json:"name"`
				Status     string `json:"status"`
				Message    string `json:"message,omitempty"`
				Detail     string `json:"detail,omitempty"`
				DurationMS int64  `json:"duration_ms"`
			}
			type jsonResponse struct {
				AnyFailed bool        `json:"any_failed"`
				Checks    []jsonCheck `json:"checks"`
			}

			var checks []jsonCheck
			for _, r := range result.Checks {
				checks = append(checks, jsonCheck{
					Group:      r.Group,
					Name:       r.Name,
					Status:     string(r.Status),
					Message:    r.Message,
					Detail:     r.Detail,
					DurationMS: r.Duration.Milliseconds(),
				})
			}

			out, err := jsonResult(jsonResponse{AnyFailed: result.AnyFailed, Checks: checks})
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("serialising results: %v", err)), nil
			}
			return out, nil
		},
	)
}
