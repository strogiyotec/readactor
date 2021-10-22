package termio

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var Config EditorConfig

type erow struct {
	size   int
	rsize  int
	chars  []byte
	render []byte
}

type WinSize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func die(err error) {
	DisableRawMode(&Config)
	io.WriteString(os.Stdout, "\x1b[2J")
	io.WriteString(os.Stdout, "\x1b[H")
	log.Fatal(err)
}

func editorReadKey() int {
	var buffer [1]byte
	var cc int
	var err error
	for cc, err = os.Stdin.Read(buffer[:]); cc != 1; cc, err = os.Stdin.Read(buffer[:]) {
	}
	if err != nil {
		die(err)
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

func getCursorPosition(rows *int, cols *int) int {
	io.WriteString(os.Stdout, "\x1b[6n")
	var buffer [1]byte
	var buf []byte
	var cc int
	for cc, _ = os.Stdin.Read(buffer[:]); cc == 1; cc, _ = os.Stdin.Read(buffer[:]) {
		if buffer[0] == 'R' {
			break
		}
		buf = append(buf, buffer[0])
	}
	if string(buf[0:2]) != "\x1b[" {
		log.Printf("Failed to read rows;cols from tty\n")
		return -1
	}
	if n, e := fmt.Sscanf(string(buf[2:]), "%d;%d", rows, cols); n != 2 || e != nil {
		if e != nil {
			log.Printf("getCursorPosition: fmt.Sscanf() failed: %s\n", e)
		}
		if n != 2 {
			log.Printf("getCursorPosition: got %d items, wanted 2\n", n)
		}
		return -1
	}
	return 0
}

/*** row operations ***/

func editorRowCxToRx(row *erow, cx int) int {
	rx := 0
	for j := 0; j < row.size && j < cx; j++ {
		if row.chars[j] == '\t' {
			rx += ((TAB_SPACE - 1) - (rx % TAB_SPACE))
		}
		rx++
	}
	return rx
}

func editorUpdateRow(row *erow) {
	tabs := 0
	for _, c := range row.chars {
		if c == '\t' {
			tabs++
		}
	}
	row.render = make([]byte, row.size+tabs*(TAB_SPACE-1))

	idx := 0
	for _, c := range row.chars {
		if c == '\t' {
			row.render[idx] = ' '
			idx++
			for (idx % TAB_SPACE) != 0 {
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

func editorAppendRow(s []byte) {
	var r erow
	r.chars = s
	r.size = len(s)
	Config.rows = append(Config.rows, r)
	editorUpdateRow(&Config.rows[Config.numRows])
	Config.numRows++
	Config.dirty = true
}

/*** file I/O ***/

func EditorOpen(filename string) {
	Config.fileName = filename
	fd, err := os.Open(filename)
	if err != nil {
		die(err)
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
		die(err)
	}
	//because we can editorAppendRow above it will treat content as dirty
	//need to reset dirty flag here
	Config.dirty = false
}

func editorMoveCursor(key int) {
	switch key {
	case ARROW_LEFT:
		if Config.cx != 0 {
			Config.cx--
		} else if Config.cy > 0 {
			Config.cy--
			Config.cx = Config.rows[Config.cy].size
		}
	case ARROW_RIGHT:
		if Config.cy < Config.numRows {
			if Config.cx < Config.rows[Config.cy].size {
				Config.cx++
			} else if Config.cx == Config.rows[Config.cy].size {
				Config.cy++
				Config.cx = 0
			}
		}
	case ARROW_UP:
		if Config.cy != 0 {
			Config.cy--
		}
	case ARROW_DOWN:
		if Config.cy < Config.numRows {
			Config.cy++
		}
	}

	rowlen := 0
	if Config.cy < Config.numRows {
		rowlen = Config.rows[Config.cy].rsize
	}
	if Config.cx > rowlen {
		Config.cx = rowlen
	}
}

func EditorProcessKeypress() {
	c := editorReadKey()
	switch c {
	case BACKSPACE, DEL_KEY:
		//TODO: handle delete
		if c == DEL_KEY {
			editorMoveCursor(ARROW_RIGHT)
		}
		editorDelChar()
		break
	case ('s' & 0x1f): //save
		editorSave()
	case '\r':
		//TODO: Enter key
	case ('q' & 0x1f): //CTRL + Q -> exit
		if Config.dirty && Config.quitPressedCnt != 0 {
			editorSetStatusMessage(
				fmt.Sprintf(
					"WARNING!!! File has unsaved changes.Press CTRL-Q %d times to quit",
					Config.quitPressedCnt,
				),
			)
			Config.quitPressedCnt--
			return
		}
		io.WriteString(os.Stdout, "\x1b[2J")
		io.WriteString(os.Stdout, "\x1b[H")
		DisableRawMode(&Config)
		os.Exit(0)
		//CTRL + L -> don't do anything
	case ('l' & 0x1f):
		break
	case HOME_KEY:
		Config.cx = 0
	case END_KEY:
		Config.cx = Config.screenCols - 1
	case PAGE_UP, PAGE_DOWN:
		dir := ARROW_DOWN
		if c == PAGE_UP {
			dir = ARROW_UP
		}
		for times := Config.screenRows; times > 0; times-- {
			editorMoveCursor(dir)
		}
	case ARROW_UP, ARROW_DOWN, ARROW_LEFT, ARROW_RIGHT:
		editorMoveCursor(c)
	default:
		editorInsertChar(byte(c))
	}

}

/*** append buffer ***/

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

/*** output ***/

func editorScroll() {
	Config.rx = 0

	if Config.cy < Config.numRows {
		Config.rx = editorRowCxToRx(&(Config.rows[Config.cy]), Config.cx)
	}

	if Config.cy < Config.rowoff {
		Config.rowoff = Config.cy
	}
	if Config.cy >= Config.rowoff+Config.screenRows {
		Config.rowoff = Config.cy - Config.screenRows + 1
	}
	if Config.rx < Config.coloff {
		Config.coloff = Config.rx
	}
	if Config.rx >= Config.coloff+Config.screenCols {
		Config.coloff = Config.rx - Config.screenCols + 1
	}
}

func EditorRefreshScreen() {
	editorScroll()
	var ab abuf
	ab.abAppend("\x1b[25l")
	ab.abAppend("\x1b[H")
	editorDrawRows(&ab)
	editorDrawStatusBar(&ab)
	editorDrawMessageBar(&ab)
	ab.abAppend(
		fmt.Sprintf(
			"\x1b[%d;%dH",
			(Config.cy-Config.rowoff)+1,
			(Config.rx-Config.coloff)+1,
		),
	)
	ab.abAppend("\x1b[?25h")
	_, e := io.WriteString(os.Stdout, ab.String())
	if e != nil {
		log.Fatal(e)
	}
}

//TODO: implement
func editorSetStatusMessage(message string) {
	Config.statusMessage = message
}

//save content of the editor to disk
func editorSave() error {
	editorContent := Config.Content()
	err := ioutil.WriteFile(
		Config.fileName,
		[]byte(editorContent),
		0644,
	)
	if err != nil {
		return err
	}
	editorSetStatusMessage(fmt.Sprintf(
		"%d bytes written to disk",
		len(editorContent),
	))
	//we saved the file, dirty flag is reset
	Config.dirty = false
	return nil
}

func InitEditor() {
	if getWindowSize(&Config.screenRows, &Config.screenCols) == -1 {
		die(fmt.Errorf("couldn't get screen size"))
	}
	//save last two rows for line number and status message
	Config.screenRows -= 2
	Config.statusMessage = "readactor. Press Ctrl-Q to quit"
	Config.dirty = false
	Config.quitPressedCnt = DIRTY_QUIT
}
