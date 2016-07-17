package fs

import (
	"os"
)

// OSFS implements fileSystem using the local disk.
type OSFS struct{}

// Open opens a file
func (OSFS) Open(name string) (File, error) {
	return os.Open(name)
}

// Stat stats a file
func (OSFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Getwd gets a wd
func (OSFS) Getwd() (string, error) {
	return os.Getwd()
}
