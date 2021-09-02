package termio

import (
	"fmt"
	"io"
	"os"
	"strings"
)

//Keys
const (
	CtrlQ = 'q' & 0x1f
)

var config Config

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

func drawRows() {
	for y := 0; y < config.screenRows-1; y++ {
		//display content of a file
		if y >= len(config.content) {
			io.WriteString(
				os.Stdout,
				"~",
			)
		} else {
			//if content is bigger than amount of columns then truncate
			length := len(config.content[y])
			if length > config.screenColumns {
				length = config.screenColumns
			}
			io.WriteString(
				os.Stdout,
				config.content[y][0:length],
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

func OpenFile(name string) error {
	rawContent, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	parts := strings.FieldsFunc(
		string(rawContent),
		func(r rune) bool { return r == '\n' },
	)
	config.content = parts
	return nil

}
