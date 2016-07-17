package archive

import (
	"encoding/json"
	"github.com/bosgood/dep-get/command"
	"github.com/bosgood/dep-get/nodejs"
	"github.com/bosgood/dep-get/lib/fs"
	"github.com/mitchellh/cli"
	"io/ioutil"
	"log"
	"path"
)

var os fs.FileSystem = fs.OSFS{}

type archiveCommand struct{}

// NewArchiveCommand is used to generate a command object
// which orchestrates package dependency discovery
// and archiving to S3
func NewArchiveCommand() (cli.Command, error) {
	cmd := &archiveCommand{}
	return cmd, nil
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
		cwd, err := os.Getwd()
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
		dirPath = args[0]
	}

	packageFilePath := path.Join(dirPath, "package.json")
	packageFileContents, err := ioutil.ReadFile(packageFilePath)
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

	log.Printf(
		"%s%s: %s",
		command.LogSuccessPrefix,
		"Read package.json",
		packageJSON,
	)

	return 0
}
