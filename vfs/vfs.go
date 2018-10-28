package vfs

import "time"

type VirtualFileSystem interface {
	NewFile(name string) 	(File, error)
	Remove(name string) 	error
	MkdirAll(path string) 	error
	RemoveAll(path string)	error
	OpenFile(name string)	(File, error)
	Create(name string)	(File, error)
	Mkdir(name string) error
	FileExisted(name string)	bool
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
}