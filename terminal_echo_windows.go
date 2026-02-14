//go:build windows

package main

// disableInputProcessing is a no-op on Windows; the console does not
// process input control characters in the same way Unix terminals do.
func disableInputProcessing() func() {
	return func() {}
}

// handleSuspendResume is a no-op on Windows (no SIGTSTP/SIGCONT).
func handleSuspendResume(cleanup, setup, redraw func()) {}
