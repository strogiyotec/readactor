package termio

func editorRowInsertChar(row *erow, at int, c byte) {
	if at < 0 || at > row.size {
		at = row.size
	}
	row.chars = append(row.chars, c)
	row.size++
	editorUpdateRow(row)
}

func editorInsertChar(c byte) {
	//if cursor is in the end of file then we need to append one empty row
	if Config.cy == Config.numRows {
		editorAppendRow([]byte{})
	}
	editorRowInsertChar(&Config.rows[Config.cy], Config.cx, c)
	Config.cx++
}
