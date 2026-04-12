package cli

import "time"

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
