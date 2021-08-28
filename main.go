package main

import "fmt"

func main() {
	fmt.Println("Hello From Readactor")
}

type Termios struct {
	Iflag  uint32   //Input mode flags
	Oflag  uint32   //Output mode flags
	Cflag  uint32   //Control mode flags
	Lflag  uint32   // Local mode flags
	Cc     [20]byte //Control Characters
	Ispeed uint32   //Input speed
	Ospeed uint32   //Output speed
}
