package archive

import (
	"encoding/json"
	"flag"
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

type archiveCommandFlags struct {
	command.BaseFlags
	platform    string
	source      string
	destination string
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
	_, flagSet, _ := getConfig([]string{})
	flagSet.PrintDefaults()
	return ""
}

func getConfig(args []string) (archiveCommandFlags, *flag.FlagSet, int) {
	var cmdConfig archiveCommandFlags

	cmdFlags := flag.NewFlagSet("archive", flag.ExitOnError)
	cmdFlags.BoolVar(&cmdConfig.Help, "help", false,
		"show command help")
	cmdFlags.StringVar(&cmdConfig.platform, "platform", "", "platform type (allowed: nodejs|python)")
	cmdFlags.StringVar(&cmdConfig.source, "source", "", "project directory (default: .)")
	cmdFlags.StringVar(&cmdConfig.destination, "destination", "", "dependencies download destination")

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

	// All missing required arguments checks go here
	var missingArg string
	if cmdConfig.platform == "" {
		missingArg = "platform"
	} else if cmdConfig.destination == "" {
		missingArg = "destination"
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

	var dirPath string
	if cmdConfig.source == "" {
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
		dirPath = cmdConfig.source
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
