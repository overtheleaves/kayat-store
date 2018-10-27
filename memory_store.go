package store

import (
	"strings"
	"github.com/overtheleaves/kayat-store/vfs"
)

type memoryStore struct {
	path string
	root vfs.File
}

func NewMemoryStore(path string) Store {
	// check directory exists
	// if not, create
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	//
	//if !vfs.isDirectoryExist(path) {
	//	os.MkdirAll(path, os.ModePerm)
	//}
	//
	//fs := &fileSystemStore{path: path}
	//return fs
}
