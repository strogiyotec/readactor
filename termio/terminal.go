package termio

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

type Config struct {
	originTerm *Termios //keep original terminal settings and restore them on exit
}

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
