package fs

import (
	"io"
	"os"
)

// FileSystem abstracts a filesystem, real or mock
type FileSystem interface {
	Open(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
	Getwd() (string, error)
}

// File represents file-based interactions
type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	Stat() (os.FileInfo, error)
}
