package main

import (
	"github.com/Dalee/fuse/bin/cmd"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
