package vfs

import "strings"

const (
	DEFAULT_PATH_DELIMITER = "/"
)
type Path struct {
	paths []string
	filename string
	delimiter string
	str 	string
	isRoot 	bool
}

type Iterator interface {
	Value() string
	HasNext() bool
}

type pathIterator struct {
	path *Path
	current int
}

func NewPath(path string) *Path {
	return NewPathWithDelimiter(path, DEFAULT_PATH_DELIMITER)
}

func NewPathWithDelimiter(path string, delimiter string) *Path {
	newPath := &Path{
		paths: make([]string, 0),
		delimiter: delimiter,
	}

	// flag to indicate root path
	if strings.HasPrefix(path, delimiter) {
		newPath.isRoot = true
	}

	// path not starts/ends with delimiter
	path = strings.TrimPrefix(path, delimiter)
	path = strings.TrimSuffix(path, delimiter)

	splits := strings.Split(path, delimiter)

	for _, p := range splits {
		if p != "" {
			newPath.paths = append(newPath.paths, p)
		}
	}

	if strings.HasSuffix(path, delimiter) || len(newPath.paths) == 0 {
		newPath.filename = ""
	} else {
		newPath.filename = newPath.paths[len(newPath.paths) - 1]
	}

	if newPath.isRoot {
		newPath.str = strings.Join(newPath.paths, delimiter)
		newPath.str = delimiter + newPath.str
	} else {
		newPath.str = strings.Join(newPath.paths, delimiter)
	}

	return newPath
}

func (p *Path) Concat(other *Path) *Path {
	if other != nil {
		if other.isRoot || p.String() == "/" {
			return NewPathWithDelimiter(p.String() + other.String(), p.delimiter)
		} else {
			return NewPathWithDelimiter(p.String() + p.delimiter + other.String(), p.delimiter)
		}
	} else {
		return NewPathWithDelimiter(p.String(), p.delimiter)
	}
}

func (p *Path) Len() int {
	return len(p.paths)
}

func (p *Path) NthPath(i int) string {
	return p.paths[i]
}

func (p *Path) FileName() string {
	return p.filename
}

func (p *Path) Iterator() Iterator {
	return newPathIterator(p)
}

func (p *Path) String() string {
	return p.str
}

func newPathIterator(path *Path) *pathIterator {
	return &pathIterator{
		path: path,
		current: -1,
	}
}

func (pi *pathIterator) Value() string {
	return pi.path.paths[pi.current]
}

func (pi *pathIterator) HasNext() bool {
	pi.current++
	return len(pi.path.paths) > pi.current
}