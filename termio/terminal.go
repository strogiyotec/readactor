package termio

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

//Config
const (
	TABS = 8
)

type editorRow struct {
	size        int    //actual size
	renderSize  int    //size of rendered row with tabs
	row         []byte // actual row content
	renderedRow []byte // row rendered with tabs
}

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
		} else if config.CursorY > 0 {
			config.CursorY--
			config.CursorX = config.rows[config.CursorY].renderSize
		}
	case ZERO:
		config.CursorX = 0
	case DOLLAR:
		config.CursorX = config.rows[config.CursorY].renderSize
	case RIGHT:
		if config.CursorY < config.screenRows {
			if config.CursorX < config.rows[config.CursorY].renderSize {
				config.CursorX++
			} else if config.CursorX == config.rows[config.CursorY].renderSize {
				config.CursorY++
				config.CursorX = 0
			}
		}
	case DOWN:
		if config.CursorY < config.numberRows {
			config.CursorY++
		}
	case TOP:
		if config.CursorY != 0 {
			config.CursorY--
		}
	}
	rowLen := 0
	if config.CursorY < config.numberRows {
		rowLen = config.rows[config.CursorY].renderSize
	}
	if config.CursorX > rowLen {
		config.CursorX = rowLen
	}
}

func toRenderIndex(row *editorRow, cursorX int) int {
	rx := 0
	for j := 0; j < row.size && j < cursorX; j++ {
		if row.row[j] == '\t' {
			rx += ((TABS - 1) - (rx % TABS))
		}
		rx++
	}
	return rx
}

//enable scrolling
func editorScroll() {
	config.renderX = 0

	if config.CursorY < config.numberRows {
		config.renderX = toRenderIndex(&(config.rows[config.CursorY]), config.CursorX)
	}
	//vertical scrolling
	if config.CursorY < config.rowOffset {
		config.rowOffset = config.CursorY
	}
	if config.CursorY >= config.rowOffset+config.screenRows {
		config.rowOffset = config.CursorY - config.screenRows + 1
	}
	//horizontal scrolling
	if config.renderX < config.columnOffset {
		config.columnOffset = config.renderX
	}
	if config.renderX >= config.columnOffset+config.screenColumns {
		config.columnOffset = config.renderX - config.screenColumns + 1
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
			config.renderX-config.columnOffset+1,
		),
	)
	//show cursor
	io.WriteString(os.Stdout, "\x1b[?25h")
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
	for y := 0; y < config.screenRows; y++ {
		fileRow := y + config.rowOffset
		//for empty lines display ~
		if fileRow >= config.numberRows {
			io.WriteString(
				os.Stdout,
				"~",
			)
		} else {
			len := config.rows[fileRow].renderSize - config.columnOffset
			if len < 0 {
				len = 0
			}
			if len > config.screenColumns {
				len = config.screenColumns
			}

			io.WriteString(
				os.Stdout,
				string(
					config.rows[fileRow].
						renderedRow[config.columnOffset:config.columnOffset+len],
				),
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

//return given string with spaces as a tab
func tabAwareString(line string) string {
	var builder strings.Builder
	for _, c := range line {
		if c == '\t' {
			builder.WriteString(strings.Repeat(" ", TABS))
		} else {
			builder.WriteRune(c)
		}
	}
	return builder.String()
}

func OpenFile(name string) error {
	fd, err := os.Open(name)
	if err != nil {
		return err
	}
	defer fd.Close()
	fp := bufio.NewReader(fd)

	for line, err := fp.ReadBytes('\n'); err == nil; line, err = fp.ReadBytes('\n') {
		// Trim trailing newlines and carriage returns
		for c := line[len(line)-1]; len(line) > 0 && (c == '\n' || c == '\r'); {
			line = line[:len(line)-1]
			if len(line) > 0 {
				c = line[len(line)-1]
			}
		}
		appendRow(line)
	}

	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func appendRow(line []byte) {
	var row editorRow
	row.row = line
	row.size = len(line)
	config.rows = append(config.rows, row)
	tabAwareRow(&config.rows[config.numberRows])
	config.numberRows++
}

//render the editor row with proper amount of tabs in the end
//and in the beginning
func tabAwareRow(eRow *editorRow) {
	//count tabs
	tabs := 0
	for _, c := range eRow.row {
		if c == '\t' {
			tabs++
		}
	}
	eRow.renderedRow = make([]byte, eRow.size+tabs*(TABS-1))
	//now replace all tabs with spaces
	idx := 0
	for _, c := range eRow.row {
		if c == '\t' {
			eRow.renderedRow[idx] = ' '
			idx++
			for idx%TABS != 0 {
				eRow.renderedRow[idx] = ' '
				idx++
			}
		} else {
			eRow.renderedRow[idx] = c
			idx++
		}
	}
	eRow.renderSize = idx
}
