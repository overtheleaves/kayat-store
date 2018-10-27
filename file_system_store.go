package store

import (
	"os"
	"io/ioutil"
	"strings"
)

type fileSystemStore struct {
	path string
}

func NewFileSystemStore(path string) Store {
	// check directory exists
	// if not, create
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	if !isDirectoryExist(path) {
		os.MkdirAll(path, os.ModePerm)
	}

	fs := &fileSystemStore{path: path}
	return fs
}

func RemoveFileSystemStore(path string) {
	os.RemoveAll(path)
}

func (fs *fileSystemStore) SubStore(subpath string) Store {
	if strings.HasPrefix(subpath, "/") {
		subpath = subpath[1:]
	}
	return NewFileSystemStore(fs.path + subpath)
}

func (fs *fileSystemStore) IsFileExist(filename string)	bool {
	return isFileExist(fs.path + filename)
}

func (fs *fileSystemStore) FileIter() <-chan FileInfo {

	ch := make(chan FileInfo)
	files, err := ioutil.ReadDir(fs.path)

	if err != nil {
		return nil
	} else {
		go func(files []os.FileInfo) {
			for _, elem := range files {

				if !elem.(os.FileInfo).IsDir() {
					// iterate files, only
					ch <- &fileInfo{elem.Name(), elem.Size()}
				}
			}

			close(ch)
		}(files)

		return ch
	}
}

func (fs *fileSystemStore) FileInfo(filename string) (FileInfo, error) {
	info, err := os.Stat(fs.path + filename)
	return info, err
}

func (fs *fileSystemStore) Read(filename string, res []byte, startOffset int64) error {
	f, err := fs.openFile(filename)

	if f != nil {
		defer f.Close()
		err := readBytes(f, res, startOffset)
		return err
	} else {
		return &os.PathError{Op: "Read", Path: fs.path + filename, Err: err}
	}
}

func (fs *fileSystemStore) Write(filename string, data []byte, startOffset int64) error {
	f, err := fs.openFile(filename)
	if f != nil {
		defer f.Close()
		err := writeBytes(f, data, startOffset)
		return err
	} else {
		return &os.PathError{Op: "Write", Path: fs.path + filename, Err: err}
	}
}

func (fs *fileSystemStore) CreateFile(filename string) error {
	f, err := os.Create(fs.path + filename)
	if f != nil {
		defer f.Close()
	}

	return err
}

func (fs *fileSystemStore) RemoveFile(filename string) error {
	return os.Remove(fs.path + filename)
}

func (fs *fileSystemStore) Clear(filename string, startOffset int64, size int64) error {
	f, err := fs.openFile(filename)
	if f != nil {
		defer f.Close()
		data := make([]byte, size)
		return writeBytes(f, data, startOffset)
	} else {
		return &os.PathError{Op: "Clear", Path: fs.path + filename, Err: err}
	}
}

func (fs *fileSystemStore) Truncate(filename string, size int64) (err error) {
	return os.Truncate(fs.path + filename, size)
}


func (fs *fileSystemStore) openFile(filename string) (*os.File, error) {
	return os.OpenFile(fs.path + filename, os.O_RDWR, os.ModeAppend)
}

func isFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func isDirectoryExist(path string) bool {
	fileinfo, err := os.Stat(path)
	return !os.IsNotExist(err) && fileinfo != nil && fileinfo.IsDir()
}

func writeBytes(f *os.File, data []byte, startOffset int64) error {
	_, err := f.WriteAt(data, startOffset)
	return err
}

func readBytes(f *os.File, res[] byte, startOffset int64) error {
	_, err := f.ReadAt(res, startOffset)
	return err
}
