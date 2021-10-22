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
	//if first row and first column just skip
	if Config.cx == 0 && Config.cy == 0 {
		return
	}
	currentRow := &Config.rows[Config.cy]
	if Config.cx > 0 {
		editorDeleteCharAt(currentRow, Config.cx-1)
		Config.cx--
	} else {
		//trying to delete the beginning of the line
		//merge this line with a line above
		Config.cx = Config.rows[Config.cy-1].size                          // move cursor to the beginning of a line above
		editorRowAppendString(&Config.rows[Config.cy-1], currentRow.chars) //append this row with one above
		editorDelRow(Config.cy)
		Config.cy--
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

//When backspace was pressed in the beginning of a line
//then merge this line with a one above
func editorRowAppendString(row *erow, rowContent []byte) {
	row.chars = append(row.chars, rowContent...)
	row.size = len(row.chars)
	editorUpdateRow(row)
	Config.dirty = true

}

func editorDelRow(at int) {
	if at < 0 || at >= Config.numRows {
		return
	}
	Config.rows = append(Config.rows[:at], Config.rows[at+1:]...)
	Config.numRows--
	Config.dirty = true
}
