package install

import (
	"testing"
)

func TestNewInstallCommand(t *testing.T) {
	cmd, err := NewInstallCommand()

	if err != nil {
		t.Errorf("err: %s", err)
	}

	if cmd.Synopsis() == "" {
		t.Errorf("Err: No synopsis text")
	}

	if cmd.Help() == "" {
		t.Errorf("Err: No help text")
	}

	if cmd.Run([]string{"--help"}) != 0 {
		t.Errorf("Err: non-zero return value for --help")
	}
}
