//go:build darwin

package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

// disableInputProcessing puts the terminal into a mode where keypresses
// don't affect the display. Disables echo, canonical mode, and extended
// input processing (VDISCARD, VREPRINT) while keeping signal generation
// (Ctrl+C) enabled. Returns a function to restore the original state.
func disableInputProcessing() func() {
	fd := int(os.Stdin.Fd())
	termios, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return func() {}
	}
	old := *termios
	// ECHO: don't echo keypresses
	// ICANON: disable canonical mode (line buffering, control char processing)
	// IEXTEN: disable extended input processing (VDISCARD=Ctrl-O, VREPRINT=Ctrl-R)
	// Keep ISIG enabled so Ctrl+C still generates SIGINT
	termios.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN
	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, termios); err != nil {
		return func() {}
	}
	return func() {
		unix.IoctlSetTermios(fd, unix.TIOCSETA, &old)
	}
}

// handleSuspendResume handles Ctrl-Z (SIGTSTP) by restoring the terminal
// before suspending and re-applying settings on resume (SIGCONT).
// Holds displayMu across the suspend/resume cycle so the main loop
// cannot print while the terminal is being restored.
func handleSuspendResume(cleanup, setup, redraw func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTSTP, syscall.SIGCONT)
	go func() {
		for sig := range ch {
			switch sig {
			case syscall.SIGTSTP:
				// Lock display so no output races with suspend
				displayMu.Lock()
				cleanup()
				// Re-send SIGTSTP with default handler to actually suspend.
				// The lock stays held; the whole process is stopped by the OS.
				signal.Reset(syscall.SIGTSTP)
				syscall.Kill(syscall.Getpid(), syscall.SIGTSTP)
			case syscall.SIGCONT:
				// Process resumed — re-apply terminal settings and redraw
				setup()
				redraw()
				signal.Notify(ch, syscall.SIGTSTP)
				displayMu.Unlock()
			}
		}
	}()
}
