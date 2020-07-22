package main

type GapTable struct {
	array []rune
	startPieceIndex int
	endPieceIndex int
}

func NewGapTable(cap int) *GapTable {
	return &GapTable{
		array: make([]rune, cap),
		startPieceIndex: 0,
		endPieceIndex: cap - 1,
	}
}

func (g *GapTable) realloc() {
	newCap := cap(g.array) * 2
	newArray := make([]rune, newCap)
	copy(newArray[:g.startPieceIndex], g.array[:g.startPieceIndex])

	newEndPieceIndex := newCap - cap(g.array[g.endPieceIndex+1:])
	copy(newArray[newEndPieceIndex:newCap], g.array[g.endPieceIndex+1:])

	g.array = newArray
	g.endPieceIndex = newEndPieceIndex - 1
}

func (g *GapTable) At(index int) rune {
	if index < g.startPieceIndex {
		return g.array[index]
	}

	return g.array[index + g.endPieceIndex - g.startPieceIndex + 1]
}

func (g *GapTable) SetAt(index int, r rune) {
	if index < g.startPieceIndex {
		g.array[index] = r
		return
	}

	g.array[index + g.endPieceIndex - g.startPieceIndex + 1] = r
}

func (g *GapTable) AppendRune(r rune) {
	g.InsertAt(g.Len(), r)
}

func (g *GapTable) InsertAt(index int, r rune) {
	if index == g.startPieceIndex {
		g.array[g.startPieceIndex] = r
		g.startPieceIndex += 1
	} else if index > g.startPieceIndex {
		if index >= g.Len() {
			// Insert E at #
			// before: [A, B, C, x, D] #
			// after:  [A, B, C, x, D, E]
			copyTarget := g.array[g.endPieceIndex+1: g.Cap()]
			_ = copy(g.array[g.endPieceIndex: g.Cap() - 1], copyTarget)
			g.array[g.Cap() - 1] = r
			g.endPieceIndex -= 1
		} else {
			copyTarget := g.array[g.endPieceIndex + 1: index + g.endPieceIndex - g.startPieceIndex + 2]
			n := copy(g.array[g.startPieceIndex:g.startPieceIndex+len(copyTarget)], copyTarget)
			g.SetAt(index, r)
			g.startPieceIndex += n
			g.endPieceIndex += n - 1
		}
	} else {
		copyTarget := g.array[index: g.startPieceIndex]
		n := copy(g.array[g.endPieceIndex - len(copyTarget) + 1 :g.endPieceIndex + 1], copyTarget)
		g.array[index] = r
		g.startPieceIndex = index + 1
		g.endPieceIndex -= n
	}

	if g.startPieceIndex > g.endPieceIndex {
		g.realloc()
	}
}

func (g *GapTable) DeleteAt(index int) {
	if index < g.startPieceIndex {
		copyTarget := g.array[index: g.startPieceIndex]
		n := copy(g.array[g.endPieceIndex - len(copyTarget) + 1 :g.endPieceIndex + 1], copyTarget)
		g.startPieceIndex -= n
		g.endPieceIndex -= n - 1
	} else {
		copyTarget := g.array[g.endPieceIndex + 1: index + g.endPieceIndex - g.startPieceIndex + 2]
		n := copy(g.array[g.startPieceIndex:g.startPieceIndex+len(copyTarget)], copyTarget)
		g.startPieceIndex += n - 1
		g.endPieceIndex += n
	}
}

func (g *GapTable) Len() int {
	return len(g.array) - g.endPieceIndex + g.startPieceIndex - 1
}

func (g *GapTable) Cap() int {
	return cap(g.array)
}

func (g *GapTable) Runes() []rune {
	return append(g.array[:g.startPieceIndex], g.array[g.endPieceIndex+1:]...)
}

func (g *GapTable) RunesString() string {
	// err(g.array, g.startPieceIndex, g.endPieceIndex)
	runes := append(g.array[:g.startPieceIndex], g.array[g.endPieceIndex+1:]...)
	return string(runes)
}
