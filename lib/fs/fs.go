package fs

import (
	"io"
	"os"
)

// FileSystem abstracts a filesystem, real or mock
type FileSystem interface {
	Open(name string) (File, error)
	Create(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
	Getwd() (string, error)
	ReadFile(filename string) ([]byte, error)
	ReadDir(dirpath string) ([]os.FileInfo, error)
}

// File represents file-based interactions
type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	Stat() (os.FileInfo, error)
}
