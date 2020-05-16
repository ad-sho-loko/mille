package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
	"unicode/utf8"
)

// Key Definitions
const (
	ControlC = 3
)

type Editor struct {
	fileName string
	rawState *terminal.State
	keyChan chan rune
	row int
	col int
	width int
	height int
}

func (e *Editor) initTerminal() {
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

func (e *Editor) flush() {
	e.write([]byte("\033[2J"))
}

func (e *Editor) moveCursor(row, col int) {
	s := fmt.Sprintf("\033[%d;%dH",row, col)
	e.write([]byte(s))
}

func (e *Editor) write(b []byte) {
	syscall.Write(0, b)
}

func (e *Editor) writeUtf8() {
}

func (e *Editor) readKeys() {
	buf := make([]byte, 64)

	for {
		if n, err := syscall.Read(0, buf); err == nil {
			b := buf[:n]
			for {
				r, n := utf8.DecodeRune(b)
				if n == 0 {
					break
				}
				e.keyChan <- r
				b = b[n:]
			}
		}
	}
}

func (e *Editor) handleEvents() {
}

func (e *Editor) restoreTerminal() {
	if err := terminal.Restore(0, e.rawState); err != nil {
		panic("Cannot restore from raw mode.")
	}
}

func (e *Editor) interpret() {
	for {
		char := <- e.keyChan
		switch char {
		case ControlC:
			e.restoreTerminal()
			return

		default:
			fmt.Println(string(char))
		}
	}
}

func run(fileName string) {
	e := &Editor{
		row: 0,
		col: 0,
		fileName: fileName,
		keyChan: make(chan rune),
	}

	e.initTerminal()
	e.flush()
	e.moveCursor(0, 0)

	go e.readKeys()
	e.interpret()
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
