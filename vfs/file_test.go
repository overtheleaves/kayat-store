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
	fs, _ := NewWrapperFileSystem(path + "/mount")
	context := fs.Context()
	f, err := fs.NewFile(context, "test/path/newfile")
	assert.NotNil(t, f)
	assert.Nil(t, err)
}
