package main

import (
	"log"
	"os"
	"github.com/mitchellh/cli"
	"github.com/bosgood/dep-get/command/fetch"
	"github.com/bosgood/dep-get/command/archive"
	"github.com/bosgood/dep-get/command/install"
)

func main() {
	c := cli.NewCLI("dep-get", "0.0.1")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"fetch":   fetch.NewFetchCommand,
		"archive": archive.NewArchiveCommand,
		"install": install.NewInstallCommand,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
