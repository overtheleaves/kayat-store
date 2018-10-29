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
	p1 := NewPath("test/path/add")
	p2 := NewPath("test/path/add2")

	n := newFileNode(newVirtualDirectory("/"))
	f := newVirtualFile("add")
	n.addFile(p1, f, 0)

	assert.Equal(t, f, n.getFile(p1, 0))
	assert.Nil(t, n.getFile(p2, 0))
}

func TestFileNode_removeFile(t *testing.T) {
	p1 := NewPath("test/path/add1")
	n := newFileNode(newVirtualDirectory("/"))
	f := newVirtualFile("add1")
	n.addFile(p1, f, 0)

	assert.NotNil(t, n.removeFile(NewPath("test/path/add2"), 0))
	assert.Nil(t, n.removeFile(p1, 0))
	assert.Nil(t, n.getFile(p1, 0))
}

func TestNewMemoryFileSystem(t *testing.T) {
	fs1, err1 := NewMemoryFileSystem("/mount")
	assert.Nil(t, err1)
	assert.NotNil(t, fs1)

	fs2, err2 := NewMemoryFileSystem("/mount")
	assert.Nil(t, fs2)
	assert.NotNil(t, err2)	// file exists error

	fs3, err3 := NewMemoryFileSystem("mount")
	assert.Nil(t, fs3)
	assert.NotNil(t, err3)	// invalid path error
}

func TestMemFileSystem_NewFile(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount_newfile")
	context := fs.Context()
	fs.NewFile(context, "/test/path/newfile")
	file, err := fs.OpenFile(context, "/test/path/newfile")
	assert.NotNil(t, file)
	assert.Nil(t, err)
}

func TestMemFileSystem_OpenFile(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount_openfile")
	context := fs.Context()
	fs.NewFile(context, "/test/path/openfile")
	file1, err1 := fs.OpenFile(context, "/test/path/openfile")
	file2, err2 := fs.OpenFile(context, "/test/path/no-such-file")

	assert.NotNil(t, file1)
	assert.Nil(t, err1)

	assert.Nil(t, file2)
	assert.NotNil(t, err2)	// no such file error
}

func TestMemFileSystem_Remove(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount_remove")
	context := fs.Context()
	fs.NewFile(context, "/test/path/removefile")

	err1 := fs.Remove(context, "/test/path/no-such-file")
	err2 := fs.Remove(context, "/test/path/removefile")

	assert.NotNil(t, err1)	// no such file error
	assert.Nil(t, err2)

	assert.False(t, fs.FileExisted(context, "/test/path/removefile"))

	file, _ := fs.OpenFile(context, "/test/path/removefile")
	assert.Nil(t, file)
}

func TestMemFileSystem_RemoveAll(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount_removeall")
	context := fs.Context()

	fs.NewFile(context, "/test/path/removefile1")
	fs.NewFile(context, "/test/path/removefile2")

	err := fs.Remove(context, "/test/path")

	assert.Nil(t, err)
	assert.False(t, fs.FileExisted(context, "/test/path/removefile1"))
	assert.False(t, fs.FileExisted(context, "/test/path/removefile2"))
	assert.False(t, fs.FileExisted(context, "/test/path/"))
	assert.True(t, fs.FileExisted(context, "/test/"))
}

func TestMemFileSystem_Mkdir(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount1")
	context := fs.Context()

	assert.Nil(t, fs.Mkdir(context, "test/mkdirall/path"))
	assert.NotNil(t, fs.Mkdir(context, "test/mkdirall/path"))	// no such file error
	assert.True(t, fs.FileExisted(context, "test/mkdirall/path"))
}

func TestMemFileSystem_CustomDelimiter(t *testing.T) {
	fs, _ := NewMemoryFileSystemWithPathDelimiter(":mount_custon_delim", ":")
	fs.Context()
}

func TestMemFileSystem_ChangeDirectory(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount_cd")
	context := fs.Context()

	fs.NewFile(context, "test/path/cd/cd")
	fs.ChangeDirectory(context, "test/path")

	assert.True(t, fs.FileExisted(context, "cd/cd"))
}

func TestMemFileSystem_ListSegments(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount_ls")
	context := fs.Context()

	fs.Mkdir(context, "test")
	fs.Mkdir(context, "test/path")
	fs.Mkdir(context, "test1")
	fs.Mkdir(context, "test2")

	expected := make(map[string]bool)
	expected["test"] = true
	expected["test1"] = true
	expected["test2"] = true

	result1, err1 := fs.ListSegments(context, "")

	assert.Nil(t, err1)
	assert.Equal(t, 3, len(expected))
	for _, stat := range result1 {
		assert.True(t, expected[stat.Name()])
		delete(expected, stat.Name())
	}

	result2, err2 := fs.ListSegments(context, "test3")
	assert.NotNil(t, err2)	// no such file error
	assert.Nil(t, result2)
}