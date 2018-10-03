package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResults(t *testing.T) {
	r := NewResults()
	assert.NotNil(t, r)
	assert.NotEmpty(t, r)
}

func TestResults_Add(t *testing.T) {
	r := NewResults()
	assert.NotNil(t, r)
	assert.NotEmpty(t, r)

	s := r.Add("path", "somefile.txt")
	assert.EqualValues(t, r, s)

	tt := r.Add("url", "example.net")
	assert.EqualValues(t, r, tt)
}

func TestResults_Merge(t *testing.T) {
	r1 := &Results{
		Paths: map[string]bool{"foobar.txt": true},
	}

	r2 := &Results{
		Paths: map[string]bool{"bar.doc": true},
	}

	mm := &Results{
		Paths: map[string]bool{
			"bar.doc":    true,
			"foobar.txt": true,
		},
	}

	tt := r1.Merge(r2)
	assert.EqualValues(t, r1, mm)
	assert.EqualValues(t, r1, tt)
}

func TestResults_Merge2(t *testing.T) {
	r1 := &Results{
		Paths: map[string]bool{"foobar.txt": true},
		URLs:  map[string]bool{"example.com": true},
	}

	r2 := &Results{
		Paths: map[string]bool{"bar.doc": true},
		URLs:  map[string]bool{"example.net": true},
	}

	mm := &Results{
		Paths: map[string]bool{
			"bar.doc":    true,
			"foobar.txt": true,
		},
		URLs: map[string]bool{
			"example.com": true,
			"example.net": true,
		},
	}

	tt := r1.Merge(r2)
	assert.EqualValues(t, r1, mm)
	assert.EqualValues(t, r1, tt)
}
