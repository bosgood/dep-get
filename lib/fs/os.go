package fs

import (
	"io/ioutil"
	"os"
)

// OSFS implements fileSystem using the local disk.
type OSFS struct{}

// Open opens a file
func (f *OSFS) Open(name string) (File, error) {
	return os.Open(name)
}

// Create creates a file
func (f *OSFS) Create(name string) (File, error) {
	return os.Create(name)
}

// Stat stats a file
func (f *OSFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Getwd gets a wd
func (f *OSFS) Getwd() (string, error) {
	return os.Getwd()
}

// ReadFile reads a file
func (f *OSFS) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}
