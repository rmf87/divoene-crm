package main

import (
	"os"

	"github.com/rmf87/divoene/cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
