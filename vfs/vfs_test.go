package vfs

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestVirtualFile_ReadWrite(t *testing.T) {
	filename := "TestVirtualFile_ReadWrite"
	file := newVirtualFile(filename)
	file.Write([]byte("test1234"))

	res := make([]byte, 8)
	file.Read(res)

	assert.Equal(t, "test1234", string(res))
}

func TestVirtualFile_ReadWriteAt(t *testing.T) {

	filename := "TestVirtualFile_ReadWriteAt"
	file := newVirtualFile(filename)

	file.Write([]byte("aaaaaaaaaaaaaaa"))
	file.WriteAt([]byte("123456789"), 9)

	res := make([]byte, 9)
	file.ReadAt(res, 9)

	assert.Equal(t, "123456789", string(res))
}

func TestPath_Iterator(t *testing.T) {
	p := NewPath("/test/path/iter/")
	iter := p.Iterator()

	var i = 0

	for iter.HasNext() {
		assert.Equal(t, p.paths[i], iter.Value())
		i++
	}
}

func TestFileNode_addFile(t *testing.T) {
	p1 := NewPath("test/path/add")
	p2 := NewPath("test/path/add2")
	iter1 := p1.Iterator()
	iter2 := p2.Iterator()
	n1 := &fileNode{file: newVirtualDirectory("/")}
	n2 := &fileNode{file: newVirtualDirectory("/")}
	f1 := newVirtualFile("add")
	f2 := newVirtualFile("add2")

	n1.addFile(p1, f1, 0)

	for iter1.HasNext() {
		n1 = n1.children[iter1.Value()]
		assert.NotNil(t, n1)
	}

	assert.Equal(t, n1.file, f1)

	n2.addFile(p2, f2, 0)

	for iter2.HasNext() {
		n2 = n2.children[iter2.Value()]
		assert.NotNil(t, n2)
	}

	assert.Equal(t, n2.file, f2)
}

func TestFileNode_getFile(t *testing.T) {
	p1 := NewPath("test/path/add")
	p2 := NewPath("test/path/add2")

	n := &fileNode{file: newVirtualDirectory("/")}
	f := newVirtualFile("add")
	n.addFile(p1, f, 0)

	assert.Equal(t, f, n.getFile(p1, 0))
	assert.Nil(t, n.getFile(p2, 0))
}