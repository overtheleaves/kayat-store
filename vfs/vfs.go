package vfs

import "time"

type VirtualFileSystem interface {
	NewFile(context *Context, pathname string) 	(File, error)
	Remove(context *Context, pathname string) 	error
	OpenFile(context *Context, name string)	(File, error)
	Create(context *Context, name string)	(File, error)
	Mkdir(context *Context, pathname string) error
	FileExisted(context *Context, name string)	bool
	ChangeDirectory(context *Context, pathname string) error
	Context() *Context
	ListSegments(context *Context, pathname string) ([]FileStat, error)
}

type File interface {
	Stat() FileStat
	Read(b []byte) (n int, err error)
	ReadAt(b []byte, off int64) (n int, err error)
	Write(b []byte) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
}

type FileStat interface {
	Name() string
	Size() int64
	ModTime() time.Time
	IsDir() bool
	Immutable() FileStat
}