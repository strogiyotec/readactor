package termio

import "strings"

const READACTOR_VERSION = "0.0.1"
const TAB_SPACE = 8

//quit without saving changes
//user has to press quit two times to confirm
const DIRTY_QUIT = 2

type EditorConfig struct {
	cx             int
	cy             int
	rx             int
	rowoff         int
	coloff         int
	screenRows     int
	screenCols     int
	numRows        int
	rows           []erow
	origTermios    *Termios
	fileName       string
	statusMessage  string
	dirty          bool // keep track if content was changed
	quitPressedCnt int  //count how many times quit was pressed
}

//the message to be displayed depending on wether
//content was modified
func (c EditorConfig) Modified() string {
	if c.dirty {
		return "(modified)"
	} else {
		return ""
	}
}

func (c *EditorConfig) Content() string {
	var content strings.Builder
	for _, row := range c.rows {
		content.Write(row.chars)
		content.WriteByte('\n')
	}
	return content.String()
}
