package termio

type Config struct {
	cx          int
	cy          int
	rx          int
	rowoff      int
	coloff      int
	screenRows  int
	screenCols  int
	numRows     int
	rows        []editorRow
	origTermios *Termios
}

func (c Config) Version() string {
	return "0.0.1"
}
