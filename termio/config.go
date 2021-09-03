package termio

type Config struct {
	originTerm    *Termios //keep original terminal settings and restore them on exit
	screenRows    int      //rows in a screen
	screenColumns int      //columns in a screen
	CursorX       int
	CursorY       int
	content       []string
	rowOffset     int
	columnOffset  int
}

func (c Config) Version() string {
	return "0.0.1"
}
