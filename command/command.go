package command

import (
	"github.com/ttacon/chalk"
)

var (
	LogErrorPrefix   = chalk.Red.Color("[ERROR]  ")
	LogSuccessPrefix = chalk.Green.Color("[SUCCESS]  ")
	LogInfoPrefix    = chalk.Yellow.Color("[INFO]  ")
)

// BaseFlags defines command flags that all commands share
type BaseFlags struct {
	Help bool
}
