//go:build darwin

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

// disableEcho turns off stdin echo so keypresses don't corrupt the display.
// Returns a function to restore the original terminal state.
func disableEcho() func() {
	fd := int(os.Stdin.Fd())
	termios, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return func() {}
	}
	old := *termios
	termios.Lflag &^= unix.ECHO
	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, termios); err != nil {
		return func() {}
	}
	return func() {
		unix.IoctlSetTermios(fd, unix.TIOCSETA, &old)
	}
}
