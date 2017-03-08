package cmd

import (
	"fuse/pkg/kubectl"
	"github.com/spf13/cobra"
	"os"
)

var (
	// RootCmd is base command
	RootCmd = &cobra.Command{
		Use:   "fuse",
		Short: "Kubernetes deploy and maintenance tool",
		Long:  `Kubernetes deploy and maintenance tool, great for CI/CD environments`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if context != "" {
				os.Setenv(kubectl.ClusterContextEnv, context)
			}
		},
	}

	// global flag for current cluster context
	context = ""
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&context, "context", "c", "", "Override CLUSTER_CONTEXT defined in environment (default \"\")")
}
