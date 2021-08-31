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

func RefreshScreen() {
	//hide cursor
	io.WriteString(os.Stdout, "\x1b[25l")
	io.WriteString(os.Stdout, "\x1b[H")
	drawRows()
	io.WriteString(os.Stdout, "\x1b[H")
	//show cursor
	io.WriteString(os.Stdout, "\x1b[25h")
}

func InitEditor() error {
	return getTerminalSize()
}

func ReadKey() (rune, error) {
	buffer := make([]byte, 1)
	_, err := os.Stdin.Read(buffer)
	if err != nil {
		return -1, nil
	}
	return rune(buffer[0]), nil
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
