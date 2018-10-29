package vfs

import (
	"time"
	"sync"
	"strings"
	"strconv"
	"errors"
)

var (
	invalidOffsetErr = errors.New("invalid offset error")
	illegalFileNameErr = errors.New("illegal file name")
	noSuchFileOrDirectoryErr = errors.New("no such file or directory")
	alreadyMountedErr = errors.New("filesystem is already mounted")
	fileExistsErr = errors.New("file exists")
	invalidContextErr = errors.New("invalid context")
	invalidMountOnPathErr = errors.New("invalid mount path. mount path should be absolute path")

	memFileSystems = make(map[string]*memFileSystem)
)

type fileNode struct {
	file 	File
	children map[string]*fileNode
}

type virtualFile struct {
	mu sync.RWMutex
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
	mu sync.RWMutex
	pwd map[*Context]*fileNode
	pathDelimiter string
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

type Context struct {

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

func (m *memFileStat) Immutable() FileStat {
	return &memFileStat{
		name: m.name,
		size: m.size,
		modTime: m.modTime,
		isDir: m.isDir,
	}
}

func (n *fileNode) addFile(path *Path, file File, i int) {

	if i > path.Len() - 1 {
		return
	}

	dir := path.NthPath(i)
	if n.children[dir] == nil {
		n.children[dir] = newFileNode(nil)
	}

	if i == path.Len() - 1 {
		n.children[dir].file = file
	} else {
		n.children[dir].file = newVirtualDirectory(dir)
		n.children[dir].addFile(path, file, i+1)
	}
}

func (n *fileNode) addDirectory(path *Path, i int) {
	if i > path.Len() - 1 {
		return
	}

	dir := path.NthPath(i)
	if n.children[dir] == nil {
		n.children[dir] = newFileNode(newVirtualDirectory(dir))
	}

	n.children[dir].addDirectory(path, i+1)
}

func (n *fileNode) getFile(path *Path, i int) File {
	node := n.getFileNode(path, i)
	if node != nil {
		return node.file
	}
	return nil
}

func (n *fileNode) getFileNode(path *Path, i int) *fileNode {
	if i > path.Len() - 1 {
		return n
	}

	dir := path.NthPath(i)
	if n.children[dir] == nil {
		return nil
	} else {
		return n.children[dir].getFileNode(path, i+1)
	}
}

func (n *fileNode) removeFile(path *Path, i int) error {
	if i > path.Len() - 1 || i < 0 {
		// invalid index i
		return noSuchFileOrDirectoryErr
	}

	dir := path.NthPath(i)
	if n.children[dir] == nil {
		// no such file or directory
		return noSuchFileOrDirectoryErr
	} else if i == path.Len() - 1 {
		// remove target
		n := n.children[dir]
		n.removeAllFiles()
		delete(n.children, dir)
		return nil
	} else {
		return n.children[dir].removeFile(path, i+1)
	}
}

func (n *fileNode) removeAllFiles() {
	for _, child := range n.children {
		child.removeAllFiles()
	}

	for dir := range n.children {
		delete(n.children, dir)
	}

	n.children = nil
	n.file = nil
}

func newFileNode(file File) *fileNode {
	return &fileNode{
		file: file,
		children: make(map[string]*fileNode),
	}
}

func NewMemoryFileSystem(mountOnPath string) (VirtualFileSystem, error) {
	return NewMemoryFileSystemWithPathDelimiter(mountOnPath, DEFAULT_PATH_DELIMITER)
}

func NewMemoryFileSystemWithPathDelimiter(mountOnPath string, delimiter string) (VirtualFileSystem, error) {
	if !strings.HasPrefix(mountOnPath, delimiter) {
		return nil, &MemFileSystemError{Err: invalidMountOnPathErr, Op: "mount", Path: mountOnPath}
	}

	if strings.HasSuffix(mountOnPath, delimiter) {
		mountOnPath = mountOnPath[:len(mountOnPath) - 1]
	}

	if memFileSystems[mountOnPath] != nil {
		return nil, &MemFileSystemError{Err: alreadyMountedErr, Op: "mount", Path: mountOnPath}
	}

	mfs := &memFileSystem{
		mount: NewPathWithDelimiter(mountOnPath, delimiter),
		rootNode: newFileNode(newVirtualDirectory(delimiter)),
		pathDelimiter: delimiter,
		pwd: make(map[*Context]*fileNode),
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

func (fs *memFileSystem) NewFile(context *Context, pathname string) (File, error) {
	path := NewPathWithDelimiter(pathname, fs.pathDelimiter)
	filename := path.FileName()

	if filename == "" {
		return nil, &MemFileSystemError{Err: illegalFileNameErr, Op: "NewFile", Path: pathname}
	}

	var err error
	fs.mu.Lock()

	wd := fs.workingDirectoryNode(context, pathname)
	file := wd.getFile(path, 0)
	if file == nil {
		// create new file
		// if file is already existed (file != nil), then just return the file
		file = newVirtualFile(filename)
		fs.rootNode.addFile(path, file, 0)
	} else {
		err = &MemFileSystemError{Err: fileExistsErr, Op: "NewFile", Path: pathname}
	}
	fs.mu.Unlock()

	return file, err
}

func (fs *memFileSystem) FileExisted(context *Context, pathname string) bool {
	wd := fs.workingDirectoryNode(context, pathname)
	if wd == nil {
		return false
	} else {
		return wd.getFile(NewPathWithDelimiter(pathname, fs.pathDelimiter), 0) != nil
	}
}

func (fs *memFileSystem) Remove(context *Context, pathname string) error {
	wd := fs.workingDirectoryNode(context, pathname)
	return wd.removeFile(NewPathWithDelimiter(pathname, fs.pathDelimiter), 0)
}

func (fs *memFileSystem) OpenFile(context *Context, pathname string) (File, error) {
	wd := fs.workingDirectoryNode(context, pathname)
	file := wd.getFile(NewPathWithDelimiter(pathname, fs.pathDelimiter), 0)
	if file == nil {
		return nil, &MemFileSystemError{Err: noSuchFileOrDirectoryErr, Op: "OpenFile", Path: pathname}
	} else {
		return file, nil
	}
}

func (fs *memFileSystem) Create(context *Context, name string) (File, error) {
	return fs.NewFile(context, name)
}

func (fs *memFileSystem) Mkdir(context *Context, pathname string) error {
	path := NewPathWithDelimiter(pathname, fs.pathDelimiter)

	var err error
	fs.mu.Lock()

	wd := fs.workingDirectoryNode(context, pathname)
	file := wd.getFile(path, 0)
	if file == nil {
		// create new directory
		fs.rootNode.addDirectory(path, 0)
	} else {
		err = &MemFileSystemError{Err: fileExistsErr, Op: "MkdirAll", Path: pathname}
	}
	fs.mu.Unlock()

	return err
}

func (fs *memFileSystem) Context() *Context {
	context := &Context{}
	fs.pwd[context] = fs.rootNode
	return context
}

func (fs *memFileSystem) ChangeDirectory(context *Context, pathname string) error {
	wd := fs.workingDirectoryNode(context, pathname)
	if wd == nil {
		return &MemFileSystemError{Err: invalidContextErr, Op: "ChangeDirectory", Path: pathname}
	} else {
		n := wd.getFileNode(NewPathWithDelimiter(pathname, fs.pathDelimiter), 0)
		if n != nil {
			fs.pwd[context] = n
			return nil
		} else {
			return &MemFileSystemError{Err: noSuchFileOrDirectoryErr, Op: "ChangeDirectory", Path: pathname}
		}
	}
}

func (fs *memFileSystem) ListSegments(context *Context, pathname string) ([]FileStat, error) {
	wd := fs.workingDirectoryNode(context, pathname)
	n := wd.getFileNode(NewPathWithDelimiter(pathname, fs.pathDelimiter), 0)

	if n == nil {
		return nil, &MemFileSystemError{Err: noSuchFileOrDirectoryErr, Op: "ListSegments", Path: pathname}
	} else {
		result := make([]FileStat, 0)
		for _, child := range wd.children {
			result = append(result, child.file.Stat().Immutable())
		}
		return result, nil
	}
}

func (fs *memFileSystem) PresentWorkingDirectoryNode(context *Context) *fileNode {
	return fs.pwd[context]
}

func (fs *memFileSystem) workingDirectoryNode(context *Context, pathname string) *fileNode {
	// if pathname starts with path delimiter (like "/"),
	// then start on root node
	if strings.HasPrefix(pathname, fs.pathDelimiter) {
		return fs.rootNode
	} else {
		return fs.PresentWorkingDirectoryNode(context)
	}
}