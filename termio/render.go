package termio

import "fmt"

func editorDrawRows(ab *abuf) {
	for y := 0; y < Config.screenRows; y++ {
		filerow := y + Config.rowoff
		if filerow >= Config.numRows {
			ab.abAppend("~")
		} else {
			len := Config.rows[filerow].rsize - Config.coloff
			if len < 0 {
				len = 0
			}
			if len > Config.screenCols {
				len = Config.screenCols
			}
			rindex := Config.coloff + len
			ab.abAppendBytes(Config.rows[filerow].render[Config.coloff:rindex])
		}
		ab.abAppend("\x1b[K")
		ab.abAppend("\r\n")
	}
}

func editorDrawMessageBar(ab *abuf) {
	ab.abAppend("\x1b[K") //clear status bar first
	msgLength := len(Config.statusMessage)
	if msgLength > Config.screenCols {
		msgLength = Config.screenCols
	}
	ab.abAppend(Config.statusMessage[:msgLength])
}

func editorDrawStatusBar(ab *abuf) {
	ab.abAppend("\x1b[7m") //switch to inverted colors
	statusBar := fmt.Sprintf(
		"%.20s - %d lines %s",
		Config.fileName,
		Config.numRows,
		Config.Modified(),
	)
	length := len(statusBar)
	if length > Config.screenCols {
		length = Config.screenCols
	}
	lineNumberBar := fmt.Sprintf(
		"%d/%d",
		Config.cy+1, //because line number is zero indexed
		Config.numRows,
	)
	ab.abAppend(statusBar[:length])
	for length < Config.screenCols {
		//when in the end of the line then print line number
		if Config.screenCols-length == len(lineNumberBar) {
			ab.abAppend(lineNumberBar)
			break
		} else {
			ab.abAppend(" ")
			length++
		}
	}
	ab.abAppend("\x1b[m") //switch back to normal format
	ab.abAppend("\r\n")   // room for status message
}
