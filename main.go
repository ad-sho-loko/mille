package main

import (
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"strings"
	"syscall"
	"unicode/utf8"
)

// Key Definitions
const (
	ControlB   = 2
	ControlC   = 3
	ControlF   = 6
	ControlH   = 8
	Enter      = 13
	ControlN   = 14
	ControlP   = 16
	ControlS   = 19
	BackSpace  = 127
	ArrowUp    = 1000
	ArrowDown  = 1001
	ArrowRight = 1002
	ArrowLeft  = 1003
)

// Color Definition
const (
	Black = 40
	Cyan  = 46
)

type Editor struct {
	filePath string
	keyChan  chan rune
	crow     int
	ccol     int
	rows     []*Row
	terminal *Terminal
}

type Terminal struct {
	termios *unix.Termios
	width   int
	height  int
}

type Row struct {
	numberOfRunes int
	runes         []rune
}

// Terminal
func makeRaw(fd int) *unix.Termios {
	termios, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		panic(err)
	}

	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, termios); err != nil {
		panic(err)
	}

	return termios
}

func (e *Editor) restoreTerminal(fd int) {
	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, e.terminal.termios); err != nil {
		panic(err)
	}
}

func getWindowSize(fd int) (int, int) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		panic(err)
	}
	return int(ws.Col), int(ws.Row)
}

func (e *Editor) initTerminal() {
	e.flush()
	e.writeHelpMenu()
	e.writeStatusBar()
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) writeHelpMenu() {
	message := "HELP: Ctrl+S = Save / Cntl+C = Quit"
	for i, ch := range message {
		e.moveCursor(e.terminal.height+1, i)
		e.write([]byte(string(ch)))
	}

	for i := len(message); i < e.terminal.width; i++ {
		e.moveCursor(e.terminal.height+1, i)
		e.write([]byte{' '})
	}
}

func (e *Editor) writeStatusBar() {
	e.setBgColor(Cyan)
	defer e.setBgColor(Black)

	// Write file name
	for i, ch := range e.filePath {
		e.moveCursor(e.terminal.height, i)
		e.write([]byte(string(ch)))
	}

	for i := len(e.filePath); i < e.terminal.width; i++ {
		e.moveCursor(e.terminal.height, i)
		e.write([]byte{' '})
	}
}

// Views
func (e *Editor) write(b []byte) {
	syscall.Write(0, b)
}

func (e *Editor) writeRow(r *Row) {
	var buf []byte

	for _, s := range r.runes {
		buf = append(buf, []byte(string(s))...)
	}

	e.flushRow()
	e.moveCursor(e.crow, 0)
	e.write(buf)
}

func (e *Editor) flush() {
	e.write([]byte("\033[2J"))
}

func (e *Editor) flushRow() {
	e.write([]byte("\033[2K"))
}

func (e *Editor) setBgColor(color int) {
	s := fmt.Sprintf("\033[%dm", color)
	e.write([]byte(s))
}

func (e *Editor) moveCursor(row, col int) {
	s := fmt.Sprintf("\033[%d;%dH", row+1, col+1) // 0-origin to 1-origin
	e.write([]byte(s))
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

	row.numberOfRunes -= 1
	row.runes = newRune

	e.writeRow(row)
}

func (e *Editor) insertAt(row *Row, col int, newRune rune) {
	if col >= len(row.runes) {
		return
	}

	var newRunes []rune

	for i, r := range row.runes {
		if i == col {
			newRunes = append(newRunes, newRune)
		}
		newRunes = append(newRunes, r)
	}

	row.numberOfRunes += 1
	row.runes = newRunes

	e.writeRow(row)
}

func (e *Editor) setRow(row int) {
	if row < 0 {
		row = 0
	}

	if row >= e.terminal.height {
		row = e.terminal.height - 1
	}

	e.crow = row
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) setCol(col int) {
	if col < 0 {
		col = 0
	}

	if col >= e.terminal.width {
		col = e.terminal.width - 1
	}

	e.ccol = col
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) setRowCol(row int, col int) {
	e.setRow(row)
	e.setCol(col)
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) numberOfRunesInRow() int { return e.rows[e.crow].numberOfRunes }

func (e *Editor) appendChar(row int, r rune) {
	e.rows[row].numberOfRunes += 1
	e.rows[row].runes[e.ccol] = r
	e.write([]byte(string(r)))
}

func (e *Editor) backspace() {
	if e.ccol > 0 {
		row := e.rows[e.crow]
		e.deleteAt(row, e.ccol-1)
		e.setRowCol(e.crow, e.ccol-1)
	} else {
		e.setRow(e.crow - 1)
		e.setCol(e.numberOfRunesInRow() - 1)
	}
}

func (e *Editor) next() {
	if e.ccol >= e.rows[e.crow].numberOfRunes {
		e.setRowCol(e.crow+1, 0)
	} else {
		e.setRowCol(e.crow, e.ccol+1)
	}
}

func (e *Editor) enter() {
	e.appendChar(e.crow, '\n')
	e.setRowCol(e.crow+1, 0)
}

func (e *Editor) saveFile() {
	sb := strings.Builder{}

	for _, r := range e.rows {
		if r.numberOfRunes >= 1 {
			for _, ch := range r.runes {
				if ch == rune(byte(0)) {
					continue
				}
				sb.WriteRune(ch)
			}
			sb.WriteByte('\n')
		}
	}

	ioutil.WriteFile("tmp", []byte(sb.String()), 0644)
}

func (e *Editor) exit() {
	e.restoreTerminal(0)
}

func (e *Editor) parseKey(b []byte) (rune, int) {
	// try parsing escape sequence
	if len(b) == 3 {
		if b[0] == byte(27) && b[1] == '[' {
			switch b[2] {
			case 'A':
				return ArrowUp, 3
			case 'B':
				return ArrowDown, 3
			case 'C':
				return ArrowRight, 3
			case 'D':
				return ArrowLeft, 3
			}
		}
	}

	// Parse bytes as UTF-8.
	return utf8.DecodeRune(b)
}

func (e *Editor) readKeys() {
	buf := make([]byte, 64)

	for {
		if n, err := syscall.Read(0, buf); err == nil {
			b := buf[:n]
			for {
				r, n := e.parseKey(b)

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
		case ControlB, ArrowLeft:
			e.setRowCol(e.crow, e.ccol-1)

		case ControlC:
			e.exit()
			return

		case ControlF, ArrowRight:
			e.next()

		case ControlH, BackSpace:
			e.backspace()

		case ControlN, ArrowDown:
			e.setRowCol(e.crow+1, e.ccol)

		case Enter:
			e.enter()

		case ControlS:
			e.saveFile()

		case ControlP, ArrowUp:
			e.setRowCol(e.crow-1, e.ccol)

		default:
			e.insertAt(e.rows[e.crow], e.ccol, r)
			e.setCol(e.ccol + 1)
		}
	}
}

func makeRows() []*Row {
	rows := make([]*Row, 16)
	for i := range rows {
		rows[i] = &Row{
			numberOfRunes: 0,
			runes:         make([]rune, 128),
		}
	}
	return rows
}

func newTerminal(fd int) *Terminal {
	termios := makeRaw(fd)
	width, height := getWindowSize(fd)

	terminal := &Terminal{
		termios: termios,
		width:   width,
		height:  height - 2, // for status|help bar

	}

	return terminal
}

func newEditor(filePath string) *Editor {
	rows := makeRows()
	terminal := newTerminal(0)

	e := &Editor{
		crow:     0,
		ccol:     0,
		rows:     rows,
		filePath: filePath,
		keyChan:  make(chan rune),
		terminal: terminal,
	}

	return e
}

func run(filePath string) {
	e := newEditor(filePath)
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
