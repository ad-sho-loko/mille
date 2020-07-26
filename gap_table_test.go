package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGapTable_Len(t *testing.T) {
	g := NewGapTable(8)
	assert.Equal(t, 0, g.Len())
}

func TestGapTable_Insert_Append(t *testing.T) {
	g := NewGapTable(4)
	g.AppendRune(r(97))
	g.AppendRune(r(98))
	g.AppendRune(r(99))

	assert.Equal(t, []rune{r(97), r(98), r(99), r(0)}, g.array)
	assert.Equal(t, 3, g.startPieceIndex)
	assert.Equal(t, 3, g.endPieceIndex)

	assert.Equal(t, "abc", g.RunesString())
}

func TestGapTable_Insert_Simple(t *testing.T) {
	g := NewGapTable(8)
	g.AppendRune(r(97))
	g.AppendRune(r(97))
	g.AppendRune(r(97))
	g.InsertAt(1, r(98))

	assert.Equal(t, []rune{r(97), r(98), r(97), r(0), r(0), r(0), r(97), r(97)}, g.array)
	assert.Equal(t, 2, g.startPieceIndex)
	assert.Equal(t, 5, g.endPieceIndex)

	assert.Equal(t, "abaa", g.RunesString())
}

func TestGapTable_Insert_Middle(t *testing.T) {
	g := NewGapTable(8)
	g.InsertAt(0, r(97))
	g.InsertAt(1, r(98))
	g.InsertAt(1, r(99))
	g.InsertAt(2, r(100))
	assert.Equal(t, []rune{r(97), r(99), r(100), r(0), r(0), r(0), r(0), r(98)}, g.array)

	assert.Equal(t, 3, g.startPieceIndex)
	assert.Equal(t, 6, g.endPieceIndex)

	assert.Equal(t, "acdb", g.RunesString())
}

func TestGapTable_Insert_Complex(t *testing.T) {
	g := NewGapTable(8)

	g.InsertAt(0, r(97))
	g.InsertAt(0, r(98))
	assert.Equal(t, 2, g.Len())
	assert.Equal(t, 1, g.startPieceIndex)
	assert.Equal(t, 6, g.endPieceIndex)
	assert.Equal(t, []rune{r(98), r(0), r(0), r(0), r(0), r(0), r(0), r(97)}, g.array)

	g.InsertAt(1, r(99))
	assert.Equal(t, 3, g.Len())
	assert.Equal(t, 2, g.startPieceIndex)
	assert.Equal(t, 6, g.endPieceIndex)
	assert.Equal(t, []rune{r(98), r(99), r(0), r(0), r(0), r(0), r(0), r(97)}, g.array)

	g.InsertAt(3, r(100))
	assert.Equal(t, 4, g.Len())
	assert.Equal(t, 2, g.startPieceIndex)
	assert.Equal(t, 5, g.endPieceIndex)
	assert.Equal(t, []rune{r(98), r(99), r(0), r(0), r(0), r(0), r(97), r(100)}, g.array)

	g.InsertAt(2, r(101))
	assert.Equal(t, []rune{r(98), r(99), r(101), r(0), r(0), r(0), r(97), r(100)}, g.array)
	assert.Equal(t, 5, g.Len())
	assert.Equal(t, 3, g.startPieceIndex)
	assert.Equal(t, 5, g.endPieceIndex)

	g.InsertAt(4, r(102))
	assert.Equal(t, 6, g.Len())
	assert.Equal(t, 5, g.startPieceIndex)
	assert.Equal(t, 6, g.endPieceIndex)
	assert.Equal(t, []rune{r(98), r(99), r(101), r(97), r(100), r(0), r(97), r(102)}, g.array)

	g.InsertAt(7, r(103))
	assert.Equal(t, 7, g.Len())
	assert.Equal(t, 5, g.startPieceIndex)
	assert.Equal(t, 5, g.endPieceIndex)
	assert.Equal(t, []rune{r(98), r(99), r(101), r(97), r(100), r(0), r(102), r(103)}, g.array)

	assert.Equal(t, g.At(0), r(98))
	assert.Equal(t, g.At(1), r(99))
	assert.Equal(t, g.At(2), r(101))
	assert.Equal(t, g.At(3), r(97))
	assert.Equal(t, g.At(4), r(100))
	assert.Equal(t, g.At(5), r(102))
	assert.Equal(t, g.At(6), r(103))
}

func TestGapTable_Reallocate_NoGap(t *testing.T) {
	g := NewGapTable(2)
	g.AppendRune(r(97))
	g.AppendRune(r(98))

	assert.Equal(t, 2, g.Len())
	assert.Equal(t, 4, g.Cap())
	assert.Equal(t, 2, g.startPieceIndex)
	assert.Equal(t, 3, g.endPieceIndex)
	assert.Equal(t, []rune{r(97), r(98), r(0), r(0)}, g.array)
}

func TestGapTable_Reallocate_Gap(t *testing.T) {
	g := NewGapTable(4)
	g.AppendRune(r(97))
	g.AppendRune(r(98))
	g.AppendRune(r(99))
	g.AppendRune(r(100))

	assert.Equal(t, 4, g.Len())
	assert.Equal(t, 8, g.Cap())
	assert.Equal(t, 4, g.startPieceIndex)
	assert.Equal(t, 7, g.endPieceIndex)
	assert.Equal(t, []rune{r(97), r(98), r(99), r(100), r(0), r(0), r(0), r(0)}, g.array)
}

func TestGapTable_Reallocate_Insert_Gap(t *testing.T) {
	g := NewGapTable(4)
	g.InsertAt(0, r(97))
	g.InsertAt(0, r(98))
	g.InsertAt(0, r(99))
	g.InsertAt(0, r(100))

	assert.Equal(t, 4, g.Len())
	assert.Equal(t, 8, g.Cap())
	assert.Equal(t, 1, g.startPieceIndex)
	assert.Equal(t, 4, g.endPieceIndex)
	assert.Equal(t, []rune{r(100), r(0), r(0), r(0), r(0), r(99), r(98), r(97)}, g.array)
}

func TestGapTable_DeleteAt_Gap(t *testing.T) {
	g := NewGapTable(4)
	g.InsertAt(0, r(97))
	g.InsertAt(0, r(98))

	g.DeleteAt(0)
	assert.Equal(t, 1, g.Len())
	assert.Equal(t, 0, g.startPieceIndex)
	assert.Equal(t, 2, g.endPieceIndex)
	assert.Equal(t, []rune{r(98), r(0), r(0), r(97)}, g.array)

	g.DeleteAt(0)
	assert.Equal(t, 0, g.Len())
	assert.Equal(t, 0, g.startPieceIndex)
	assert.Equal(t, 3, g.endPieceIndex)
	assert.Equal(t, []rune{r(97), r(0), r(0), r(97)}, g.array)
}

func TestGapTable_DeleteAt_NoGap(t *testing.T) {
	g := NewGapTable(4)
	g.AppendRune(r(97))
	g.AppendRune(r(98))
	g.AppendRune(r(99))

	g.DeleteAt(2)
	assert.Equal(t, 2, g.Len())
	assert.Equal(t, 2, g.startPieceIndex)
	assert.Equal(t, 3, g.endPieceIndex)
	assert.Equal(t, []rune{r(97), r(98), r(99), r(0)}, g.array)

	g.DeleteAt(1)
	assert.Equal(t, 1, g.Len())
	assert.Equal(t, 1, g.startPieceIndex)
	assert.Equal(t, 3, g.endPieceIndex)
	assert.Equal(t, []rune{r(97), r(98), r(99), r(0)}, g.array)

	g.DeleteAt(0)
	assert.Equal(t, 0, g.Len())
	assert.Equal(t, 0, g.startPieceIndex)
	assert.Equal(t, 3, g.endPieceIndex)
	assert.Equal(t, []rune{r(97), r(98), r(99), r(0)}, g.array)
}

