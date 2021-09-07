package termio

import (
	"fmt"
	"io"
	"os"
	"strings"
)

//Keys
const (
	CtrlQ  = 'q' & 0x1f
	ZERO   = '0' //in vim zero moves to the beginning of a line
	DOLLAR = '$'
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
	case ZERO:
		config.CursorX = 0
	case DOLLAR:
		config.CursorX = len(config.content[config.CursorY])
	case RIGHT:
		if len(config.content) > 0 &&
			config.CursorX < len(config.content[config.CursorY]) {
			config.CursorX++
		}
	case DOWN:
		if config.CursorY < len(config.content) {
			config.CursorY++
		}
	case TOP:
		if config.CursorY != 0 {
			config.CursorY--
		}
	}
	rowLen := 0
	if config.CursorY < config.screenColumns {
		rowLen = len(config.content[config.CursorY])
	}
	if config.CursorX > rowLen {
		config.CursorX = rowLen
	}
}

//enable scrolling
func editorScroll() {
	//vertical scrolling
	if config.CursorY < config.rowOffset {
		config.rowOffset = config.CursorY
	}
	if config.CursorY >= config.rowOffset+config.screenRows {
		config.rowOffset = config.CursorY - config.screenRows + 1
	}
	//horizontal scrolling
	if config.CursorX < config.columnOffset {
		config.columnOffset = config.CursorX
	}
	if config.CursorX >= config.columnOffset+config.screenColumns {
		config.columnOffset = config.CursorX - config.screenColumns + 1
	}
}

func RefreshScreen() {
	editorScroll()
	//hide cursor
	io.WriteString(os.Stdout, "\x1b[25l")
	io.WriteString(os.Stdout, "\x1b[H")
	drawRows()
	//change cursor position
	io.WriteString(
		os.Stdout,
		fmt.Sprintf(
			"\x1b[%d;%dH",
			config.CursorY-config.rowOffset+1,
			config.CursorX-config.columnOffset+1,
		),
	)
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
	str := fmt.Sprintf("%d\r\n", buffer[0])
	os.WriteFile("/tmp/readactor/test.txt", []byte(str), 0666)
	return int(buffer[0]), nil
}

func drawRows() {
	for y := 0; y < config.screenRows-1; y++ {
		displayRow := y + config.rowOffset
		//display content of a file
		if displayRow >= len(config.content) {
			io.WriteString(
				os.Stdout,
				"~",
			)
		} else {
			//offset columns
			displayLength := len(config.content[displayRow]) - config.columnOffset
			//if cursor is moving left and became negative
			if displayLength < 0 {
				displayLength = 0
			}
			if displayLength > config.screenColumns {
				displayLength = config.screenColumns
			}
			io.WriteString(
				os.Stdout,
				config.content[displayRow][config.columnOffset:displayLength-1],
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
