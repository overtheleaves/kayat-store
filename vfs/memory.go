package vfs

import (
	"time"
	"sync"
	"fmt"
	"strings"
	"strconv"
)

var (
	invalidOffsetErr = fmt.Errorf("invalid offset error")
	illegalFileNameErr = fmt.Errorf("illegal file name")
	noSuchFileOrDirectoryErr = fmt.Errorf("no such file or directory")
	alreadyMountedErr = fmt.Errorf("filesystem is already mounted")
	fileExistsErr = fmt.Errorf("File exists")

	memFileSystems = make(map[string]*memFileSystem)
)

type fileNode struct {
	file 	File
	children map[string]*fileNode
}

type virtualFile struct {
	mu sync.RWMutex
	children []File
	data     []byte
	stat 	*memFileStat
}

type memFileStat struct {
	name string
	size int64
	modTime time.Time
	isDir 	bool
}

type memFileSystem struct {
	mount 	*Path
	rootNode *fileNode
	mu sync.Mutex
}

type MemFileSystemError struct {
	Err error
	Op string
	Path string
}

type MemFileError struct {
	Err error
	Op string
	Offset int64
}

func (e *MemFileSystemError) Error() string {
	return e.Op + ": " + e.Path + ": " + e.Err.Error()
}

func (e *MemFileError) Error() string {
	return e.Op + ": " + e.Err.Error() + ": (offset: " + strconv.FormatInt(e.Offset, 10) + ")"
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

func (n *fileNode) addFile(path *Path, file File, i int) {

	if i > path.Len() - 1 {
		return
	}

	dir := path.NthPath(i)
	if n.children[dir] == nil {
		n.children[dir] = &fileNode{}
	}

	if i == path.Len() - 1 {
		n.children[dir].file = file
	} else {
		n.children[dir].file = newVirtualDirectory(dir)
		n.children[dir].addFile(path, file, i+1)
	}
}

func (n *fileNode) getFile(path *Path, i int) File {
	if i > path.Len() - 1 {
		return n.file
	}

	dir := path.NthPath(i)
	if n.children[dir] == nil {
		return nil
	} else {
		return n.children[dir].getFile(path, i+1)
	}
}

func NewMemoryFileSystem(mountOnPath string) (VirtualFileSystem, error) {

	if strings.HasSuffix(mountOnPath, "/") {
		mountOnPath = mountOnPath[:len(mountOnPath) - 1]
	}

	if memFileSystems[mountOnPath] != nil {
		return nil, &MemFileSystemError{Err: alreadyMountedErr, Op: "mount", Path: mountOnPath}
	}

	mfs := &memFileSystem{
		mount: NewPath(mountOnPath),
		rootNode: &fileNode{
			file: newVirtualDirectory("/"),
		},
	}

	memFileSystems[mountOnPath] = mfs

	return mfs, nil
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
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(b) < len(f.data) {
		n = len(b)
	} else {
		n = len(f.data)
	}

	copy(b, f.data)
	return n, nil
}

func (f *virtualFile) ReadAt(b []byte, off int64) (n int, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.stat.Size() <= off {
		// invalid offset
		return 0, &MemFileError{Err: invalidOffsetErr, Op: "ReadAt", Offset: off}
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
	f.stat.size = int64(len(f.data))

	return n, nil
}

func (f *virtualFile) WriteAt(b []byte, off int64) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	n = len(b)
	// expand data and copy original data into new data pool
	if f.stat.Size() < off + int64(len(b)) {
		original := f.data
		f.data = make([]byte, len(b) + int(off) + 1)
		copy(f.data, original)
	}

	copy(f.data[off:], b)
	f.stat.size = int64(len(f.data))

	return n, nil
}

func (fs *memFileSystem) NewFile(name string) (File, error) {
	path := NewPath(name)
	filename := path.FileName()

	if filename == "" {
		return nil, &MemFileSystemError{Err: illegalFileNameErr, Op: "NewFile", Path: name}
	}

	var err error
	fs.mu.Lock()

	file := fs.rootNode.getFile(path, 0)
	if file == nil {
		// create new file
		// if file is already existed (file != nil), then just return the file
		file = newVirtualFile(filename)
		fs.rootNode.addFile(path, file, 0)
	} else {
		err = &MemFileSystemError{Err: fileExistsErr, Op: "NewFile", Path: name}
	}
	fs.mu.Unlock()

	return file, err
}

func (fs *memFileSystem) FileExisted(name string) bool {
	return fs.rootNode.getFile(NewPath(name), 0) != nil
}

func (fs *memFileSystem) Remove(name string) error {
	return nil
}

func (fs *memFileSystem) MkdirAll(path string) error {
	return nil
}

func (fs *memFileSystem) RemoveAll(path string)	error {
	return nil
}

func (fs *memFileSystem) OpenFile(name string) (File, error) {
	return nil, nil
}

func (fs *memFileSystem) Create(name string) (File, error) {
	return fs.NewFile(name)
}

func (fs *memFileSystem) Mkdir(name string) error {
	return nil
}