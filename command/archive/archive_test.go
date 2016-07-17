package archive

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

func TestArchiveCommandBasics(t *testing.T) {
	cmd, err := NewArchiveCommand()

	if err != nil {
		t.Errorf("err: %s", err)
	}

	if cmd.Synopsis() == "" {
		t.Errorf("Err: No synopsis text")
	}

	if cmd.Help() == "" {
		t.Errorf("Err: No help text")
	}
}

func TestArchiveCommandHelp(t *testing.T) {
	mockFS := &fs.MockFS{
		ReadFileError: &invalidAssertionError{},
	}
	cmd, err := newArchiveCommandWithFS(mockFS)

	if err != nil {
		t.Errorf("err: %s", err)
	}

	if cmd.Run([]string{"--help"}) != 0 {
		t.Errorf("Err: non-zero return value for --help")
	}
}

func TestArchiveCommandLoadFile(t *testing.T) {
	var dummyPackageJSONBytes = []byte("{}")
	mockFS := &fs.MockFS{
		ReadFileResult: dummyPackageJSONBytes,
	}
	cmd, err := newArchiveCommandWithFS(mockFS)

	if err != nil {
		t.Errorf("err: %s", err)
	}

	if cmd.Run([]string{}) != 0 {
		t.Errorf("Err: non-zero return value for no args")
	}
}
