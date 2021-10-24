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
		editorInsertRow(Config.numRows, []byte{})
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

func editorInsertRow(at int, content []byte) {
	if at < 0 || at > Config.numRows {
		return
	}
	var r erow
	r.chars = content
	r.size = len(content)
	//means user pressed enter in the first row
	//in this case move everything one line below
	//and create a new empty line on top
	if at == 0 {
		t := make([]erow, 1)
		t[0] = r
		Config.rows = append(t, Config.rows...)
	} else if at == Config.numRows {
		//just append a new row to the end of the file
		Config.rows = append(Config.rows, r)
	} else {
		//user pressed enter somewhere between first and last row
		//create an empty line and move everything below this line
		//one level down
		t := make([]erow, 1)
		t[0] = r
		Config.rows = append(
			Config.rows[:at],
			append(t, Config.rows[at:]...)...,
		)
	}
	editorUpdateRow(&Config.rows[at])
	Config.numRows++
	Config.dirty = true
}

//Callback for Enter key
func editorInsertNewLine() {
	//If we are in the beginning of a file
	//then just create an empty row
	if Config.cx == 0 {
		editorInsertRow(Config.cy, []byte{})
	} else {
		row := &Config.rows[Config.cy]
		//If user pressed Enter in the middle of the line
		//then split row into two groups
		//before cursor and after cursor
		//make a new line from characters after cursor
		editorInsertRow(Config.cy+1, row.chars[Config.cx:])
		//keep only those characters which were before cursor
		//in this row
		Config.rows[Config.cy].chars = Config.rows[Config.cy].chars[:Config.cx]
		Config.rows[Config.cy].size = len(Config.rows[Config.cy].chars)
		editorUpdateRow(&Config.rows[Config.cy])
	}
	Config.dirty = true
	//move cursor to the beginning of next line
	Config.cy++
	Config.cx = 0

}
