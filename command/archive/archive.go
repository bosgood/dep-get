package archive

import (
	"bitbucket.org/bosgood/dep-get/command"
	"bitbucket.org/bosgood/dep-get/lib/fs"
	"flag"
	"fmt"
	// "github.com/aws/aws-sdk-go/service/s3"
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

func getConfig(args []string) (archiveCommandFlags, *flag.FlagSet, error) {
	var cmdConfig archiveCommandFlags

	cmdFlags := flag.NewFlagSet("install", flag.ExitOnError)
	cmdFlags.BoolVar(&cmdConfig.Help, "help", false, "show command help")
	cmdFlags.StringVar(&cmdConfig.platform, "platform", "", "platform type (allowed: nodejs|python)")
	cmdFlags.StringVar(&cmdConfig.source, "source", "", "project directory (default: .)")

	if err := cmdFlags.Parse(args); err != nil {
		errMsg := fmt.Sprintf(
			"%s%s: %s\n",
			command.LogErrorPrefix,
			"Error parsing args",
			err,
		)
		return cmdConfig, cmdFlags, &command.ConfigError{
			Explanation: errMsg,
		}
	}

	if cmdConfig.Help {
		return cmdConfig, cmdFlags, &command.ConfigError{}
	}

	// All missing required argument checks go here
	var missingArg string
	if cmdConfig.platform == "" {
		missingArg = "platform"
	}

	if cmdConfig.source == "" {
		missingArg = "source"
	}

	if missingArg != "" {
		errMsg := fmt.Sprintf(
			"%s%s: %s\n",
			command.LogErrorPrefix,
			"Missing required argument",
			missingArg,
		)
		return cmdConfig, cmdFlags, &command.ConfigError{
			Explanation: errMsg,
		}
	}

	if cmdConfig.platform != "nodejs" {
		errMsg := fmt.Sprintf(
			"%s%s\n",
			command.LogErrorPrefix,
			"Only nodejs supported at the moment",
		)
		return cmdConfig, cmdFlags, &command.ConfigError{
			Explanation: errMsg,
		}
	}

	return cmdConfig, cmdFlags, nil
}

func (c *archiveCommand) Run(args []string) int {
	cmdConfig, _, err := getConfig(args)
	if err != nil {
		errMsg := err.Error()
		if errMsg != "" {
			fmt.Print(err.Error())
		}
		return cli.RunResultHelp
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
