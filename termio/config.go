package termio

const READACTOR_VERSION = "0.0.1"
const TAB_SPACE = 8

type EditorConfig struct {
	cx          int
	cy          int
	rx          int
	rowoff      int
	coloff      int
	screenRows  int
	screenCols  int
	numRows     int
	rows        []erow
	origTermios *Termios
}
