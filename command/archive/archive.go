package archive

import (
	"flag"
	"fmt"
	"github.com/bosgood/dep-get/command"
	"github.com/bosgood/dep-get/lib/fs"
	"github.com/mitchellh/cli"
	"path"
)

type archiveCommand struct {
	os     fs.FileSystem
	config archiveCommandFlags
}

type archiveCommandFlags struct {
	command.BaseFlags
	platform string
	source   string
}

var realOS fs.FileSystem = &fs.OSFS{}

func newArchiveCommandWithFS(os fs.FileSystem) (cli.Command, error) {
	cmd := &archiveCommand{
		os: os,
	}
	return cmd, nil
}

// NewArchiveCommand is used to generate a command object
// which downloads dependencies from S3 and installs them
func NewArchiveCommand() (cli.Command, error) {
	return newArchiveCommandWithFS(realOS)
}

func (c *archiveCommand) Synopsis() string {
	return "Installs archived application dependencies"
}

func (c *archiveCommand) Help() string {
	_, flagSet, _ := getConfig([]string{})
	flagSet.PrintDefaults()
	return ""
}

func getConfig(args []string) (archiveCommandFlags, *flag.FlagSet, int) {
	var cmdConfig archiveCommandFlags

	cmdFlags := flag.NewFlagSet("install", flag.ExitOnError)
	cmdFlags.BoolVar(&cmdConfig.Help, "help", false, "show command help")
	cmdFlags.StringVar(&cmdConfig.platform, "platform", "", "platform type (allowed: nodejs|python)")
	cmdFlags.StringVar(&cmdConfig.source, "source", "", "project directory (default: .)")

	if err := cmdFlags.Parse(args); err != nil {
		fmt.Printf(
			"%s%s: %s\n",
			command.LogErrorPrefix,
			"Error parsing args",
			err,
		)
		return cmdConfig, cmdFlags, 1
	}

	if cmdConfig.Help {
		return cmdConfig, cmdFlags, cli.RunResultHelp
	}

	// All missing required argument checks go here
	var missingArg string
	if cmdConfig.platform == "" {
		missingArg = "platform"
	}

	if missingArg != "" {
		fmt.Printf(
			"%s%s: %s\n",
			command.LogErrorPrefix,
			"Missing required argument",
			missingArg,
		)
		return cmdConfig, cmdFlags, cli.RunResultHelp
	}

	if cmdConfig.platform != "nodejs" {
		fmt.Printf(
			"%s%s\n",
			command.LogErrorPrefix,
			"Only nodejs supported at the moment",
		)
		return cmdConfig, cmdFlags, 1
	}

	return cmdConfig, cmdFlags, 0
}

func (c *archiveCommand) Run(args []string) int {
	cmdConfig, _, ret := getConfig(args)
	if ret != 0 {
		return ret
	}

	archives, err := c.os.ReadDir(cmdConfig.source)
	if err != nil {
		fmt.Printf(
			"%sError reading archive path: %s\n",
			command.LogErrorPrefix,
			err,
		)
	}

	for _, fileInfo := range archives {
		archiveFilePath := path.Join(
			cmdConfig.source,
			fileInfo.Name(),
		)

		fmt.Printf(
			"%sArchiving dependency: %s\n",
			command.LogInfoPrefix,
			archiveFilePath,
		)
	}

	return 0
}
