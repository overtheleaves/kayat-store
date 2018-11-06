package vfs

import (
	"time"
	"fmt"
	"errors"
)

var (
	invalidOffsetErr         = errors.New("invalid offset error")
	illegalFileNameErr       = errors.New("illegal file name")
	noSuchFileOrDirectoryErr = errors.New("no such file or directory")
	alreadyMountedErr        = errors.New("filesystem is already mounted")
	fileExistsErr            = errors.New("file exists")
	invalidContextErr        = errors.New("invalid context")
	invalidMountOnPathErr    = errors.New("invalid mount path. mount __dir_name_ should be absolute __dir_name_")
	fileReadWriteErr         = errors.New("cannot open file to read/write")
	nestedMountedErr         = func(path string) error {
		return errors.New(fmt.Sprintf("mount path cannot be sub/parent directory of already mounted file system %s", path))
	}
)

type VirtualFileSystem interface {
	NewFile(context *Context, pathname string) 	(File, error)
	Remove(context *Context, pathname string) 	error
	OpenFile(context *Context, name string)	(File, error)
	Create(context *Context, name string)	(File, error)
	Mkdir(context *Context, pathname string) error
	FileExisted(context *Context, pathname string)	bool
	ChangeDirectory(context *Context, pathname string) error
	Context() *Context
	ListSegments(context *Context, pathname string) ([]FileStat, error)
	PresentWorkingDirectory(context *Context) string
	Type() string
}

type File interface {
	Stat() FileStat
	Read(b []byte) (n int, err error)
	ReadAt(b []byte, off int64) (n int, err error)
	Write(b []byte) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
	Delete()
}

type FileStat interface {
	Name() string
	Size() int64
	ModTime() time.Time
	IsDir() bool
	Immutable() FileStat
}