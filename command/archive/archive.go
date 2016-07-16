package archive

import (
	"encoding/json"
	"github.com/bosgood/dep-get/nodejs"
	"github.com/mitchellh/cli"
	"github.com/ttacon/chalk"
	"io/ioutil"
	"log"
	"os"
	"path"
)

var (
	logErrorPrefix   = chalk.Red.Color("[ERROR]")
	logSuccessPrefix = chalk.Green.Color("[SUCCESS]")
)

type command struct{}

// NewArchiveCommand is used to generate a command object
// which orchestrates package dependency discovery
// and archiving to S3
func NewArchiveCommand() (cli.Command, error) {
	cmd := &command{}
	return cmd, nil
}

func (c *command) Synopsis() string {
	return "Archives application dependencies"
}

func (c *command) Help() string {
	return "I'm super helpful"
}

func (c *command) Run(args []string) int {
	var dirPath string

	if len(args) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			log.Printf("Can't read current directory: %s", err)
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
			"%s %s: %s",
			logErrorPrefix,
			"Can't open the package.json file",
			err,
		)
		return 1
	}

	var packageJSON nodejs.PackageJSON
	err = json.Unmarshal(packageFileContents, &packageJSON)
	if err != nil {
		log.Printf(
			"%s %s: %s",
			logErrorPrefix,
			"Failed to decode the package.json file",
			err,
		)
		return 1
	}

	log.Printf(
		"%s %s: %s",
		logSuccessPrefix,
		"Read package.json",
		packageJSON,
	)

	return 0
}
