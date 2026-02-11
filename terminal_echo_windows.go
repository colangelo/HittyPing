//go:build windows

package main

// disableEcho is a no-op on Windows; the console does not echo by default
// in the same way Unix terminals do.
func disableEcho() func() {
	return func() {}
}
