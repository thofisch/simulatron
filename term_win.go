// +build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

func getTermSize() termSize {

	out, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		panic(err)
	}

	c, tr := get_term_size(out)

	fmt.Println(c)
	fmt.Println(tr)

	return termSize{cols: int(c.x), rows: int(c.y)}
}

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var proc_get_console_screen_buffer_info = kernel32.NewProc("GetConsoleScreenBufferInfo")

type short int16
type word uint16

type coord struct {
	x short
	y short
}

type small_rect struct {
	left   short
	top    short
	right  short
	bottom short
}

type console_screen_buffer_info struct {
	size                coord
	cursor_position     coord
	attributes          word
	window              small_rect
	maximum_window_size coord
}

var tmp_info console_screen_buffer_info

func get_term_size(out syscall.Handle) (coord, small_rect) {
	err := get_console_screen_buffer_info(out, &tmp_info)
	if err != nil {
		panic(err)
	}
	return tmp_info.size, tmp_info.window
}

func get_console_screen_buffer_info(h syscall.Handle, info *console_screen_buffer_info) (err error) {
	r0, _, e1 := syscall.Syscall(proc_get_console_screen_buffer_info.Addr(), 2, uintptr(h), uintptr(unsafe.Pointer(info)), 0)
	if int(r0) == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
