// +build windows

package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

func main() {
	var hStdout, hNewScreenBuffer HANDLE
	// var srctReadRect SMALL_RECT
	// var srctWriteRect SMALL_RECT
	// CHAR_INFO chiBuffer[160]; // [2][80];
	// var coordBufSize COORD
	// var coordBufCoord COORD
	// var fSuccess BOOL

	hStdout = getStdHandle(syscall.STD_OUTPUT_HANDLE)

	var mode uint32

	GetConsoleMode(hStdout, &mode)
	fmt.Printf("%08x", mode)

	hNewScreenBuffer, _ = CreateConsoleScreenBuffer(
		GENERIC_READ|GENERIC_WRITE,       // read/write access
		FILE_SHARE_READ|FILE_SHARE_WRITE, // shared
		nil,                              // default security attributes
		CONSOLE_TEXTMODE_BUFFER,          // must be TEXTMODE
		0)                                // reserved; must be NULL

	if hStdout == INVALID_HANDLE_VALUE || hNewScreenBuffer == INVALID_HANDLE_VALUE {
		panic("CreateConsoleScreenBuffer failed")
	}

	err := SetConsoleActiveScreenBuffer(hNewScreenBuffer)
	if err != nil {
		panic("SetConsoleActiveScreenBuffer failed")
	}

	err = SetConsoleMode(hNewScreenBuffer, 0x0007)
	if err != nil {
		panic(err)
	}

	GetConsoleMode(hNewScreenBuffer, &mode)
	fmt.Printf("%08x", mode)

	ch := syscall.StringToUTF16("\033[10;10H\033[1;31mthis is a test of the emergency broadcast system\033[0m")
	nw := uint32(len(ch))
	syscall.WriteConsole(syscall.Handle(hNewScreenBuffer), &ch[0], nw, &nw, nil)

	time.Sleep(2 * time.Second)

	SetConsoleActiveScreenBuffer(STD_OUTPUT_HANDLE)

	fmt.Println("\033[1mthis is a test of the emergency broadcast system\033[0m")

	// stdoutHandle := getStdHandle(syscall.STD_OUTPUT_HANDLE)

	// var info, err = GetConsoleScreenBufferInfo(stdoutHandle)

	// if err != nil {
	// 	panic("could not get console screen buffer info")
	// }

	// fmt.Printf("max x: %d max y: %d\n", info.MaximumWindowSize.X, info.MaximumWindowSize.Y)

}

// func getTermSize() termSize {

// 	out, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
// 	if err != nil {
// 		panic(err)
// 	}

// 	c, tr := get_term_size(out)

// 	fmt.Println(c)
// 	fmt.Println(tr)

// 	return termSize{cols: int(c.x), rows: int(c.y)}
// }

const (
	STD_OUTPUT_HANDLE       HANDLE = 0xFFFFFFF5
	INVALID_HANDLE_VALUE    HANDLE = 0xFFFFFFFF
	GENERIC_READ            DWORD  = 0x80000000 //	Requests read access to the console screen buffer, enabling the process to read data from the buffer.
	GENERIC_WRITE           DWORD  = 0x40000000 //	Requests write access to the console screen buffer, enabling the process to write data to the buffer.
	FILE_SHARE_READ         DWORD  = 0x00000001
	FILE_SHARE_WRITE        DWORD  = 0x00000002
	CONSOLE_TEXTMODE_BUFFER DWORD  = 0x00000001

	ENABLE_VIRTUAL_TERMINAL_PROCESSING WORD = 0x0004
)

type (
	LPVOID uintptr
	HANDLE uintptr
	SHORT  int16
	WORD   uint16
	BOOL   int32
	DWORD  uint32

	SMALL_RECT struct {
		Left   SHORT
		Top    SHORT
		Right  SHORT
		Bottom SHORT
	}

	COORD struct {
		X SHORT
		Y SHORT
	}

	CONSOLE_SCREEN_BUFFER_INFO struct {
		Size              COORD
		CursorPosition    COORD
		Attributes        WORD
		Window            SMALL_RECT
		MaximumWindowSize COORD
	}

	SECURITY_ATTRIBUTES struct {
		NLength              uint32
		LpSecurityDescriptor uintptr
		BInheritHandle       BOOL
	}

	PSECURITY_ATTRIBUTES *SECURITY_ATTRIBUTES
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var getConsoleScreenBufferInfoProc = kernel32.NewProc("GetConsoleScreenBufferInfo")
var createConsoleScreenBuffer = kernel32.NewProc("CreateConsoleScreenBuffer")
var setConsoleActiveScreenBuffer = kernel32.NewProc("SetConsoleActiveScreenBuffer")
var getConsoleMode = kernel32.NewProc("GetConsoleMode")
var setConsoleMode = kernel32.NewProc("SetConsoleMode")

func getError(r1, r2 uintptr, lastErr error) error {
	// If the function fails, the return value is zero.
	if r1 == 0 {
		if lastErr != nil {
			return lastErr
		}
		return syscall.EINVAL
	}
	return nil
}

func getStdHandle(stdhandle int) HANDLE {
	handle, err := syscall.GetStdHandle(stdhandle)
	if err != nil {
		panic(fmt.Errorf("could not get standard io handle %d", stdhandle))
	}
	return HANDLE(handle)
}

// GetConsoleScreenBufferInfo retrieves information about the specified console screen buffer.
// http://msdn.microsoft.com/en-us/library/windows/desktop/ms683171(v=vs.85).aspx
func GetConsoleScreenBufferInfo(handle uintptr) (*CONSOLE_SCREEN_BUFFER_INFO, error) {
	var info CONSOLE_SCREEN_BUFFER_INFO
	if err := getError(getConsoleScreenBufferInfoProc.Call(handle, uintptr(unsafe.Pointer(&info)), 0)); err != nil {
		return nil, err
	}
	return &info, nil
}

func CreateConsoleScreenBuffer(dwDesiredAccess DWORD, dwShareMode DWORD, lpSecurityAttributes /*const*/ PSECURITY_ATTRIBUTES, dwFlags DWORD, lpScreenBufferData LPVOID) (HANDLE, error) {

	r1, r2, err := createConsoleScreenBuffer.Call(
		uintptr(dwDesiredAccess),
		uintptr(dwShareMode),
		uintptr(unsafe.Pointer(lpSecurityAttributes)),
		uintptr(dwFlags),
		uintptr(unsafe.Pointer(lpScreenBufferData)))

	if e := getError(r1, r2, err); e != nil {
		return 0, e
	}
	return HANDLE(r1), nil
}

func SetConsoleActiveScreenBuffer(hConsoleOutput HANDLE) error {
	if err := getError(setConsoleActiveScreenBuffer.Call(uintptr(hConsoleOutput))); err != nil {
		return err
	}
	return nil
}

func GetConsoleMode(hConsoleHandle HANDLE, lpMode *uint32) error {
	if err := getError(getConsoleMode.Call(uintptr(hConsoleHandle), uintptr(unsafe.Pointer(lpMode)))); err != nil {
		return err
	}
	return nil
}

func SetConsoleMode(hConsoleHandle HANDLE, dwMode DWORD) error {
	if err := getError(setConsoleMode.Call(uintptr(hConsoleHandle), uintptr(dwMode))); err != nil {
		return err
	}
	return nil
}
