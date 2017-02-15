package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd is the main root application
var RootCmd = &cobra.Command{
	Use:   "fuse",
	Short: "Kubernetes deploy and maintenance tool",
	Long:  `Kubernetes deploy and maintenance tool, great for CI/CD environments`,
}
