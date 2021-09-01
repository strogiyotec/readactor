package termio

type Config struct {
	originTerm    *Termios //keep original terminal settings and restore them on exit
	screenRows    int      //rows in a screen
	screenColumns int      //columns in a screen
	CursorX       int
	CursorY       int
	numRows       int
	contentRow    EditorRow
}

type EditorRow struct {
	numRows int    //number of rows in a text file
	content []byte //content of a file
}

func (c Config) Version() string {
	return "0.0.1"
}
