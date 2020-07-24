package main

import (
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

// TODO:
// Less Bug
// Many Key Shortcut
// Multi-Byte
// Loading File

// Key Definitions
const (
	Dummy      = -1
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

type messageType int
const (
	resetMessage messageType = iota + 1
)

type Editor struct {
	filePath string
	keyChan  chan rune
	timeChan chan messageType
	crow     int
	ccol     int
	rows     []*Row
	terminal *Terminal
	n int  // numberOfRows
}

type Terminal struct {
	termios *unix.Termios
	width   int
	height  int
}

type Row struct {
	chars *GapTable
	render *GapTable
}

func debug(a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, a)
}

func (e *Editor) debugRowRunes() {
	i := 0
	for i < e.n {
		_, _ = fmt.Fprintln(os.Stderr, i, ":", e.rows[i].chars.RunesString())
		i += 1
	}
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
	e.writeHelpMenu("HELP: Ctrl+S = Save / Ctrl+C = Quit")
	e.writeStatusBar()
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) writeHelpMenu(message string) {
	prevRow, prevCol := e.crow, e.ccol

	for i, ch := range message {
		e.moveCursor(e.terminal.height+1, i)
		e.write([]byte(string(ch)))
	}

	for i := len(message); i < e.terminal.width; i++ {
		e.moveCursor(e.terminal.height+1, i)
		e.write([]byte{' '})
	}

	e.moveCursor(prevRow, prevCol)
}

func (e *Editor) writeStatusBar() {
	e.setBgColor(Cyan)
	defer e.setBgColor(Black)

	// Write file name
	for i, ch := range e.filePath {
		e.moveCursor(e.terminal.height, i)
		e.write([]byte(string(ch)))
	}

	// Write Spacer
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

	for _, r := range r.chars.Runes() {
		buf = append(buf, []byte(string(r))...)
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

func (e *Editor) updateRowRunes(row *Row) {
	e.writeRow(row)
}

// Models
func (r *Row) deleteAt(col int) {
	if col >= r.len() {
		return
	}

	r.chars.DeleteAt(col)
}

func (r *Row) insertAt(colPos int, newRune rune) {
	if colPos > r.len() {
		colPos = r.len()
	}

	r.chars.InsertAt(colPos, newRune)
}

func (r *Row) len() int { return r.chars.Len() }
func (r *Row) visibleLen() int { return r.chars.VisibleLen() }

func (e *Editor) deleteRune(row *Row, col int) {
	if e.ccol == 0 {
		if e.crow != 0 {
			// e.g) row=1, col=0
			e.n -= 1
			e.crow -= 1
			row.deleteAt(e.numberOfRunesInRow() - 1)
			e.updateRowRunes(row)
			e.setRowCol(e.crow, e.numberOfRunesInRow() - 1)
		}
	} else {
		row.deleteAt(col)
		e.updateRowRunes(row)
		e.setRowCol(e.crow, e.ccol - 1)
	}
}

func (e *Editor) insertRune(row *Row, col int, newRune rune) {
	row.insertAt(col, newRune)
	e.updateRowRunes(row)
}

func (e *Editor) replaceRow(newRune []rune) {
	gt := NewGapTable(128)

	for _, r := range newRune {
		gt.AppendRune(r)
	}

	r := &Row{
		chars: gt,
	}

	e.rows[e.crow] = r
	e.updateRowRunes(r)
}

func (e *Editor) setRowPos(row int) {
	if row < 0 {
		row = 0
	}

	if row >= e.n {
		row = e.n - 1
	}

	if row >= e.terminal.height {
		row = e.terminal.height - 1
	}

	e.crow = row
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) setColPos(col int) {
	if col < 0 {
		col = 0
	}

	if col >= e.rows[e.crow].visibleLen() {
		col = e.rows[e.crow].visibleLen()
	}

	if col >= e.terminal.width {
		col = e.terminal.width - 1
	}

	e.ccol = col
	e.moveCursor(e.crow, e.ccol)
}

func (e *Editor) setRowCol(row int, col int) {
	if row > e.n && col > e.rows[e.crow].visibleLen() {
		return
	}

	e.setRowPos(row)
	e.setColPos(col)
}

func (e *Editor) numberOfRunesInRow() int { return e.rows[e.crow].chars.Len() }

func (e *Editor) backspace() {
	row := e.rows[e.crow]
	e.deleteRune(row, e.ccol-1)
}

func (e *Editor) back() {
	if e.ccol == 0 {
		if e.crow > 0 {
			e.setRowCol(e.crow - 1, e.rows[e.crow - 1].visibleLen())
		}
	} else {
		e.setRowCol(e.crow, e.ccol - 1)
	}
}

func (e *Editor) next() {
	if e.ccol >= e.rows[e.crow].visibleLen() {
		if e.crow + 1 < e.n {
			e.setRowCol(e.crow+1, 0)
		}
	} else {
		e.setRowCol(e.crow, e.ccol+1)
	}
}

func (e *Editor) newLine() {
	e.n += 1

	// Update the current row.
	currentRow := e.rows[e.crow]
	currentRowNewRunes := currentRow.chars.Runes()[:e.ccol]
	nextRowNewPrefixRunes := append([]rune{}, currentRow.chars.Runes()[e.ccol:currentRow.chars.Len()]...)
	currentRowNewRunes = append(currentRowNewRunes, '\n')
	e.replaceRow(currentRowNewRunes)
	e.setRowCol(e.crow+1, 0)

	// Update the next row.
	nextRow := e.rows[e.crow]
	nextRowRunes := append(nextRowNewPrefixRunes, nextRow.chars.Runes()...)
	e.replaceRow(nextRowRunes)
	e.setRowCol(e.crow, 0)

	e.debugRowRunes()
}

func (e *Editor) saveFile() {
	sb := strings.Builder{}

	for _, r := range e.rows {
		if r.len() >= 1 {
			for _, ch := range r.chars.Runes() {
				sb.WriteRune(ch)
			}
		}
	}

	ioutil.WriteFile("tmp", []byte(sb.String()), 0644)
}

func (e *Editor) exit() {
	e.restoreTerminal(0)
}

func (e *Editor) parseKey(b []byte) (rune, int) {
	// Try parsing escape sequence
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
			default:
				return Dummy, 0
			}
		}
	}

	// parse bytes as UTF-8.
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
			e.back()

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
			e.newLine()

		case ControlS:
			e.saveFile()
			e.writeHelpMenu("Saved!")
			e.timeChan <- resetMessage

		case ControlP, ArrowUp:
			e.setRowCol(e.crow - 1, e.ccol)

		default:
			e.insertRune(e.rows[e.crow], e.ccol, r)
			e.setColPos(e.ccol + 1)
		}
	}
}

func (e *Editor) timerEventPoller(){
	for {
		switch <- e.timeChan {
		case resetMessage:
			t := time.NewTimer(2 * time.Second)
			<-t.C
			e.writeHelpMenu("")
		}
	}
}

func newTerminal(fd int) *Terminal {
	termios := makeRaw(fd)
	width, height := getWindowSize(fd)

	terminal := &Terminal{
		termios: termios,
		width:   width,
		height:  height - 2, // for status, message bar
	}

	return terminal
}

func makeRow() []*Row {
	// TODO
	var rows = make([]*Row, 32)
	for i := range rows {
		rows[i] = &Row {
			chars: NewGapTable(128),
		}
	}
	return rows
}

func newEditor(filePath string) *Editor {
	rows := makeRow()
	terminal := newTerminal(0)

	e := &Editor{
		crow:     0,
		ccol:     0,
		rows:     rows,
		filePath: filePath,
		keyChan:  make(chan rune),
		timeChan: make(chan messageType),
		terminal: terminal,
		n: 1,
	}

	return e
}

func run(filePath string) {
	e := newEditor(filePath)
	e.initTerminal()
	go e.readKeys()
	go e.timerEventPoller()
	e.interpretKey()
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		fmt.Println("Usage: ./mille <filename>")
		return
	}

	run(args[0])
}
