package main

import (
	"stroidex/internal/cli"
)

func main() {
	stroidokCLI := cli.NewCLI()
	if err := stroidokCLI.Execute(); err != nil {
		cli.PrintError(err)
	}
}