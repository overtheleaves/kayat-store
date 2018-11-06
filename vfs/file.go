package vfs

import (
	"os"
	"strings"
	"io/ioutil"
)

var (
	mountInfoFile = ".vfs_mount_info"
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

	if !strings.HasPrefix(mountOnPath, delimiter) {
		return nil, &WrapperFileSystemError{Err: invalidMountOnPathErr, Op: "mount", Path: mountOnPath}
	}

	if strings.HasSuffix(mountOnPath, delimiter) {
		mountOnPath = mountOnPath[:len(mountOnPath) - 1]
	}

	nested, nestedPath := isNestedFilePath(mountOnPath, delimiter)
	if nested {
		return nil, &WrapperFileSystemError{Err: nestedMountedErr(nestedPath), Op: "mount", Path: mountOnPath}
	}

	wfs := &wrapperFileSystem {
		pwd:           make(map[*Context]*Path),
		mount:         NewPathWithDelimiter(mountOnPath, delimiter),
		pathDelimiter: delimiter,
	}

	// is directory existed?
	if !isDirectoryExist(mountOnPath) {
		// if not, create directory first
		err := os.MkdirAll(mountOnPath, os.ModePerm)
		if err != nil {
			return nil, &WrapperFileSystemError{Err: err, Op: "mount", Path: mountOnPath}
		}
	}

	// create mount info file
	f, err := os.Create(mountOnPath + delimiter + mountInfoFile)
	if err != nil {
		return nil, &WrapperFileSystemError{Err: err, Op: "mount", Path: mountOnPath}
	} else {
		f.Write([]byte(mountOnPath))
		f.Close()
	}

	return wfs, nil
}

func (w *wrapperFileSystem) NewFile(context *Context, pathname string) (File, error) {

	filepath := NewPathWithDelimiter(pathname, w.pathDelimiter)
	filename := filepath.FileName()

	if filename == "" {
		return nil, &WrapperFileSystemError{Err: illegalFileNameErr, Op: "NewFile", Path: pathname}
	}

	fullPath := w.workingDirectory(context, pathname) + w.pathDelimiter + pathname
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
	fullPath := w.workingDirectory(context, pathname) + w.pathDelimiter + pathname

	err := os.RemoveAll(fullPath)

	if err != nil {
		return &WrapperFileSystemError{Err: err, Op: "Remove", Path: pathname}
	} else {
		return nil
	}
}

func (w *wrapperFileSystem) OpenFile(context *Context, pathname string)	(File, error) {
	fullPath := w.workingDirectory(context, pathname) + w.pathDelimiter + pathname

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
	if w.FileExisted(context, pathname) {
		return &WrapperFileSystemError{Err: fileExistsErr, Op: "Mkdir", Path: pathname}
	} else {
		fullPath := w.workingDirectory(context, pathname) + w.pathDelimiter + pathname
		err := os.MkdirAll(fullPath, os.ModePerm)
		if err != nil {
			return &WrapperFileSystemError{Err: err, Op: "Mkdir", Path: pathname}
		}
	}

	return nil
}

func (w *wrapperFileSystem) FileExisted(context *Context, pathname string) bool {
	fullPath := w.workingDirectory(context, pathname) + w.pathDelimiter + pathname
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
		if strings.HasPrefix(pathname, w.pathDelimiter) {
			// replace
			w.pwd[context] = NewPathWithDelimiter(pathname, w.pathDelimiter)
		} else {
			// concat
			w.pwd[context] = w.pwd[context].Concat(NewPathWithDelimiter(pathname, w.pathDelimiter))
		}
		return nil
	}
}

func (w *wrapperFileSystem) Context() *Context {
	context := &Context{}
	w.pwd[context] = NewPathWithDelimiter("/", w.pathDelimiter)
	return context
}

func (w *wrapperFileSystem) ListSegments(context *Context, pathname string) ([]FileStat, error) {
	return nil, nil
}

func (w *wrapperFileSystem) PresentWorkingDirectory(context *Context) *Path {
	return w.pwd[context]
}

func (w *wrapperFileSystem) workingDirectory(context *Context, pathname string) string {
	// if pathname starts with __dir_name_ pathDelimiter (like "/"),
	// then start on mount root
	if strings.HasPrefix(pathname, w.pathDelimiter) {
		return w.mount.String()
	} else {
		return w.mount.Concat(w.PresentWorkingDirectory(context)).String()
	}
}

func isDirectoryExist(path string) bool {
	fileinfo, err := os.Stat(path)
	return !os.IsNotExist(err) && fileinfo != nil && fileinfo.IsDir()
}

func isNestedFilePath(path string, delimiter string) (bool, string) {

	// find path already mounted
	info, _ := os.Stat(path + delimiter + mountInfoFile)
	if info != nil && info.Name() == mountInfoFile {
		return true, path
	}

	// find children paths already mounted
	found, foundPath := findMountedChildren(path, delimiter)
	if found {
		return found, foundPath
	}

	// find parent paths already mounted
	p := NewPathWithDelimiter(path, delimiter)

	for i := p.Len() - 1; i >= 0; i-- {
		path = strings.TrimSuffix(path, p.NthPath(i))
		info, err := os.Stat(path + mountInfoFile)

		if err != nil {
			// cannot retrieve anymore
			return false, ""
		}

		if info != nil && info.Name() == mountInfoFile {
			return true, path
		}
	}

	return false, ""
}

func findMountedChildren(path string, delimiter string) (bool, string) {
	info, err := os.Stat(path)

	if err != nil {
		return false, ""	// cannot retrieve anymore
	}

	// find .vfs_mount_info file
	if info.Name() == mountInfoFile {
		return true, path
	}

	if info.IsDir() {
		infos, err := ioutil.ReadDir(path)
		if err != nil {
			return false, "" // cannot retrieve anymore
		} else {
			for _, i := range infos {
				found, path := findMountedChildren(path + delimiter + i.Name(), delimiter)
				if found {
					return found, path
				}
			}
		}
	}

	return false, ""
}

