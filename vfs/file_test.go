package vfs

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"os"
)

var path = ""

func TestMain(m *testing.M) {
	path, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	retCode := m.Run()
	os.Exit(retCode)
}

func TestWrapperFileSystem_NewFile(t *testing.T) {
	fs, _ := NewWrapperFileSystem(path + "/mount_newfile")
	context := fs.Context()
	f, err := fs.NewFile(context, "test/path/newfile")
	assert.NotNil(t, f)
	assert.Nil(t, err)

	fs.ChangeDirectory(context, "/test")	// cd /test
	f1, err1 := fs.NewFile(context, "path/file2")	// file create /test/path/file2
	assert.NotNil(t, f1)
	assert.Nil(t, err1)
}

func TestWrapperFileSystem_Remove(t *testing.T) {
	fs, _ := NewWrapperFileSystem(path + "/mount_remove")
	context := fs.Context()
	fs.NewFile(context, "test/path/file")

	assert.Nil(t, fs.Remove(context, "test/path/file"))
	assert.NotNil(t, fs.Remove(context, "test/path/file")) // no such file or directory err
	assert.NotNil(t, fs.Remove(context, "test/path/file22")) // no such file or directory err

	fs.NewFile(context, "text/path/file")
	assert.Nil(t, fs.Remove(context, "test"))
	assert.NotNil(t, fs.Remove(context, "test/path/file"))
	assert.NotNil(t, fs.Remove(context, "test/path"))
}

func TestWrapperFileSystem_FileExisted(t *testing.T) {
	fs, _ := NewWrapperFileSystem(path + "/mount_fileexisted")
	context := fs.Context()
	fs.NewFile(context, "test/path/file")

	assert.True(t, fs.FileExisted(context, "test/path/file"))
	assert.True(t, fs.FileExisted(context, "test/path"))
	assert.True(t, fs.FileExisted(context, "test"))
	assert.False(t, fs.FileExisted(context, "test/path/file2"))
}

func TestWrapperFileSystem_ChangeDirectory(t *testing.T) {
	fs, _ := NewWrapperFileSystem(path + "/mount_cd")
	context := fs.Context()
	fs.NewFile(context, "test/path/file")

	assert.Equal(t, "/", fs.(*wrapperFileSystem).PresentWorkingDirectory(context).String())

	assert.Nil(t, fs.ChangeDirectory(context, "test"))
	assert.Equal(t, "/test", fs.(*wrapperFileSystem).PresentWorkingDirectory(context).String())

	assert.Nil(t, fs.ChangeDirectory(context, "path"))
	assert.Equal(t, "/test/path", fs.(*wrapperFileSystem).PresentWorkingDirectory(context).String())

	assert.Nil(t, fs.ChangeDirectory(context, "/"))
	assert.Equal(t, "/", fs.(*wrapperFileSystem).PresentWorkingDirectory(context).String())
}

func TestWrapperFileSystem_Mkdir(t *testing.T) {
	fs, _ := NewWrapperFileSystem(path + "/mount_mkdir")
	context := fs.Context()

	assert.Nil(t, fs.Mkdir(context, "test/path/"))
	assert.True(t, fs.FileExisted(context, "test/path"))

	assert.Nil(t, fs.Mkdir(context, "test2"))
	assert.Nil(t, fs.ChangeDirectory(context, "test2"))

	assert.Equal(t, "/test2", fs.(*wrapperFileSystem).PresentWorkingDirectory(context).String())

	assert.Nil(t, fs.Mkdir(context, "path/dir"))
	assert.True(t, fs.FileExisted(context, "path/dir"))

	fs.ChangeDirectory(context, "/")
	assert.True(t, fs.FileExisted(context, "test2/path/dir"))
}