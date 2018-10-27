package store

/**
 Directory-based Store Interface
 */
type Store interface {
	IsFileExist(filename string)	bool
	FileIter()	<-chan FileInfo
	FileInfo(filename string)	(FileInfo, error)
	Read(filename string, res []byte, startOffset int64) error
	Write(filename string, data []byte, startOffset int64) error
	Clear(filename string, startOffset int64, size int64) error
	CreateFile(filename string) error
	RemoveFile(filename string) error
	SubStore(subpath string) Store
	Truncate(filename string, size int64) error
}

type FileInfo interface {
	Name()	string
	Size()	int64
}

type fileInfo struct {
	name	string
	size	int64
}

func (f *fileInfo) Name() string {
	return f.name
}

func (f *fileInfo) Size() int64 {
	return f.size
}
