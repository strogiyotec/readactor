package termio

import "strings"

const READACTOR_VERSION = "0.0.1"
const TAB_SPACE = 8

type EditorConfig struct {
	cx            int
	cy            int
	rx            int
	rowoff        int
	coloff        int
	screenRows    int
	screenCols    int
	numRows       int
	rows          []erow
	origTermios   *Termios
	fileName      string
	statusMessage string
}

func (c *EditorConfig) Content() string {
	var content strings.Builder
	for _, row := range c.rows {
		content.Write(row.chars)
		content.WriteByte('\n')
	}
	return content.String()
}
