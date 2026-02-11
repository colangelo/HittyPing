//go:build linux

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

// disableInputProcessing puts the terminal into a mode where keypresses
// don't affect the display. Disables echo, canonical mode, and extended
// input processing (VDISCARD, VREPRINT) while keeping signal generation
// (Ctrl+C) enabled. Returns a function to restore the original state.
func disableInputProcessing() func() {
	fd := int(os.Stdin.Fd())
	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return func() {}
	}
	old := *termios
	// ECHO: don't echo keypresses
	// ICANON: disable canonical mode (line buffering, control char processing)
	// IEXTEN: disable extended input processing (VDISCARD=Ctrl-O, VREPRINT=Ctrl-R)
	// Keep ISIG enabled so Ctrl+C still generates SIGINT
	termios.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN
	if err := unix.IoctlSetTermios(fd, unix.TCSETS, termios); err != nil {
		return func() {}
	}
	return func() {
		unix.IoctlSetTermios(fd, unix.TCSETS, &old)
	}
}
