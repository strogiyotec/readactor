package main

import (
	"os"

	"github.com/strogiyotec/readactor/termio"
)

func main() {
	termio.EnableRawMode(&termio.Config)
	defer termio.DisableRawMode(&termio.Config)
	termio.InitEditor()
	if len(os.Args) > 1 {
		termio.EditorOpen(os.Args[1])
	}

	for {
		termio.EditorRefreshScreen()
		termio.EditorProcessKeypress()
	}
}
