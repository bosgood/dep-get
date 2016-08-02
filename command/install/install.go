package install

import (
	"flag"
	"fmt"
	"github.com/bosgood/dep-get/command"
	"github.com/bosgood/dep-get/lib/fs"
	"github.com/mitchellh/cli"
	"regexp"
)

type installCommand struct {
	os     fs.FileSystem
	config installCommandFlags
}

type installCommandFlags struct {
	command.BaseFlags
	platform     string
	source       string
	whitelistStr string
	whitelist    *regexp.Regexp
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

func getConfig(args []string) (installCommandFlags, *flag.FlagSet, int) {
	var cmdConfig installCommandFlags

	cmdFlags := flag.NewFlagSet("install", flag.ExitOnError)
	cmdFlags.BoolVar(&cmdConfig.Help, "help", false, "show command help")

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

	return cmdConfig, cmdFlags, 0
}

func (c *installCommand) Run(args []string) int {
	_, _, ret := getConfig(args)
	if ret != 0 {
		return ret
	}
	return 0
}
