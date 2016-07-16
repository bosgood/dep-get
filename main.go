package main

import (
	"log"
	"os"
	"github.com/mitchellh/cli"
	"github.com/bosgood/dep-get/command/archive"
)

func main() {
	c := cli.NewCLI("dep-get", "0.0.1")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"archive": archive.NewArchiveCommand,
		// "install": NewInstallCommand,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
