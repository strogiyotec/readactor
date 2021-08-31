package termio

type Config struct {
	originTerm    *Termios //keep original terminal settings and restore them on exit
	screenRows    int
	screenColumns int
	CursorX       int
	CursorY       int
}

func (c Config) Version() string {
	return "0.0.1"
}
