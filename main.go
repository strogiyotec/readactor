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
	//buffer := make([]byte, 1)
	//for cc, err := os.Stdin.Read(buffer); err == nil && cc == 1; cc, err = os.Stdin.Read(buffer) {
	//	var r rune
	//	r = rune(buffer[0])
	//	//quit on ctrl q
	//	if r == CtrlQ {
	//		break
	//	}
	//	if strconv.IsPrint(r) {
	//		fmt.Printf("%d  %c\r\n", buffer[0], r)
	//	} else {
	//		fmt.Printf("%d\r\n", buffer[0])
	//	}
	//}
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
	case TOP, DOWN, LEFT, RIGHT: //vim movement
		{
			MoveCursor(r)
		}
	}
	return nil
}
