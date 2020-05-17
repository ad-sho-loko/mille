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
	ControlB  = 2
	ControlC  = 3
	ControlF  = 6
	ControlH  = 8
	Enter     = 13
	ControlN  = 14
	ControlP  = 16
	BackSpace = 127
)

type Editor struct {
	fileName string
	rawState *terminal.State
	keyChan  chan rune
	crow     int
	ccol     int
	rows     []*Row
	width    int
	height   int
}

type Row struct {
	n int
	runes []rune
}

// Terminal
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
	e.flush()
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) restoreTerminal() {
	if err := terminal.Restore(0, e.rawState); err != nil {
		panic("Cannot restore from raw mode.")
	}
}

// Views
func (e *Editor) flush() {
	e.write([]byte("\033[2J"))
}

func (e *Editor) flushRow(row int){
	e.write([]byte("\033[2K"))
}

func (e *Editor) moveCursor(row, col int) {
	s := fmt.Sprintf("\033[%d;%dH", row+1, col+1) // 0-origin to 1-origin
	e.write([]byte(s))
}

func (e *Editor) write(b []byte) {
	syscall.Write(0, b)
}

func (e *Editor) writeRow(r *Row) {
	var buf []byte

	for _, s := range r.runes {
		buf = append(buf, []byte(string(s))...)
	}

	e.flushRow(e.crow)
	e.moveCursor(e.crow, 0)
	syscall.Write(0, buf)
}

// Models
func (e *Editor) deleteAt(row *Row, col int) {
	if col >= len(row.runes) {
		return
	}

	var newRune []rune

	for i, r := range row.runes {
		if i != col {
			newRune = append(newRune, r)
		}
	}

	row.runes = newRune
	e.writeRow(row)
}

func (e *Editor) setRow(row int) {
	if row < 0 {
		row = 0
	}

	if row >= e.height {
		row = e.height - 1
	}

	e.crow = row
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) setCol(col int) {
	if col < 0 {
		col = 0
	}

	if col >= e.width {
		col = e.width - 1
	}

	e.ccol = col
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) setRowCol(row int, col int) {
	e.setRow(row)
	e.setCol(col)
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) setChar(row, col int, r rune) {
	e.rows[row].runes = append(e.rows[row].runes, r)
	e.write([]byte(string(r)))
}

func (e *Editor) backspace() {
	if e.ccol > 0 {
		row := e.rows[e.crow]
		e.deleteAt(row, e.ccol-1)
		e.setRowCol(e.crow, e.ccol-1)
	} else {
		e.setRowCol(e.crow-1, e.ccol)
	}
}

func (e *Editor) enter() {
	e.setRowCol(e.crow+1, 0)
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

func (e *Editor) interpretKey() {
	for {
		r := <-e.keyChan

		switch r {
		case ControlB:
			e.setRowCol(e.crow, e.ccol-1)

		case ControlC:
			e.restoreTerminal()
			return

		case ControlF:
			e.setRowCol(e.crow, e.ccol+1)

		case ControlH, BackSpace:
			e.backspace()

		case ControlN:
			e.setRowCol(e.crow+1, e.ccol)

		case Enter:
			e.enter()

		case ControlP:
			e.setRowCol(e.crow-1, e.ccol)

		default:
			e.setChar(e.crow, e.ccol, r)
			e.setCol(e.ccol + 1)
		}
	}
}

func run(fileName string) {
	// TODO: バッファの初期化
	rows := make([]*Row, 16)
	for i := range rows {
		rows[i] = &Row{
			n: i + 1,
			runes: []rune{},
		}
	}

	e := &Editor{
		crow:     0,
		ccol:     0,
		rows:     rows,
		fileName: fileName,
		keyChan:  make(chan rune),
	}

	e.initTerminal()
	go e.readKeys()
	e.interpretKey()
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
