package main

import (
	"fmt"
	"os"
	"strconv"

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
	buffer := make([]byte, 1)
	for cc, err := os.Stdin.Read(buffer); buffer[0] != 'q' && err == nil && cc == 1; cc, err = os.Stdin.Read(buffer) {
		var r rune
		r = rune(buffer[0])
		if strconv.IsPrint(r) {
			fmt.Printf("%d  %c\r\n", buffer[0], r)
		} else {
			fmt.Printf("%d\r\n", buffer[0])
		}
	}
}
