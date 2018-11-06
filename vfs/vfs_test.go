package vfs

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"reflect"
	"path/filepath"
	"os"
)

var __dir_name_ = ""

func TestMain(m *testing.M) {
	__dir_name_, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	retCode := m.Run()
	os.Exit(retCode)
}

/**
 Get virtual file systems to be tested
 */
func GetVirtualFileSystems(mountOnPath string) ([]VirtualFileSystem, []error) {

	res := make([]VirtualFileSystem, 0)
	errs := make([]error, 0)

	mfs, merr := NewMemoryFileSystem(mountOnPath)
	res = append(res, mfs)
	errs = append(errs, merr)

	ffs, ferr := NewWrapperFileSystem(mountOnPath)
	res = append(res, ffs)
	errs = append(errs, ferr)

	return res, errs
}

/*
 Assert Helper function
 */
func assertApplyAll(t *testing.T, object interface{},
	f func(t assert.TestingT, object interface{}, msgAndArgs ...interface{}) bool) {

	objValue := reflect.ValueOf(object)

	switch objValue.Kind() {
		case reflect.Array, reflect.Map, reflect.Slice:
			for i := 0 ; i < objValue.Len() ; i++ {
				f(t, objValue.Index(i).Interface())
			}
			break
	}
}

func TestNewVirtualFileSystems(t *testing.T) {
	fs1, err1 := GetVirtualFileSystems(__dir_name_ + "/root/mount")

	assertApplyAll(t, fs1, assert.NotNil)
	assertApplyAll(t, err1, assert.Nil)

	fs2, err2 := GetVirtualFileSystems(__dir_name_ + "/root/mount")
	assertApplyAll(t, fs2, assert.Nil)
	assertApplyAll(t, err2, assert.NotNil) // file exists error

	fs3, err3 := GetVirtualFileSystems("mount")
	assertApplyAll(t, fs3, assert.Nil)
	assertApplyAll(t, err3, assert.NotNil) // invalid path

	fs4, err4 := GetVirtualFileSystems(__dir_name_+ "/root/mount/sub")
	assertApplyAll(t, fs4, assert.Nil)
	assertApplyAll(t, err4, assert.NotNil) // nested err

	fs5, err5 := GetVirtualFileSystems(__dir_name_ + "/root")
	assertApplyAll(t, fs5, assert.Nil)
	assertApplyAll(t, err5, assert.NotNil) // nested err
}

func TestVirtualFileSystems_NewFile(t *testing.T) {

	vfs, errs := GetVirtualFileSystems(__dir_name_ + "/mount_newfile")
	assertApplyAll(t, vfs, assert.NotNil)
	assertApplyAll(t, errs, assert.Nil)

	for _, fs := range vfs {
		context := fs.Context()

		f, err := fs.NewFile(context, "/test/path/newfile")
		assert.NotNil(t, f)
		assert.Nil(t, err)

		of, oerr := fs.OpenFile(context, "test/path/newfile")
		assert.NotNil(t, of)
		assert.Nil(t, oerr)
	}
}

func TestVirtualFileSystems_Remove(t *testing.T) {

	vfs, errs := GetVirtualFileSystems(__dir_name_ + "/mount_remove")
	assertApplyAll(t, vfs, assert.NotNil)
	assertApplyAll(t, errs, assert.Nil)

	for _, fs := range vfs {
		context := fs.Context()

		fs.NewFile(context, "test/path/file")

		assert.Nil(t, fs.Remove(context, "test/path/file"))
		assert.NotNil(t, fs.Remove(context, "test/path/file")) // no such file or directory err
		assert.NotNil(t, fs.Remove(context, "test/path/file22")) // no such file or directory err

		assert.Nil(t, fs.Remove(context, "test"))	// remove test dir
		assert.NotNil(t, fs.Remove(context, "test/path/file"))  // no such file or directory err
		assert.NotNil(t, fs.Remove(context, "test/path")) // no such file or directory err
	}
}

