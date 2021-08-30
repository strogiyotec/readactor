package termio

type Config struct {
	originTerm    *Termios //keep original terminal settings and restore them on exit
	screenRows    int
	screenColumns int
}
