package fs

import (
	"os"
)

// MockFS implements fileSystem using mock data
type MockFS struct {
	OpenResult     File
	OpenError      error
	StatResult     os.FileInfo
	StatError      error
	GetwdResult    string
	GetwdError     error
	ReadFileResult []byte
	ReadFileError  error
}

// Open opens a file
func (m MockFS) Open(name string) (File, error) {
	return m.OpenResult, m.OpenError
}

// Stat stats a file
func (m MockFS) Stat(name string) (os.FileInfo, error) {
	return m.StatResult, m.StatError
}

// Getwd gets a wd
func (m MockFS) Getwd() (string, error) {
	return m.GetwdResult, m.GetwdError
}

// ReadFile reads a file
func (m MockFS) ReadFile(filename string) ([]byte, error) {
	return m.ReadFileResult, m.ReadFileError
}
