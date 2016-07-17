package archive

import (
	"encoding/json"
	"fmt"
	"github.com/bosgood/dep-get/command"
	"github.com/bosgood/dep-get/lib/fs"
	"github.com/bosgood/dep-get/nodejs"
	"github.com/mitchellh/cli"
	"path"
)

var realOS fs.FileSystem = &fs.OSFS{}

type archiveCommand struct {
	npmShrinkwrap nodejs.NPMShrinkwrap
	os            fs.FileSystem
}

func newArchiveCommandWithFS(os fs.FileSystem) (cli.Command, error) {
	cmd := &archiveCommand{
		os: os,
	}
	return cmd, nil
}

// NewArchiveCommand is used to generate a command object
// which orchestrates package dependency discovery
// and archiving to S3
func NewArchiveCommand() (cli.Command, error) {
	return newArchiveCommandWithFS(realOS)
}

func (c *archiveCommand) Synopsis() string {
	return "Archives application dependencies"
}

func (c *archiveCommand) Help() string {
	return "I'm super helpful"
}

func (c *archiveCommand) Run(args []string) int {
	var dirPath string

	if len(args) == 0 {
		cwd, err := c.os.Getwd()
		if err != nil {
			fmt.Printf(
				"%s%s: %s\n",
				command.LogErrorPrefix,
				"Can't read current directory",
				err,
			)
			return 1
		}
		dirPath = cwd
	} else {
		if args[0] == "--help" {
			return cli.RunResultHelp
		}

		dirPath = args[0]
	}

	packageFilePath := path.Join(dirPath, nodejs.DependenciesFileName)
	packageFileContents, err := c.os.ReadFile(packageFilePath)
	if err != nil {
		fmt.Printf(
			"%s%s: %s\n",
			command.LogErrorPrefix,
			"Can't open the dependencies file",
			err,
		)
		return 1
	}

	var npmShrinkwrap nodejs.NPMShrinkwrap
	err = json.Unmarshal(packageFileContents, &npmShrinkwrap)
	if err != nil {
		fmt.Printf(
			"%s%s: %s\n",
			command.LogErrorPrefix,
			"Failed to decode the dependencies file file",
			err,
		)
		return 1
	}

	c.npmShrinkwrap = npmShrinkwrap

	// fmt.Printf(
	// 	"%s%s: %s",
	// 	command.LogSuccessPrefix,
	// 	"Read dependencies file",
	// 	npmShrinkwrap,
	// )

	fmt.Printf(
		"%sFound %d top-level dependencies.\n",
		command.LogSuccessPrefix,
		len(npmShrinkwrap.Dependencies),
	)

	for k, v := range npmShrinkwrap.Dependencies {
		fmt.Printf("%s: %s\n", k, v.Version)
	}

	return 0
}
