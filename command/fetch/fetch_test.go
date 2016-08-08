package fetch

import (
	"github.com/bosgood/dep-get/lib/fs"
	"testing"
)

type invalidAssertionError struct {
	message string
}

func (e *invalidAssertionError) Error() string {
	if e.message != "" {
		return e.message
	}
	return "Invalid assertion"
}

func TestFetchCommandBasics(t *testing.T) {
	cmd, err := NewFetchCommand()

	if err != nil {
		t.Errorf("err: %s", err)
	}

	if cmd.Synopsis() == "" {
		t.Errorf("Err: No synopsis text")
	}

	// Disabled, output is printed directly
	// if cmd.Help() == "" {
	// 	t.Errorf("Err: No help text")
	// }
}

func TestFetchCommandHelp(t *testing.T) {
	mockFS := &fs.MockFS{
		ReadFileError: &invalidAssertionError{},
	}
	cmd, err := newFetchCommandWithFS(mockFS)

	if err != nil {
		t.Errorf("err: %s", err)
	}

	if cmd.Run([]string{"--help"}) != -18511 {
		t.Errorf("Err: non-zero return value for --help")
	}
}

func TestFetchCommandLoadFile(t *testing.T) {
	var dummyPackageJSONBytes = []byte("{}")
	mockFS := &fs.MockFS{
		ReadFileResult: dummyPackageJSONBytes,
	}
	cmd, err := newFetchCommandWithFS(mockFS)

	if err != nil {
		t.Errorf("err: %s", err)
	}

	if cmd.Run([]string{"--platform", "nodejs", "--destination", "tmp/"}) != 0 {
		t.Errorf("Err: non-zero return value for no args")
	}
}
