package install

import (
	"github.com/bosgood/dep-get/command"
	"github.com/mitchellh/cli"
)

// NewInstallCommand is used to generate a command object
// which downloads dependencies from S3 and installs them
func NewInstallCommand() (cli.Command, error) {
	cmd := &command.Command{
		SynopsisText: "nope",
		RetVal:       0,
	}
	return cmd, nil
}
