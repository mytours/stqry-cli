package cli

import "testing"

func TestCheckStatusSymbols(t *testing.T) {
	tests := []struct {
		status checkStatus
		want   string
	}{
		{statusPass, "✓"},
		{statusFail, "✗"},
		{statusSkip, "-"},
		{statusInfo, "ℹ"},
		{statusWarn, "⚠"},
	}
	for _, tt := range tests {
		if got := doctorSymbol(tt.status); got != tt.want {
			t.Errorf("doctorSymbol(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}
