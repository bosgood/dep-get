package archive

import (
	"encoding/json"
	"github.com/bosgood/dep-get/command"
	"github.com/bosgood/dep-get/nodejs"
	"github.com/bosgood/dep-get/lib/fs"
	"github.com/mitchellh/cli"
	"log"
	"path"
)

var realOS fs.FileSystem = fs.OSFS{}

type archiveCommand struct {
	packageJSON nodejs.PackageJSON
	os fs.FileSystem
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
			log.Printf(
				"%s%s: %s",
				command.LogErrorPrefix,
				"Can't read current directory",
				err,
			)
			return 1
		}
		dirPath = cwd
	} else {
		if args[0] == "--help" {
			log.Println(c.Help())
			return 0
		}

		dirPath = args[0]
	}

	packageFilePath := path.Join(dirPath, "package.json")
	packageFileContents, err := c.os.ReadFile(packageFilePath)
	if err != nil {
		log.Printf(
			"%s%s: %s",
			command.LogErrorPrefix,
			"Can't open the package.json file",
			err,
		)
		return 1
	}

	var packageJSON nodejs.PackageJSON
	err = json.Unmarshal(packageFileContents, &packageJSON)
	if err != nil {
		log.Printf(
			"%s%s: %s",
			command.LogErrorPrefix,
			"Failed to decode the package.json file",
			err,
		)
		return 1
	}

	c.packageJSON = packageJSON

	log.Printf(
		"%s%s: %s",
		command.LogSuccessPrefix,
		"Read package.json",
		packageJSON,
	)

	return 0
}
