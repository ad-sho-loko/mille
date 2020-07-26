package main

type GapTable struct {
	array              []rune
	startPieceIndex    int
	endPieceIndex      int
	invisibleRuneCount int
}

func NewGapTable(cap int) *GapTable {
	return &GapTable{
		array:              make([]rune, cap),
		startPieceIndex:    0,
		endPieceIndex:      cap - 1,
		invisibleRuneCount: 0,
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

	return g.array[index+g.endPieceIndex-g.startPieceIndex+1]
}

func (g *GapTable) SetAt(index int, r rune) {
	if index < g.startPieceIndex {
		g.array[index] = r
		return
	}

	g.array[index+g.endPieceIndex-g.startPieceIndex+1] = r
}

func (g *GapTable) AppendRune(r rune) {
	g.InsertAt(g.Len(), r)
}

// See gap_table_test.go how it works.
func (g *GapTable) InsertAt(index int, r rune) {
	if r == 0x0a {
		g.invisibleRuneCount += 1
	}

	if index == g.startPieceIndex {
		// Insert E at #
		// before: [A, B, C, #, x, x, D]
		// after:  [A, B, C, E, x, x, D]
		// O(1) if not reallocating the array
		g.array[g.startPieceIndex] = r
		g.startPieceIndex += 1
	} else if index > g.startPieceIndex {
		if index >= g.Len() {
			// e.g.) Insert F at #
			// before: [A, B, C, x, x, D, E] #
			//                   ^  ^
			//                   s  e
			// after:  [A, B, C, x, x, D, E, F]
			//                   ^  ^
			//                   s  e
			copyTarget := g.array[g.endPieceIndex+1 : g.Cap()]
			_ = copy(g.array[g.endPieceIndex:g.Cap()-1], copyTarget)
			g.array[g.Cap() - 1] = r
			g.endPieceIndex -= 1
		} else {
			// e.g.) Insert G at #
			// before: [A, B, C, x, x, x, D, #, E, F]
			//                   ^     ^
			//                   s     e
			// after:  [A, B, C, D, x, x, D, G, E, F]
			//                      ^     ^
			//                      s     e
			copyTarget := g.array[g.endPieceIndex+1 : index+g.endPieceIndex-g.startPieceIndex+2]
			n := copy(g.array[g.startPieceIndex:g.startPieceIndex+len(copyTarget)], copyTarget)
			g.SetAt(index, r)
			g.startPieceIndex += n
			g.endPieceIndex += n - 1
		}
	} else {
		// e.g.) Insert G at #
		// before: [A, B, #, C, x, x, x, D, E, F]
		//                      ^     ^
		//                      s     e
		// after:  [A, B, G, x, x, x, C, D, E, F]
		//                   ^     ^
		//                   s     e
		copyTarget := g.array[index:g.startPieceIndex]
		n := copy(g.array[g.endPieceIndex-len(copyTarget)+1:g.endPieceIndex+1], copyTarget)
		g.array[index] = r
		g.startPieceIndex = index + 1
		g.endPieceIndex -= n
	}

	if g.startPieceIndex > g.endPieceIndex {
		g.realloc()
	}
}

func (g *GapTable) DeleteAt(index int) {
	if g.At(index) == 0x0a {
		g.invisibleRuneCount -= 1
	}

	if index == g.startPieceIndex - 1 {
		g.startPieceIndex -= 1
	} else if index < g.startPieceIndex {
		copyTarget := g.array[index:g.startPieceIndex]
		n := copy(g.array[g.endPieceIndex-len(copyTarget)+1:g.endPieceIndex+1], copyTarget)
		g.startPieceIndex -= n
		g.endPieceIndex -= n - 1
	} else {
		copyTarget := g.array[g.endPieceIndex+1 : index+g.endPieceIndex-g.startPieceIndex+2]
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

func (g *GapTable) VisibleLen() int {
	return g.Len() - g.invisibleRuneCount
}

func (g *GapTable) Runes() []rune {
	return append(g.array[:g.startPieceIndex], g.array[g.endPieceIndex+1:]...)
}

func (g *GapTable) RunesString() string {
	runes := append(g.array[:g.startPieceIndex], g.array[g.endPieceIndex+1:]...)
	return string(runes)
}
