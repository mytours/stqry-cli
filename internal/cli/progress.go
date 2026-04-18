package cli

// progressEnabled reports whether upload progress messages should be written
// to stderr. Progress is opt-in via the --progress global flag, modelled on
// dd(1)'s status=progress: scripted callers (the common case) get clean
// stderr by default, and interactive users opt in explicitly when they want
// the visual feedback.
func progressEnabled() bool {
	return flagProgress
}
