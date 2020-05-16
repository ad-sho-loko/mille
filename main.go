package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

const (
	ControlC = 3
)

type Editor struct {
	fileName string
	rawState *terminal.State
	width int
	height int
}

var e = &Editor{}

func initTerminal() {
	state, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}

	width, height, err := terminal.GetSize(0)
	if err != nil {
		panic(err)
	}

	e.width = width
	e.height = height
	e.rawState = state
}

func readKey(keyCh chan []byte) {
	buf := make([]byte, 256)

	for {
		if n, err := syscall.Read(0, buf); err == nil {
			keyCh <- buf[:n]
		}
	}
}

func handleEvents() {
}

func restoreTerminal() {
	if err := terminal.Restore(0, e.rawState); err != nil {
		panic("Cannot restore from raw mode.")
	}
}

func run(fileName string) {
	e.fileName = fileName
	initTerminal()

	keyCh := make(chan []byte)
	go readKey(keyCh)

	for {
		ch := <- keyCh
		fmt.Println(ch)

		if ch[0] == ControlC {
			restoreTerminal()
			return
		}
	}
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		fmt.Println("Usage: go run . <filename>")
		return
	}

	run(args[0])
}
