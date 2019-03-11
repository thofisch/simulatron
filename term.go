package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
)

var term Term

func main() {
	fmt.Println("vim-go")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	newTerm, _ := NewTerm()
	go func() {
		<-c
		newTerm.Close()
		os.Exit(1)
	}()
	defer newTerm.Close()

	fmt.Println("Yo!")

	time.Sleep(5 * time.Second)

}

type Term interface {
	Close()
}

type Terminal struct {
}

func NewTerm() (Term, error) {
	var err error

	out, err = os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}
	in, err = syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	//err = setup_term()
	//if err != nil {
	//	return fmt.Errorf("termbox: error while reading terminfo data: %v", err)
	//}

	//signal.Notify(sigwinch, syscall.SIGWINCH)
	//signal.Notify(sigio, syscall.SIGIO)

	//_, err = fcntl(in, syscall.F_SETFL, syscall.O_ASYNC|syscall.O_NONBLOCK)
	//if err != nil {
	//	return err
	//}
	//_, err = fcntl(in, syscall.F_SETOWN, syscall.Getpid())
	//if runtime.GOOS != "darwin" && err != nil {
	//	return err
	//}

	//var orig_tios syscall_Termios
	//
	//err = tcgetattr(out.Fd(), &orig_tios)
	//if err != nil {
	//	return nil, err
	//}
	//
	//tios := orig_tios
	//tios.Iflag &^= syscall_IGNBRK | syscall_BRKINT | syscall_PARMRK | syscall_ISTRIP | syscall_INLCR | syscall_IGNCR | syscall_ICRNL | syscall_IXON
	//tios.Lflag &^= syscall_ECHO | syscall_ECHONL | syscall_ICANON | syscall_ISIG | syscall_IEXTEN
	//tios.Cflag &^= syscall_CSIZE | syscall_PARENB
	//tios.Cflag |= syscall_CS8
	//tios.Cc[syscall_VMIN] = 1
	//tios.Cc[syscall_VTIME] = 0
	//
	//err = tcsetattr(out.Fd(), &tios)
	//if err != nil {
	//	return nil, err
	//}

	out.WriteString(t_enter_ca)
	out.WriteString(t_enter_keypad)
	out.WriteString(t_hide_cursor)
	out.WriteString(t_clear_screen)

	//var back_buffer cellbuf
	//var front_buffer cellbuf

	termw, termh = get_term_size(out.Fd())
	//back_buffer.init(termw, termh)
	//front_buffer.init(termw, termh)
	//back_buffer.clear()
	//front_buffer.clear()

	printClose:= func() {
		fmt.Printf("\033[1;%dH[X]", termw-(3-1))
	}
	printClose()
	fmt.Printf("\033[1;1H")

	signal.Notify(sigwinch, syscall.SIGWINCH)

	go func() {
		for {
			select {
			case <-sigwinch:
				termw, termh = get_term_size(out.Fd())
				printClose()
			case <-quit:
				return
			}
		}
	}()

	return &Terminal{}, nil
}

func (term Terminal) Close() {
	quit <- 1
	out.WriteString(t_show_cursor)
	out.WriteString(t_sgr0)
	out.WriteString(t_clear_screen)
	out.WriteString(t_exit_ca)
	out.WriteString(t_exit_keypad)
	//tcsetattr(out.Fd(), &orig_tios)

	out.Close()

	syscall.Close(in)
	// reset the state, so that on next Init() it will work again
	//termw = 0
	//termh = 0
	//input_mode = InputEsc
	//out = nil
	//in = 0
	//lastfg = attr_invalid
	//lastbg = attr_invalid
	//lastx = coord_invalid
	//lasty = coord_invalid
	//cursor_x = cursor_hidden
	//cursor_y = cursor_hidden
	//foreground = ColorDefault
	//background = ColorDefault
	//IsInit = false
}

var (
	// termbox inner state
	//orig_tios      syscall_Termios
	//back_buffer    cellbuf
	//front_buffer   cellbuf
	termw int
	termh int
	//input_mode     = InputEsc
	//output_mode    = OutputNormal
	out *os.File
	in  int
	//lastfg         = attr_invalid
	//lastbg         = attr_invalid
	//lastx          = coord_invalid
	//lasty          = coord_invalid
	//cursor_x       = cursor_hidden
	//cursor_y       = cursor_hidden
	foreground = ColorDefault
	background = ColorDefault
	//inbuf          = make([]byte, 0, 64)
	//outbuf         bytes.Buffer
	sigwinch       = make(chan os.Signal, 1)
	//sigio          = make(chan os.Signal, 1)
	quit           = make(chan int)
	//input_comm     = make(chan input_event)
	//interrupt_comm = make(chan struct{})
	//intbuf         = make([]byte, 0, 16)

	// grayscale indexes
	//grayscale = []Attribute{
	//	0, 17, 233, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244,
	//	245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 232,
	//}
)

const (
	ColorDefault Attribute = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

// Cell attributes, it is possible to use multiple attributes by combining them
// using bitwise OR ('|'). Although, colors cannot be combined. But you can
// combine attributes and a single color.
//
// It's worth mentioning that some platforms don't support certain attributes.
// For example windows console doesn't support AttrUnderline. And on some
// terminals applying AttrBold to background may result in blinking text. Use
// them with caution and test your code on various terminals.
const (
	AttrBold Attribute = 1 << (iota + 9)
	AttrUnderline
	AttrReverse
)

type winsize struct {
	rows    uint16
	cols    uint16
	xpixels uint16
	ypixels uint16
}

func get_term_size(fd uintptr) (int, int) {
	var sz winsize
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
	return int(sz.cols), int(sz.rows)
}

type syscall_Termios struct {
	Iflag     uint64
	Oflag     uint64
	Cflag     uint64
	Lflag     uint64
	Cc        [20]uint8
	Pad_cgo_0 [4]byte
	Ispeed    uint64
	Ospeed    uint64
}

const (
	t_enter_ca     = "\x1b[?1049h"        // enter Alternate Screen
	t_exit_ca      = "\x1b[?1049l"        // exit Alternate Screen
	t_show_cursor  = "\x1b[?12l\x1b[?25h" // show cursor
	t_hide_cursor  = "\x1b[?25l"          // hide cursor
	t_clear_screen = "\x1b[h\x1b[2j"      // clear screen
	t_sgr0         = "\x1b(b\x1b[m"       // sgr0
	t_underline    = "\x1b[4m"            // underline
	t_bold         = "\x1b[1m"            // bold
	t_blink        = "\x1b[5m"            // blink
	t_reverse      = "\x1b[7m"            // reverse
	t_enter_keypad = "\x1b[?1h\x1b"       // enter_keypad
	t_exit_keypad  = "\x1b[?1l\x1b"       // exit_keypad
)

type cellbuf struct {
	width  int
	height int
	cells  []Cell
}

func (this *cellbuf) init(width, height int) {
	this.width = width
	this.height = height
	this.cells = make([]Cell, width*height)
}

func (this *cellbuf) resize(width, height int) {
	if this.width == width && this.height == height {
		return
	}

	oldw := this.width
	oldh := this.height
	oldcells := this.cells

	this.init(width, height)
	this.clear()

	minw, minh := oldw, oldh

	if width < minw {
		minw = width
	}
	if height < minh {
		minh = height
	}

	for i := 0; i < minh; i++ {
		srco, dsto := i*oldw, i*width
		src := oldcells[srco : srco+minw]
		dst := this.cells[dsto : dsto+minw]
		copy(dst, src)
	}
}

func (this *cellbuf) clear() {
	for i := range this.cells {
		c := &this.cells[i]
		c.Ch = ' '
		c.Fg = foreground
		c.Bg = background
	}
}

const cursor_hidden = -1

func is_cursor_hidden(x, y int) bool {
	return x == cursor_hidden || y == cursor_hidden
}

type Attribute uint16

type Cell struct {
	Ch rune
	Fg Attribute
	Bg Attribute
}
