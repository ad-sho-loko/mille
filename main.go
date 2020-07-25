package main

import (
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

// Key Definitions
const (
	DummyKey   = -1
	ControlA   = 1
	ControlB   = 2
	ControlC   = 3
	ControlE   = 5
	ControlF   = 6
	ControlH   = 8
	Tab        = 9
	Enter      = 13
	ControlN   = 14
	ControlP   = 16
	ControlS   = 19
	ControlV   = 22
	BackSpace  = 127
	ArrowUp    = 1000
	ArrowDown  = 1001
	ArrowRight = 1002
	ArrowLeft  = 1003
)

// Color Definition
type color int

const (
	DummyColor color = 37
	FgGreen          = 32
	FgCyan           = 36
	BgBlack          = 40
	BgCyan           = 46
)

type messageType int

const (
	resetMessage messageType = iota + 1
)

type Keyword string

const (
	Break       Keyword = "break"
	Default             = "default"
	Func                = "func"
	Interface           = "interface"
	Select              = "select"
	Case                = "case"
	Defer               = "defer"
	Go                  = "go"
	Map                 = "map"
	Struct              = "struct"
	Chan                = "chan"
	Else                = "else"
	Goto                = "goto"
	Package             = "package"
	Switch              = "switch"
	Const               = "const"
	Fallthrough         = "fallthrough"
	If                  = "if"
	Range               = "range"
	Type                = "type"
	Continue            = "continue"
	For                 = "for"
	Import              = "import"
	Return              = "return"
	Var                 = "var"
)

var keywordColor = map[Keyword]color{
	Break:       FgCyan,
	Default:     FgCyan,
	Interface:   FgCyan,
	Select:      FgCyan,
	Case:        FgCyan,
	Defer:       FgCyan,
	Go:          FgCyan,
	Map:         FgCyan,
	Struct:      FgCyan,
	Chan:        FgCyan,
	Else:        FgCyan,
	Goto:        FgCyan,
	Switch:      FgCyan,
	Const:       FgCyan,
	Fallthrough: FgCyan,
	Return:      FgCyan,
	Range:       FgCyan,
	Type:        FgCyan,
	Continue:    FgCyan,
	For:         FgCyan,
	If:          FgCyan,
	Package:     FgCyan,
	Import:      FgCyan,
	Func:        FgCyan,
	Var:         FgCyan,
}

type Editor struct {
	filePath string
	keyChan  chan rune
	timeChan chan messageType
	crow     int
	ccol     int
	scroolrow  int
	rows     []*Row
	terminal *Terminal
	n        int  // numberOfRows
	debug    bool // for debug
}

type Terminal struct {
	termios *unix.Termios
	width   int
	height  int
}

type Row struct {
	chars *GapTable
}

func (e *Editor) debugPrint(a ...interface{}) {
	if e.debug {
		_, _ = fmt.Fprintln(os.Stderr, a...)
	}
}

func (e *Editor) debugDetailPrint(a ...interface{}) {
	if e.debug {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", a...)
	}
}

func (e *Editor) debugRowRunes() {
	if e.debug {
		i := 0
		for i < e.n {
			_, _ = fmt.Fprintln(os.Stderr, i, ":", e.rows[i].chars.Runes())
			i += 1
		}
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
	e.setBgColor(BgCyan)
	defer e.setBgColor(BgBlack)

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

func (e *Editor) writeWithColor(b []byte, colors []color) {
	var newBuf []byte

	for i, c := range colors {
		s := fmt.Sprintf("\033[%dm", c)
		newBuf = append(newBuf, []byte(s)...)
		newBuf = append(newBuf, b[i])
	}

	syscall.Write(0, newBuf)
}

func (e *Editor) highlight(b []byte) []color {
	colors := make([]color, len(b))
	for i := range colors {
		colors[i] = DummyColor
	}

	// ASCII-only
	ascii := string(b)

	// Keywords
	for key := range keywordColor {
		index := strings.Index(ascii, string(key))
		if index != -1 {
			for i := 0; i < len(string(key)); i += 1 {
				colors[index+i] = keywordColor[key]
			}
		}
	}

	// String Literal
	isStringLit := false
	for i, b := range ascii {
		if b == '"' || isStringLit {
			if b == '"' { isStringLit = !isStringLit }
			colors[i] = FgGreen
		}
	}

	return colors
}

func (e *Editor) writeRow(r *Row) {
	var buf []byte

	for _, r := range r.chars.Runes() {
		buf = append(buf, []byte(string(r))...)
	}

	e.moveCursor(e.crow, 0)
	e.flushRow()

	// If the extension of fileName is .go, write with highlights.
	if filepath.Ext(e.filePath) == ".go" {
		colors := e.highlight(buf)
		e.writeWithColor(buf, colors)
	} else {
		e.write(buf)
	}
}

func (e *Editor) flush() {
	e.write([]byte("\033[2J"))
}

func (e *Editor) flushRow() {
	e.write([]byte("\033[2K"))
}

func (e *Editor) setBgColor(color color) {
	s := fmt.Sprintf("\033[%dm", color)
	e.write([]byte(s))
}

func (e *Editor) moveCursor(row, col int) {
	s := fmt.Sprintf("\033[%d;%dH", row+1, col+1) // 0-origin to 1-origin
	e.write([]byte(s))
}

func (e *Editor) updateRowRunes(row *Row) {
	e.debugPrint("DEBUG: row updated at", e.crow, "for", row.chars.Runes())
	e.writeRow(row)
}

func (e *Editor) refreshAllRows() {
	for i := 0; i < e.terminal.height; i += 1 {
		e.crow = i
		e.writeRow(e.rows[e.scroolrow + i])
	}

	e.setRowCol(0, 0)
}


func (e *Editor) setRowPos(row int) {
	if row < 0 {
		if e.scroolrow > 0 {
			e.scroolrow -= 1
			e.refreshAllRows()
		}
		row = 0
	}

	if row >= e.n {
		row = e.n - 1
	}

	if row >= e.terminal.height {
		e.scroolrow += 1
		row = e.terminal.height - 1
		e.refreshAllRows()
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

func (r *Row) len() int        { return r.chars.Len() }
func (r *Row) visibleLen() int { return r.chars.VisibleLen() }

func (e *Editor) deleteRune(row *Row, col int) {
	row.deleteAt(col)
	e.updateRowRunes(row)
	e.setRowCol(e.crow, e.ccol-1)
}

func (e *Editor) insertRune(row *Row, col int, newRune rune) {
	row.insertAt(col, newRune)
	e.updateRowRunes(row)
}

func (e *Editor) deleteRow(row int) {
	gt := NewGapTable(128)
	r := &Row{
		chars: gt,
	}

	e.rows[row] = r

	prevRowPos := e.crow
	e.crow = row
	e.updateRowRunes(r)
	e.crow = prevRowPos
}

func (e *Editor) replaceRune(col int, newRune []rune) {
	gt := NewGapTable(128)

	for _, r := range newRune {
		gt.AppendRune(r)
	}

	r := &Row{
		chars: gt,
	}

	prevRowPos := e.crow
	e.crow = col
	e.rows[e.crow] = r
	e.updateRowRunes(r)
	e.crow = prevRowPos
}

func (e *Editor) copyRow(dst int, src int) {
	r := &Row{
		chars: e.rows[src].chars,
	}

	e.rows[dst] = r
	e.updateRowRunes(r)
}

func (e *Editor) reallocBufferIfNeeded() {
	if e.n == len(e.rows) {
		newCap := cap(e.rows) * 2
		newRows := make([]*Row, newCap)
		copy(newRows, e.rows)
		e.rows = newRows
		e.debugPrint("DEBUG: realloc occurred")
	}
}

func (e *Editor) numberOfRunesInRow() int { return e.rows[e.crow].chars.Len() }

func (e *Editor) backspace() {
	row := e.rows[e.crow]

	if e.ccol == 0 {
		if e.crow > 0 {
			e.n -= 1
			e.crow -= 1

			prevRow := e.rows[e.crow]

			restoreRowPos := e.crow
			restoreColPos := prevRow.len() - 1

			// Update the previous row.
			newRunes := append([]rune{}, prevRow.chars.Runes()[:prevRow.len()-1]...)
			newRunes = append(newRunes, row.chars.Runes()...)
			e.replaceRune(e.crow, newRunes)

			// Update the trailing rows.
			e.crow += 1
			for e.crow < e.n {
				e.copyRow(e.crow, e.crow+1)
				e.crow += 1
			}

			// Delete the last row
			e.deleteRow(e.n)
			e.setRowCol(restoreRowPos, restoreColPos)
		}
	} else {
		e.deleteRune(row, e.ccol-1)
	}

	e.debugRowRunes()
}

func (e *Editor) back() {
	if e.ccol == 0 {
		if e.crow > 0 {
			e.setRowCol(e.crow - 1, e.rows[e.crow - 1].visibleLen())
		}
	} else {
		e.setRowCol(e.crow, e.ccol-1)
	}
}

func (e *Editor) next() {
	if e.ccol >= e.rows[e.crow].visibleLen() {
		if e.crow+1 < e.n {
			e.setRowCol(e.crow, 0)
		}
	} else {
		e.setRowCol(e.crow, e.ccol+1)
	}
}

func (e *Editor) newLine() {
	// Update the trailing rows.
	newLineRowPos := e.crow
	e.crow = e.n

	for e.crow > newLineRowPos+1 {
		e.copyRow(e.crow, e.crow-1)
		e.crow -= 1
	}

	e.n += 1
	e.reallocBufferIfNeeded()

	newLineRow := e.rows[newLineRowPos]

	// Update the next row.
	nextRowRunes := append([]rune{}, newLineRow.chars.Runes()[e.ccol:]...)
	e.replaceRune(e.crow, nextRowRunes)

	// Update the current row.
	currentRowNewRunes := append([]rune{}, newLineRow.chars.Runes()[:e.ccol]...)
	currentRowNewRunes = append(currentRowNewRunes, '\n')
	e.setRowCol(newLineRowPos, 0)
	e.replaceRune(e.crow, currentRowNewRunes)

	e.setRowCol(e.crow+1, 0)
	e.debugRowRunes()
}

func existsFile(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func saveFile(filePath string, rows []*Row) {
	sb := strings.Builder{}

	for _, r := range rows {
		if r.len() >= 1 {
			for _, ch := range r.chars.Runes() {
				sb.WriteRune(ch)
			}
		}
	}

	_ = ioutil.WriteFile(filePath, []byte(sb.String()), 0644)
}

func loadFile(filePath string) *Editor {
	e := &Editor{
		crow:     0,
		ccol:     0,
		scroolrow: 0,
		filePath: filePath,
		keyChan:  make(chan rune),
		timeChan: make(chan messageType),
		n:        1,
	}

	rows := makeRows()

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	gt := NewGapTable(128)

	for _, b := range bytes {
		// ASCII-only
		gt.AppendRune(rune(b))

		if b == '\n' {
			rows[e.n - 1] = &Row { chars: gt }
			e.n += 1
			gt = NewGapTable(128)
		}
	}

	rows[e.n - 1] = &Row{ chars: gt }
	e.rows = rows

	return e
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
				return DummyKey, 0
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
		case ControlA:
			e.setRowCol(e.crow, 0)

		case ControlB, ArrowLeft:
			e.back()

		case ControlC:
			e.exit()
			return

		case ControlE:
			e.setRowCol(e.crow, e.numberOfRunesInRow())

		case ControlF, ArrowRight:
			e.next()

		case ControlH, BackSpace:
			e.backspace()

		case ControlN, ArrowDown:
			e.setRowCol(e.crow + 1, e.ccol)

		case Tab:
			for i:=0; i<4; i+=1 {
				e.insertRune(e.rows[e.crow], e.ccol, rune(' '))
			}
			e.setColPos(e.ccol + 4)

		case Enter:
			e.newLine()

		case ControlS:
			saveFile(e.filePath, e.rows)
			e.writeHelpMenu("Saved!")
			e.timeChan <- resetMessage

		case ControlP, ArrowUp:
			e.setRowCol(e.crow - 1, e.ccol)

		// for debug
		case ControlV:
			e.debugDetailPrint(e)

		default:
			e.debugPrint(r)
			e.insertRune(e.rows[e.crow], e.ccol, r)
			e.setColPos(e.ccol + 1)
		}
	}
}

func (e *Editor) pollTimerEvent() {
	for {
		switch <-e.timeChan {
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

func makeRows() []*Row {
	var rows = make([]*Row, 128)
	for i := range rows {
		rows[i] = &Row{
			chars: NewGapTable(128),
		}
	}
	return rows
}

func newEditor(filePath string, debug bool) *Editor {
	terminal := newTerminal(0)

	if existsFile(filePath) {
		e := loadFile(filePath)
		e.debug = debug
		e.terminal = terminal
		return e
	}

	rows := makeRows()
	return &Editor{
		crow:     0,
		ccol:     0,
		scroolrow: 0,
		rows:     rows,
		filePath: filePath,
		keyChan:  make(chan rune),
		timeChan: make(chan messageType),
		terminal: terminal,
		n:        1,
		debug:    debug,
	}
}

func run(filePath string, debug bool) {
	e := newEditor(filePath, debug)
	e.initTerminal()
	e.refreshAllRows()

	go e.readKeys()
	go e.pollTimerEvent()
	e.interpretKey()
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 || flag.NArg() > 3 {
		fmt.Println("Usage: ./mille <filename> [--debug]")
		return
	}

	debug := flag.NArg() == 2 && flag.Arg(1) == "--debug"
	run(flag.Arg(0), debug)
}
