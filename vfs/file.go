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

	fullPath := w.workingFullPathString(context, pathname)
	path := strings.TrimSuffix(fullPath, filename)

	// is directory existed?
	if !isDirectoryExist(path) {
		// if not, create directory first
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return nil, &WrapperFileSystemError{Err: err, Op: "NewFile", Path: pathname}
		}
	}

	// file create
	f, err := os.Create(fullPath)
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

	// check if file existed.
	if !w.FileExisted(context, pathname) {
		return &WrapperFileSystemError{Err: noSuchFileOrDirectoryErr, Op: "Remove", Path: pathname}
	}

	// get context's working directory
	fullPath := w.workingFullPathString(context, pathname)

	err := os.RemoveAll(fullPath)

	if err != nil {
		return &WrapperFileSystemError{Err: err, Op: "Remove", Path: pathname}
	} else {
		return nil
	}
}

func (w *wrapperFileSystem) OpenFile(context *Context, pathname string)	(File, error) {
	fullPath := w.workingFullPathString(context, pathname)
	f, err := os.OpenFile(fullPath, os.O_RDWR, os.ModeAppend)

	if err != nil {
		return nil, &WrapperFileSystemError{Err: err, Op: "OpenFile", Path: pathname}
	} else {
		return newWrapperFile(f), nil
	}
}

func (w *wrapperFileSystem) Create(context *Context, pathname string) (File, error) {
	return w.NewFile(context, pathname)
}

func (w *wrapperFileSystem) Mkdir(context *Context, pathname string) error {
	return nil
}

func (w *wrapperFileSystem) FileExisted(context *Context, pathname string) bool {
	fullPath := w.workingFullPathString(context, pathname)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}

func (w *wrapperFileSystem) ChangeDirectory(context *Context, pathname string) error {

	if w.pwd[context] == nil {
		return &WrapperFileSystemError{ Err: invalidContextErr, Op: "ChangeDirectory", Path: pathname}
	}

	if !w.FileExisted(context, pathname) {
		return &WrapperFileSystemError{Err: noSuchFileOrDirectoryErr, Op: "ChangeDirectory", Path: pathname}
	} else {
		fullPathString := w.workingFullPathString(context, pathname)
		fullPathString = strings.TrimSuffix(fullPathString, w.pathDelimiter)
		w.pwd[context] = NewPathWithDelimiter(fullPathString, w.pathDelimiter)
		return nil
	}
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

func (w *wrapperFileSystem) workingFullPathString(context *Context, pathname string) string {
	wd := w.workingDirectory(context, pathname)

	pathname = strings.TrimPrefix(pathname, w.pathDelimiter)
	return wd + w.pathDelimiter + pathname
}

func isDirectoryExist(path string) bool {
	fileinfo, err := os.Stat(path)
	return !os.IsNotExist(err) && fileinfo != nil && fileinfo.IsDir()
}
