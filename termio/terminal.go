package termio

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

//local flags
const (
	readByByte    = syscall.ICANON
	disableCtrlCZ = syscall.ISIG
	disableCtrlV  = syscall.IEXTEN
)

//input flags
const (
	disableCtrlSQ = syscall.IXON
	disableCtrlM  = syscall.ICRNL
)

//output flags
const (
	disableOutputProcessing = syscall.OPOST
)

//Keys
const (
	CtrlQ = 'q' & 0x1f
)

var config Config

//This is the copy of struct from C termios lib
type Termios struct {
	Iflag  uint32   //Input mode flags
	Oflag  uint32   //Output mode flags
	Cflag  uint32   //Control mode flags
	Lflag  uint32   // Local mode flags
	Cc     [20]byte //Control Characters
	Ispeed uint32   //Input speed
	Ospeed uint32   //Output speed
}
type WinSize struct {
	Row uint16
	Col uint16
}

//Move cursor according to vim keys
func MoveCursor(key int) {
	switch key {
	case LEFT:
		if config.CursorX != 0 {
			config.CursorX--
		}
	case RIGHT:
		if config.CursorX != config.screenColumns-1 {
			config.CursorX++
		}
	case DOWN:
		if config.CursorY != config.screenRows-1 {
			config.CursorY++
		}
	case TOP:
		if config.CursorY != 0 {
			config.CursorY--
		}
	}
}

func RefreshScreen() {
	//hide cursor
	io.WriteString(os.Stdout, "\x1b[25l")
	io.WriteString(os.Stdout, "\x1b[H")
	drawRows()
	//change cursor position
	io.WriteString(os.Stdout, fmt.Sprintf("\x1b[%d;%dH", config.CursorY+1, config.CursorX+1))
	//show cursor
	io.WriteString(os.Stdout, "\x1b[25h")
}

func InitEditor() error {
	//cursor in the top left corner
	config.CursorX = 0
	config.CursorY = 0
	return getTerminalSize()
}

func ReadKey() (int, error) {
	buffer := make([]byte, 1)
	_, err := os.Stdin.Read(buffer)
	if err != nil {
		return 0, nil
	}
	//handle special key presses
	if buffer[0] == '\x1b' {
		var seq [2]byte
		cc, _ := os.Stdin.Read(seq[:])
		if cc != 2 {
			return '\x1b', nil
		}
		if seq[0] == '[' {
			//If it's page down/up keypress
			if seq[1] >= '0' && seq[1] <= '9' {
				if cc, err = os.Stdin.Read(buffer[:]); cc != 1 {
					return '\x1b', nil
				}
				if buffer[0] == '~' {
					switch seq[1] {
					case '5':
						return PAGE_UP, nil
					case '6':
						return PAGE_DOWN, nil
					}
				}
			} else {
				//if arrow keys were pressed replaced them with vim specific movement bindings
				//when arrow key is pressed , terminal sands '\x1b','[' followed by A,B,C,D
				switch seq[1] {
				case 'A':
					return TOP, nil
				case 'B':
					return DOWN, nil
				case 'C':
					return RIGHT, nil
				case 'D':
					return LEFT, nil
				}
			}
		}
	}
	return int(buffer[0]), nil
}
func DisableRawMode() error {
	return TcSetAttr(os.Stdin.Fd(), config.originTerm)
}

func EnableRawMode() error {
	//store original terminal config before modifing it
	term, err := TcGetAttr(os.Stdin.Fd())
	if err != nil {
		return err
	}
	config.originTerm = term
	//modify current terminal
	raw := *config.originTerm
	raw.Lflag &^= syscall.ECHO | readByByte | disableCtrlCZ | disableCtrlV
	raw.Iflag &^= disableCtrlSQ | disableCtrlM |
		/*other flags for old terminals*/
		syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
	raw.Oflag &^= disableOutputProcessing
	raw.Cflag &^= syscall.CS8
	return TcSetAttr(os.Stdin.Fd(), &raw)

}

func TcSetAttr(fd uintptr, termios *Termios) error {
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TCSETS+1),
		uintptr(unsafe.Pointer(termios)),
	)
	if err != 0 {
		return err
	}
	return nil
}

//Copy of C tcgetattr function
func TcGetAttr(fd uintptr) (*Termios, error) {
	terminal := &Termios{}
	_, _, errCode := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		syscall.TCGETS,
		uintptr(unsafe.Pointer(terminal)),
	)
	if errCode != 0 {
		return nil, errors.New("Error getting terminal attributes")
	}
	return terminal, nil
}

func drawRows() {
	for y := 0; y < config.screenRows-1; y++ {
		if y == config.screenRows/3 {
			io.WriteString(
				os.Stdout,
				fmt.Sprintf("Kilo editor -- version %s", config.Version()),
			)
		} else {
			io.WriteString(
				os.Stdout,
				"~",
			)
		}
		io.WriteString(
			os.Stdout,
			"\x1b[K",
		)
		if y < config.screenRows-1 {
			io.WriteString(os.Stdout, "\r\n")
		}
	}
}
func getTerminalSize() error {
	var w WinSize
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdout.Fd(),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&w)),
	)
	//move cursor bottom right
	if err != 0 {
		return errors.New(
			fmt.Sprintf(
				"Error getting terminal size, IOCTL returned %d",
				err,
			),
		)
	}
	config.screenRows = int(w.Row)
	config.screenColumns = int(w.Col)
	return nil
}
