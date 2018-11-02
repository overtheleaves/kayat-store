package vfs

import (
	"os"
	"strings"
)

/**
 os filesystem wrapper
*/

type wrapperFileSystem struct {
	pwd           map[*Context]*Path
	mount         *Path
	pathDelimiter string
}

type wrapperFile struct {
	f *os.File
}

type WrapperFileSystemError struct {
	Err error
	Op string
	Path string
}

func (e *WrapperFileSystemError) Error() string {
	return e.Op + ": " + e.Path + ": " + e.Err.Error()
}

func newWrapperFile(f *os.File) *wrapperFile {
	return &wrapperFile{f: f}
}

func (f *wrapperFile) Stat() FileStat {
	return nil
}

func (f *wrapperFile) Read(b []byte) (n int, err error) {
	return f.f.Read(b)
}

func (f *wrapperFile) ReadAt(b []byte, off int64) (n int, err error) {
	return f.f.ReadAt(b, off)
}

func (f *wrapperFile) Write(b []byte) (n int, err error) {
	return f.f.Write(b)
}

func (f *wrapperFile) WriteAt(b []byte, off int64) (n int, err error) {
	return f.f.WriteAt(b, off)
}

func (f *wrapperFile) Delete() {

}

func NewWrapperFileSystem(mountOnPath string) (VirtualFileSystem, error) {
	return NewWrapperFileSystemWithPathDelimiter(mountOnPath, DEFAULT_PATH_DELIMITER)
}

func NewWrapperFileSystemWithPathDelimiter(mountOnPath string, delimiter string) (VirtualFileSystem, error) {
	return &wrapperFileSystem{
		pwd:           make(map[*Context]*Path),
		mount:         NewPathWithDelimiter(mountOnPath, delimiter),
		pathDelimiter: delimiter,
	}, nil
}

func (w *wrapperFileSystem) NewFile(context *Context, pathname string) (File, error) {

	filepath := NewPathWithDelimiter(pathname, w.pathDelimiter)
	filename := filepath.FileName()

	if filename == "" {
		return nil, &MemFileSystemError{Err: illegalFileNameErr, Op: "NewFile", Path: pathname}
	}

	wd := w.workingDirectory(context, pathname)

	// is directory existed?
	if !isDirectoryExist(wd + w.pathDelimiter + filepath.PathString()) {
		err := os.MkdirAll(wd + w.pathDelimiter + filepath.PathString(), os.ModePerm)
		if err != nil {
			return nil, &WrapperFileSystemError{Err: err, Op: "NewFile", Path: pathname}
		}
	}

	f, err := os.Create(wd + w.pathDelimiter + pathname)
	if f != nil {
		defer f.Close()
	}

	if err != nil {
		return nil, &WrapperFileSystemError{Err: err, Op: "NewFile", Path: pathname}
	}

	file := newWrapperFile(f)

	return file, err
}

func (w *wrapperFileSystem) Remove(context *Context, pathname string) error {
	return nil
}

func (w *wrapperFileSystem) OpenFile(context *Context, name string)	(File, error) {
	return nil, nil
}

func (w *wrapperFileSystem) Create(context *Context, name string) (File, error) {
	return nil, nil
}

func (w *wrapperFileSystem) Mkdir(context *Context, pathname string) error {
	return nil
}

func (w *wrapperFileSystem) FileExisted(context *Context, name string) bool {
	return true
}

func (w *wrapperFileSystem) ChangeDirectory(context *Context, pathname string) error {
	return nil
}

func (w *wrapperFileSystem) Context() *Context {
	context := &Context{}
	w.pwd[context] = w.mount
	return context
}

func (w *wrapperFileSystem) ListSegments(context *Context, pathname string) ([]FileStat, error) {
	return nil, nil
}

func (w *wrapperFileSystem) PresentWorkingDirectory(context *Context) string {
	pwd := w.pwd[context]
	if pwd != nil {
		return pwd.FullPathString()
	} else {
		return ""
	}
}
func (w *wrapperFileSystem) workingDirectory(context *Context, pathname string) string {
	// if pathname starts with path pathDelimiter (like "/"),
	// then start on mount root
	if strings.HasPrefix(pathname, w.pathDelimiter) {
		return w.mount.FullPathString()
	} else {
		return w.PresentWorkingDirectory(context)
	}
}

func isDirectoryExist(path string) bool {
	fileinfo, err := os.Stat(path)
	return !os.IsNotExist(err) && fileinfo != nil && fileinfo.IsDir()
}
