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

func TestNewMemoryFileSystem(t *testing.T) {
	fs1, err1 := NewMemoryFileSystem("/root/mount")
	assert.Nil(t, err1)
	assert.NotNil(t, fs1)

	fs2, err2 := NewMemoryFileSystem("/root/mount")
	assert.Nil(t, fs2)
	assert.NotNil(t, err2)	// file exists error

	fs3, err3 := NewMemoryFileSystem("mount")
	assert.Nil(t, fs3)
	assert.NotNil(t, err3)	// invalid __dir_name_ error

	fs4, err4 := NewMemoryFileSystem("/root/mount/sub")
	assert.Nil(t, fs4)
	assert.NotNil(t, err4)	// overlapped err

	fs5, err5 := NewMemoryFileSystem("/root")
	assert.Nil(t, fs5)
	assert.NotNil(t, err5)	// overlapped err
}

func TestMemFileSystem_OpenFile(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount_openfile")
	context := fs.Context()
	fs.NewFile(context, "/test/__dir_name_/openfile")
	file1, err1 := fs.OpenFile(context, "/test/__dir_name_/openfile")
	file2, err2 := fs.OpenFile(context, "/test/__dir_name_/no-such-file")

	assert.NotNil(t, file1)
	assert.Nil(t, err1)

	assert.Nil(t, file2)
	assert.NotNil(t, err2)	// no such file error
}

func TestMemFileSystem_RemoveAll(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount_removeall")
	context := fs.Context()

	f1, _ := fs.NewFile(context, "/test/__dir_name_/removefile1")
	f2, _ := fs.NewFile(context, "/test/__dir_name_/removefile2")

	err := fs.Remove(context, "/test/__dir_name_")

	assert.Nil(t, err)
	assert.False(t, fs.FileExisted(context, "/test/__dir_name_/removefile1"))
	assert.False(t, fs.FileExisted(context, "/test/__dir_name_/removefile2"))
	assert.False(t, fs.FileExisted(context, "/test/__dir_name_/"))
	assert.True(t, fs.FileExisted(context, "/test/"))

	_, ferr1 := f1.Write([]byte("aaaa"))
	_, ferr2 := f2.WriteAt([]byte("aaaa"), 0)

	assert.NotNil(t, ferr1)	// cannot open file to read/write error
	assert.NotNil(t, ferr2)	// cannot open file to read/write error
}

func TestMemFileSystem_Mkdir(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount1")
	context := fs.Context()

	assert.Nil(t, fs.Mkdir(context, "test/mkdirall/__dir_name_"))
	assert.NotNil(t, fs.Mkdir(context, "test/mkdirall/__dir_name_"))	// no such file error
	assert.True(t, fs.FileExisted(context, "test/mkdirall/__dir_name_"))
}

func TestMemFileSystem_CustomDelimiter(t *testing.T) {
	fs, _ := NewMemoryFileSystemWithPathDelimiter(":mount_custon_delim", ":")
	fs.Context()
}

func TestMemFileSystem_ChangeDirectory(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount_cd")
	context := fs.Context()

	fs.NewFile(context, "test/__dir_name_/cd/cd")
	fs.ChangeDirectory(context, "test/__dir_name_")

	assert.True(t, fs.FileExisted(context, "cd/cd"))
}

func TestMemFileSystem_ListSegments(t *testing.T) {
	fs, _ := NewMemoryFileSystem("/mount_ls")
	context := fs.Context()

	fs.Mkdir(context, "test")
	fs.Mkdir(context, "test/__dir_name_")
	fs.Mkdir(context, "test1")
	fs.Mkdir(context, "test2")

	expected := make(map[string]bool)
	expected["test"] = true
	expected["test1"] = true
	expected["test2"] = true

	result1, err1 := fs.ListSegments(context, "/")

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