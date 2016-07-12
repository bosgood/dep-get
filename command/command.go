package command

import (
// "github.com/mitchellh/cli"
)

type Command struct {
	SynopsisText string
	RetVal       int
}

func (c *Command) Help() string {
	return "I'm super helpful"
}

func (c *Command) Run(args []string) int {
	return c.RetVal
}

func (c *Command) Synopsis() string {
	return c.SynopsisText
}
