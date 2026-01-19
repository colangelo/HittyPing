//go:build unix

package main

import (
	"syscall"
	"unsafe"
)

// getTermWidth returns terminal width, defaulting to 80
func getTermWidth() int {
	type winsize struct {
		Row, Col, Xpixel, Ypixel uint16
	}
	var ws winsize
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)))
	if ws.Col == 0 {
		return 80
	}
	return int(ws.Col)
}
