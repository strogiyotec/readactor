package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	. "github.com/strogiyotec/readactor/termio"
)

func main() {
	fmt.Println("Hello From Readactor")
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
func getCursorPosition(rows *int, cols *int) int {
	io.WriteString(os.Stdout, "\x1b[6n")
	fmt.Printf("\r\n")
	var buffer [1]byte
	var cc int
	for cc, _ = os.Stdin.Read(buffer[:]); cc == 1; cc, _ = os.Stdin.Read(buffer[:]) {
		if buffer[0] > 20 && buffer[0] < 0x7e {
		} else {
			fmt.Printf("%d\r\n", buffer[0])
		}
		fmt.Printf("%d ('%c')\r\n", buffer[0], buffer[0])
	}

	ReadKey()
	return -1
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
	}
	return nil
}
