package archive

import (
	"github.com/bosgood/dep-get/command"
	"github.com/mitchellh/cli"
)

// NewArchiveCommand is used to generate a command object
// which orchestrates package dependency discovery
// and archiving to S3
func NewArchiveCommand() (cli.Command, error) {
	cmd := &command.Command{
		SynopsisText: "yup",
		RetVal:       0,
	}
	return cmd, nil
}
