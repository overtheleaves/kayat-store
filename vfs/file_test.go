package vfs

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestWrapperFileSystem_FileExisted(t *testing.T) {
	fs, _ := NewWrapperFileSystem(__dir_name_ + "/mount_fileexisted")
	context := fs.Context()
	fs.NewFile(context, "test/path/file")

	assert.True(t, fs.FileExisted(context, "test/path/file"))
	assert.True(t, fs.FileExisted(context, "test/path"))
	assert.True(t, fs.FileExisted(context, "test"))
	assert.False(t, fs.FileExisted(context, "test/path/file2"))
}

func TestWrapperFileSystem_ChangeDirectory(t *testing.T) {
	fs, _ := NewWrapperFileSystem(__dir_name_ + "/mount_cd")
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
	fs, _ := NewWrapperFileSystem(__dir_name_ + "/mount_mkdir")
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