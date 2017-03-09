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
			// override ClusterContextEnv with provided flag
			if contextFlag != "" {
				os.Setenv(kubectl.ClusterContextEnv, contextFlag)
			}
		},
	}

	// global flag for current cluster context
	contextFlag = ""
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&contextFlag, "context", "c", "", "Override CLUSTER_CONTEXT defined in environment (default \"\")")
}
