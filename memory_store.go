package store

import (
	"github.com/overtheleaves/kayat-store/vfs"
)

type memoryStore struct {
	path string
	root vfs.File
}
