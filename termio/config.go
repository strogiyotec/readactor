package termio

type Config struct {
	originTerm    *Termios //keep original terminal settings and restore them on exit
	screenRows    int      //rows in a screen
	screenColumns int      //columns in a screen
	CursorX       int      //current x position
	CursorY       int      //current y position
	rows          []editorRow
	rowOffset     int
	columnOffset  int
	renderX       int //of there are tabs in the file then renderX is bigger than currentX
	numberRows    int
}

func (c Config) Version() string {
	return "0.0.1"
}
