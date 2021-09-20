package main

import (
	"errors"
	"fmt"
	"os"

	. "github.com/strogiyotec/readactor/termio"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: readactor FILE_NAME")
		return
	}
	err := EnableRawMode()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer DisableRawMode()
	err = InitEditor()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = OpenFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	for true {
		RefreshScreen()
		err := processKeypress()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func processKeypress() error {
	r, err := ReadKey()
	if err != nil {
		return err
	}
	switch r {
	case CtrlQ:
		{
			RefreshScreen()
			return errors.New("Stop command")
		}
	case TOP, DOWN, LEFT, RIGHT, ZERO, DOLLAR: //vim movement
		{
			MoveCursor(r)
		}
	}
	return nil
}
