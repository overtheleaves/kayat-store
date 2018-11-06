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

func TestFileNode_addFile(t *testing.T) {
	p1 := NewPath("test/__dir_name_/add")
	p2 := NewPath("test/__dir_name_/add2")
	iter1 := p1.Iterator()
	iter2 := p2.Iterator()
	n1 := newFileNode(newVirtualDirectory("/"))
	n2 := newFileNode(newVirtualDirectory("/"))
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
	p1 := NewPath("test/__dir_name_/add")
	p2 := NewPath("test/__dir_name_/add2")

	n := newFileNode(newVirtualDirectory("/"))
	f := newVirtualFile("add")
	n.addFile(p1, f, 0)

	assert.Equal(t, f, n.getFile(p1, 0))
	assert.Nil(t, n.getFile(p2, 0))
}

func TestFileNode_removeFile(t *testing.T) {
	p1 := NewPath("test/__dir_name_/add1")
	n := newFileNode(newVirtualDirectory("/"))
	f := newVirtualFile("add1")
	n.addFile(p1, f, 0)

	assert.NotNil(t, n.removeFile(NewPath("test/__dir_name_/add2"), 0))
	assert.Nil(t, n.removeFile(p1, 0))
	assert.Nil(t, n.getFile(p1, 0))
}

func TestMemFileSystem_CustomDelimiter(t *testing.T) {
	fs, _ := NewMemoryFileSystemWithPathDelimiter(":mount_custon_delim", ":")
	fs.Context()
}