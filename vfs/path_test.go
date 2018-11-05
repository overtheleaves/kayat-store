package vfs

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPath_Iterator(t *testing.T) {
	p := NewPath("/test/path/iter/")
	iter := p.Iterator()

	var i = 0

	for iter.HasNext() {
		assert.Equal(t, p.paths[i], iter.Value())
		i++
	}
}
