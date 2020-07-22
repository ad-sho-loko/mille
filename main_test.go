package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func r(rn int) rune {
	return rune(rn)
}

func makeEmptyRow() *Row {
	var runes []rune
	return &Row{
		n:0,
		runes: runes,
	}
}

func makeAlphaRow() *Row {
	runes := []rune {
		97,
		98,
		99,
	}

	return &Row{
		n:3,
		runes: runes,
	}
}

// TODO: Add tests for MultiByte

func TestDeleteAt_Empty(t *testing.T) {
	row := makeEmptyRow()
	row.deleteAt(0)
	assert.Equal(t, 0, row.n)
}

func TestDeleteAt_ASCII(t *testing.T) {
	row := makeAlphaRow()

	row.deleteAt(0)
	assert.Equal(t, 2, row.n)
	assert.Equal(t, r(98), row.runes[0])
	assert.Equal(t, r(99), row.runes[1])

	row.deleteAt(1)
	assert.Equal(t, 1, row.n)
	assert.Equal(t, r(98), row.runes[0])

	row.deleteAt(0)
	assert.Equal(t, 0, row.n)
}


func TestInsertAt_Empty(t *testing.T) {
	row := makeEmptyRow()
	row.insertAt(0, r(100))
	assert.Equal(t, 1, row.n)
	assert.Equal(t, r(100), row.runes[0])
}

func TestInsertAt_ASCII(t *testing.T) {
	row := makeAlphaRow()
	row.insertAt(3, r(100))
	assert.Equal(t, 4, row.n)
	assert.Equal(t, r(100), row.runes[3])

	row.insertAt(100, r(101))
	assert.Equal(t, 5, row.n)
	assert.Equal(t, r(101), row.runes[4])
}
