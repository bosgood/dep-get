package install

import (
	"flag"
	"fmt"
	"github.com/bosgood/dep-get/command"
	"github.com/bosgood/dep-get/lib/fs"
	"github.com/mitchellh/cli"
	"os/exec"
	"path"
)

type installCommand struct {
	os     fs.FileSystem
	config installCommandFlags
}

type installCommandFlags struct {
	command.BaseFlags
	platform string
	source   string
}

var realOS fs.FileSystem = &fs.OSFS{}

func newInstallCommandWithFS(os fs.FileSystem) (cli.Command, error) {
	cmd := &installCommand{
		os: os,
	}
	return cmd, nil
}

// NewInstallCommand is used to generate a command object
// which downloads dependencies from S3 and installs them
func NewInstallCommand() (cli.Command, error) {
	return newInstallCommandWithFS(realOS)
}

func (c *installCommand) Synopsis() string {
	return "Installs archived application dependencies"
}

func (c *installCommand) Help() string {
	_, flagSet, _ := getConfig([]string{})
	flagSet.PrintDefaults()
	return ""
}

func getConfig(args []string) (installCommandFlags, *flag.FlagSet, error) {
	var cmdConfig installCommandFlags

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

func (c *installCommand) Run(args []string) int {
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
			"%sCaching dependency: %s\n",
			command.LogInfoPrefix,
			fileInfo.Name(),
		)

		cmd := exec.Command("npm", "cache", "add", archiveFilePath)
		err = cmd.Run()
		if err != nil {
			fmt.Printf(
				"%sError executing command: %s\n",
				command.LogErrorPrefix,
				err,
			)
			return 1
		}
	}

	return 0
}
