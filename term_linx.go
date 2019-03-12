// +build !windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

func getTermSize() termSize {
	type winsize struct {
		rows    uint16
		cols    uint16
		xpixels uint16
		ypixels uint16
	}

	var sz winsize
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
	return termSize{cols: int(sz.cols), rows: int(sz.rows)}
}
