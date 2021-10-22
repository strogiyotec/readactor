package termio

func editorRowInsertChar(row *erow, at int, c byte) {
	if at < 0 || at > row.size {
		at = row.size
	}
	row.chars = append(row.chars, c)
	row.size++
	editorUpdateRow(row)
}

func editorDelChar() {
	//cursor passed the end of file
	if Config.cy == Config.numRows {
		return
	}
	if Config.cx > 0 {
		editorDeleteCharAt(&Config.rows[Config.cy], Config.cx-1)
		Config.cx--
	}
}

//delete char from a row
func editorDeleteCharAt(row *erow, at int) {
	if at < 0 || at > row.size {
		return
	}
	if at == row.size {
		row.chars = row.chars[:len(row.chars)-1]
	} else {
		row.chars = append(row.chars[:at], row.chars[at+1:]...)
	}
	row.size--
	Config.dirty = true
	editorUpdateRow(row)
}

func editorInsertChar(c byte) {
	//if cursor is in the end of file then we need to append one empty row
	if Config.cy == Config.numRows {
		editorAppendRow([]byte{})
	}
	editorRowInsertChar(&Config.rows[Config.cy], Config.cx, c)
	Config.cx++
	Config.dirty = true
}
