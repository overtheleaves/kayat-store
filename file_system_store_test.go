package store

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"os"
	"fmt"
)

var path = "fs_test"

func TestMain(m *testing.M) {
	retCode := m.Run()
	RemoveFileSystemStore(path)
	os.Exit(retCode)
}

func TestFileSystemStore_CreateFile(t *testing.T) {
	filename := "TestFileSystemStore_CreateFile"
	f := NewFileSystemStore(path)
	f.CreateFile(filename)
	assert.True(t, f.IsFileExist(filename))
}

func TestFileSystemStore_FileIter(t *testing.T) {

	filename := "TestFileSystemStore_FileIter"
	f := NewFileSystemStore(path)
	f.CreateFile(fmt.Sprintf("%s%d", filename, 1))
	f.CreateFile(fmt.Sprintf("%s%d", filename, 2))
	f.CreateFile(fmt.Sprintf("%s%d", filename, 3))

	res := map[string]bool{}

	for i := range f.FileIter() {
		res[i.Name()] = true
	}

	assert.True(t, res[fmt.Sprintf("%s%d", filename, 1)])
	assert.True(t, res[fmt.Sprintf("%s%d", filename, 2)])
	assert.True(t, res[fmt.Sprintf("%s%d", filename, 3)])
}

func TestFileSystemStore_WriteRead(t *testing.T) {
	filename := "TestFileSystemStore_WriteRead"
	f := NewFileSystemStore(path)
	f.CreateFile(filename)

	assert.Nil(t, f.Write(filename, []byte("test"), 0))

	res := make([]byte, 4)
	assert.Nil(t, f.Read(filename, res, 0))
	
	assert.Equal(t, "test", string(res))
}

func TestFileSystemStore_Clear(t *testing.T) {
	filename := "TestFileSystemStore_Clear"
	f := NewFileSystemStore(path)
	f.CreateFile(filename)
	assert.Nil(t, f.Write(filename, []byte("aaaaaaaaaa"), 0))
	assert.Nil(t, f.Clear(filename, 3, 5))

	res := make([]byte, 10)
	assert.Nil(t, f.Read(filename, res, 0))

	assert.Equal(t, "aaa\x00\x00\x00\x00\x00aa", string(res))
}

func TestFileSystemStore_ReadWhenFileNotExisted(t *testing.T) {
	filename := "TestFileSystemStore_ReadWhenFileNotExisted"
	f := NewFileSystemStore(path)

	res := make([]byte, 10)
	// error should be raised
	assert.NotNil(t, f.Read(filename, res, 0))
}


func TestFileSystemStore_WriteWhenFileNotExisted(t *testing.T) {
	filename := "TestFileSystemStore_WriteWhenFileNotExisted"
	f := NewFileSystemStore(path)

	data := make([]byte, 10)
	// error should be raised
	assert.NotNil(t, f.Write(filename, data, 0))
}

func TestFileSystemStore_ClearWhenFileNotExisted(t *testing.T) {
	filename := "TestFileSystemStore_ClearWhenFileNotExisted"
	f := NewFileSystemStore(path)

	// error should be raised
	assert.NotNil(t, f.Clear(filename, 0, 10))
}