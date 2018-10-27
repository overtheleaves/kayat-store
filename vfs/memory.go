package vfs

import (
	"time"
	"sync"
	"fmt"
)

var (
	root = newVirtualDirectory("_")
)

type virtualFile struct {
	mu sync.Mutex
	children []File
	data     []byte
	stat 	FileStat
}

type memFileStat struct {
	name string
	size int64
	modTime time.Time
	isDir 	bool
}

func (m *memFileStat) Name() string {
	return m.name
}

func (m *memFileStat) Size() int64 {
	return m.size
}

func (m *memFileStat) ModTime() time.Time {
	return m.modTime
}

func (m *memFileStat) IsDir() bool {
	return m.isDir
}

func newVirtualDirectory(name string) File {
	return &virtualFile{
		stat: &memFileStat{
			name: name,
			size: 0,
			modTime: time.Now(),
			isDir: true,
		},
	}
}

func newVirtualFile(name string) File {
	return &virtualFile{
		stat: &memFileStat{
			name: name,
			size: 0,
			modTime: time.Now(),
			isDir: false,
		},
	}
}

func (f *virtualFile) Stat() FileStat {
	return f.stat
}

func (f *virtualFile) Read(b []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(b) < len(f.data) {
		n = len(b)
	} else {
		n = len(f.data)
	}

	copy(b, f.data)
	return n, nil
}

func (f *virtualFile) ReadAt(b []byte, off int64) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.stat.Size() <= off {
		// invalid offset
		return 0, fmt.Errorf("invalid offset = %d, file size = %d", off, f.stat.Size())
	}

	if int64(len(b)) < f.stat.Size() - off {
		n = len(b)
	} else {
		n = len(f.data) - int(off)
	}

	copy(b, f.data[off:])
	return n, nil
}

func (f *virtualFile) Write(b []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	n = len(b)

	if f.stat.Size() <= int64(len(b)) {
		f.data = make([]byte, len(b))
	}

	copy(f.data, b)

	return n, nil
}

func (f *virtualFile) WriteAt(b []byte, off int64) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	n = len(b)
	if f.stat.Size() < off + int64(len(b)) {
		original := f.data
		f.data = make([]byte, len(b) + int(off) + 1)
		copy(f.data, original)
	}

	copy(f.data[off:], b)
	return n, nil
}