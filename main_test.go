package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func r(rn int) rune {
	return rune(rn)
}

func makeEmptyRow() *Row {
	return &Row{
		chars: NewGapTable(128),
	}
}

func makeAlphaRow() *Row {
	gt := NewGapTable(128)
	gt.AppendRune(97)
	gt.AppendRune(98)
	gt.AppendRune(99)

	return &Row{
		chars: gt,
	}
}

func TestDeleteAt_Empty(t *testing.T) {
	row := makeEmptyRow()
	row.deleteAt(0)
	assert.Equal(t, 0, row.len())
}

func TestDeleteAt_ASCII(t *testing.T) {
	row := makeAlphaRow()

	row.deleteAt(0)
	assert.Equal(t, 2, row.len())
	assert.Equal(t, r(98), row.chars.At(0))
	assert.Equal(t, r(99), row.chars.At(1))

	row.deleteAt(1)
	assert.Equal(t, 1, row.len())
	assert.Equal(t, r(98), row.chars.At(0))

	row.deleteAt(0)
	assert.Equal(t, 0, row.len())
}

func TestInsertAt_Empty(t *testing.T) {
	row := makeEmptyRow()
	row.insertAt(0, r(100))
	assert.Equal(t, 1, row.len())
	assert.Equal(t, r(100), row.chars.At(0))
}

func TestInsertAt_ASCII(t *testing.T) {
	row := makeAlphaRow()
	row.insertAt(3, r(100))
	assert.Equal(t, 4, row.len())
	assert.Equal(t, r(100), row.chars.At(3))

	row.insertAt(100, r(101))
	assert.Equal(t, 5, row.len())
	assert.Equal(t, r(101), row.chars.At(4))
}
