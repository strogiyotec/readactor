package termio

import (
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
			renderLine := tabAwareString(config.content[config.CursorY])
			config.CursorX = len(renderLine)
		}
	case ZERO:
		config.CursorX = 0
	case DOLLAR:
		config.CursorX = len(config.content[config.CursorY])
	case RIGHT:
		if config.CursorY < config.screenRows {
			renderLine := tabAwareString(config.content[config.CursorY])
			if config.CursorX < len(renderLine) {
				config.CursorX++
			} else if config.CursorX == len(renderLine) {
				config.CursorY++
				config.CursorX = 0
			}
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
	if config.CursorY < config.screenRows {
		rowLen = len(config.content[config.CursorY])
	}
	if config.CursorX > rowLen {
		config.CursorX = rowLen
	}
}

//enable scrolling
func editorScroll() {
	config.renderX = 0
	if config.CursorY < len(config.content) {
		calculateRenderIndex()
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
		if fileRow >= len(config.content) {
			io.WriteString(
				os.Stdout,
				"~",
			)
		} else {
			line := tabAwareString(config.content[fileRow])
			//offset columns
			displayLength := len(line) - config.columnOffset
			//if cursor is moving left and became negative
			if displayLength < 0 {
				displayLength = 0
			}
			if displayLength > config.screenColumns {
				displayLength = config.screenColumns
			}
			io.WriteString(
				os.Stdout,
				line[config.columnOffset:config.columnOffset+displayLength],
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
	rawContent, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	parts := strings.FieldsFunc(
		string(rawContent),
		func(r rune) bool { return r == '\n' },
	)
	for _, row := range parts {
		line := []byte(row)
		for c := line[len(line)-1]; len(line) > 0 && (c == '\n' || c == '\r'); {
			line = line[:len(line)-1]
			if len(line) > 0 {
				c = line[len(line)-1]
			}
		}
		appendRow(line)
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

//calculate cursor position with tabs in mind
func calculateRenderIndex() {
	row := config.content[config.CursorY]
	renderX := 0
	for i := 0; i < config.CursorX && i < len(config.content[config.CursorY]); i++ {
		if row[i] == '\t' {
			renderX += (TABS - 1) - (renderX % TABS)
		}
		renderX++
	}
	config.renderX = renderX
}
