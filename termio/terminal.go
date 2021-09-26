package termio

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

//Config
const (
	TABS = 8
)

type editorRow struct {
	size   int
	rsize  int
	chars  []byte
	render []byte
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

func EditorMoveCursor(key int) {
	switch key {
	case ARROW_LEFT:
		if config.cx != 0 {
			config.cx--
		} else if config.cy > 0 {
			config.cy--
			config.cx = config.rows[config.cy].rsize
		}
	case ARROW_RIGHT:
		if config.cy < config.numRows {
			if config.cx < config.rows[config.cy].rsize {
				config.cx++
			} else if config.cx == config.rows[config.cy].rsize {
				config.cy++
				config.cx = 0
			}
		}
	case ARROW_UP:
		if config.cy != 0 {
			config.cy--
		}
	case ARROW_DOWN:
		if config.cy < config.numRows {
			config.cy++
		}
	}

	rowlen := 0
	if config.cy < config.numRows {
		rowlen = config.rows[config.cy].rsize
	}
	if config.cx > rowlen {
		config.cx = rowlen
	}
}

func editorRowCxToRx(row *editorRow, cx int) int {
	rx := 0
	for j := 0; j < row.size && j < cx; j++ {
		if row.chars[j] == '\t' {
			rx += ((TABS - 1) - (rx % TABS))
		}
		rx++
	}
	return rx
}

//enable scrolling
func editorScroll() {
	config.rx = 0

	if config.cy < config.numRows {
		config.rx = editorRowCxToRx(&(config.rows[config.cy]), config.cx)
	}

	if config.cy < config.rowoff {
		config.rowoff = config.cy
	}
	if config.cy >= config.rowoff+config.screenRows {
		config.rowoff = config.cy - config.screenRows + 1
	}
	if config.rx < config.coloff {
		config.coloff = config.rx
	}
	if config.rx >= config.coloff+config.screenCols {
		config.coloff = config.rx - config.screenCols + 1
	}
}

type abuf struct {
	buf []byte
}

func (p abuf) String() string {
	return string(p.buf)
}

func (p *abuf) abAppend(s string) {
	p.buf = append(p.buf, []byte(s)...)
}

func (p *abuf) abAppendBytes(b []byte) {
	p.buf = append(p.buf, b...)
}
func RefreshScreen() {
	editorScroll()
	var ab abuf
	ab.abAppend("\x1b[25l")
	ab.abAppend("\x1b[H")
	drawRows(&ab)
	ab.abAppend(fmt.Sprintf("\x1b[%d;%dH", (config.cy-config.rowoff)+1, (config.rx-config.coloff)+1))
	ab.abAppend("\x1b[?25h")
	_, e := io.WriteString(os.Stdout, ab.String())
	if e != nil {
		log.Fatal(e)
	}
}

func InitEditor() error {
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

func drawRows(ab *abuf) {
	for y := 0; y < config.screenRows; y++ {
		filerow := y + config.rowoff
		if filerow >= config.numRows {
			ab.abAppend("~")
		} else {
			len := config.rows[filerow].rsize - config.coloff
			if len < 0 {
				len = 0
			}
			if len > config.screenCols {
				len = config.screenCols
			}
			ab.abAppendBytes(config.rows[filerow].render[config.coloff : config.coloff+len])
		}
		ab.abAppend("\x1b[K")
		if y < config.screenRows-1 {
			ab.abAppend("\r\n")
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
		editorAppendRow(line)
	}

	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func editorAppendRow(s []byte) {
	var r editorRow
	r.chars = s
	r.size = len(s)
	config.rows = append(config.rows, r)
	editorUpdateRow(&config.rows[config.numRows])
	config.numRows++
}

func editorUpdateRow(row *editorRow) {
	tabs := 0
	for _, c := range row.chars {
		if c == '\t' {
			tabs++
		}
	}
	row.render = make([]byte, row.size+tabs*(TABS-1))

	idx := 0
	for _, c := range row.chars {
		if c == '\t' {
			row.render[idx] = ' '
			idx++
			for (idx % TABS) != 0 {
				row.render[idx] = ' '
				idx++
			}
		} else {
			row.render[idx] = c
			idx++
		}
	}
	row.rsize = idx
}

func EditorReadKey() int {
	var buffer [1]byte
	var cc int
	var err error
	for cc, err = os.Stdin.Read(buffer[:]); cc != 1; cc, err = os.Stdin.Read(buffer[:]) {
	}
	if err != nil {
		return -1
	}
	if buffer[0] == '\x1b' {
		var seq [2]byte
		if cc, _ = os.Stdin.Read(seq[:]); cc != 2 {
			return '\x1b'
		}

		if seq[0] == '[' {
			if seq[1] >= '0' && seq[1] <= '9' {
				if cc, err = os.Stdin.Read(buffer[:]); cc != 1 {
					return '\x1b'
				}
				if buffer[0] == '~' {
					switch seq[1] {
					case '1':
						return HOME_KEY
					case '3':
						return DEL_KEY
					case '4':
						return END_KEY
					case '5':
						return PAGE_UP
					case '6':
						return PAGE_DOWN
					case '7':
						return HOME_KEY
					case '8':
						return END_KEY
					}
				}
				// XXX - what happens here?
			} else {
				switch seq[1] {
				case 'A':
					return ARROW_UP
				case 'B':
					return ARROW_DOWN
				case 'C':
					return ARROW_RIGHT
				case 'D':
					return ARROW_LEFT
				case 'H':
					return HOME_KEY
				case 'F':
					return END_KEY
				}
			}
		} else if seq[0] == '0' {
			switch seq[1] {
			case 'H':
				return HOME_KEY
			case 'F':
				return END_KEY
			}
		}

		return '\x1b'
	}
	return int(buffer[0])
}
func EditorProcessKeypress() {
	c := EditorReadKey()
	switch c {
	case ('q' & 0x1f):
		io.WriteString(os.Stdout, "\x1b[2J")
		io.WriteString(os.Stdout, "\x1b[H")
		DisableRawMode()
		os.Exit(0)
	case HOME_KEY:
		config.cx = 0
	case END_KEY:
		config.cx = config.screenCols - 1
	case PAGE_UP, PAGE_DOWN:
		dir := ARROW_DOWN
		if c == PAGE_UP {
			dir = ARROW_UP
		}
		for times := config.screenRows; times > 0; times-- {
			editorMoveCursor(dir)
		}
	case ARROW_UP, ARROW_DOWN, ARROW_LEFT, ARROW_RIGHT:
		editorMoveCursor(c)
	}
}

func editorMoveCursor(key int) {
	switch key {
	case ARROW_LEFT:
		if config.cx != 0 {
			config.cx--
		} else if config.cy > 0 {
			config.cy--
			config.cx = config.rows[config.cy].rsize
		}
	case ARROW_RIGHT:
		if config.cy < config.numRows {
			if config.cx < config.rows[config.cy].rsize {
				config.cx++
			} else if config.cx == config.rows[config.cy].rsize {
				config.cy++
				config.cx = 0
			}
		}
	case ARROW_UP:
		if config.cy != 0 {
			config.cy--
		}
	case ARROW_DOWN:
		if config.cy < config.numRows {
			config.cy++
		}
	}

	rowlen := 0
	if config.cy < config.numRows {
		rowlen = config.rows[config.cy].rsize
	}
	if config.cx > rowlen {
		config.cx = rowlen
	}
}
