package cmd

import (
	"errors"
	"fmt"
	"fuse/pkg/kubectl"
	"github.com/spf13/cobra"
)

var (
	// command itself
	execCmd = &cobra.Command{
		Use:   "exec",
		Short: "Execute command in pods by selector",
		Long:  ``,
		RunE:  execCmdHandler,
	}

	// Flags
	execCommand         = ""
	deploymentSelectors = make([]string, 0)
)

// register all flags
func init() {
	execCmd.Flags().StringSliceVar(&deploymentSelectors, "deployments", []string{}, "Deployment selector (e.g. app=myapp)")
	execCmd.Flags().StringVar(&execCommand, "command", "", "Command to execute")
	RootCmd.AddCommand(execCmd)
}

// command handler
func execCmdHandler(cmd *cobra.Command, args []string) error {
	if execCommand == "" {
		return errors.New("No command provided")
	}

	if len(deploymentSelectors) == 0 {
		// get deployment by selector
		return errors.New("No deployment selectors provided")
	}

	// get deployment list by selector
	deploymentList := make([]kubectl.Deployment, 0)
	for _, s := range deploymentSelectors {
		resourceList, err := kubectl.CommandDeploymentListBySelector(namespaceFlag, []string{s}).RunAndParse()
		if err != nil {
			return err
		}

		dl := resourceList.ToDeploymentList()
		deploymentList = append(deploymentList, dl...)
	}

	podList := make([]kubectl.Pod, 0)

	// for each deployment, find all pods
	for _, d := range deploymentList {
		podResourceList, err := kubectl.CommandPodListBySelector(namespaceFlag, d.GetPodSelector()).RunAndParse()
		if err != nil {
			return err
		}

		pl := podResourceList.ToPodList()
		podList = append(podList, pl...)
	}

	// for each container in pod, exec provided command
	for _, pod := range podList {
		for _, c := range pod.Spec.Containers {
			podName := pod.GetName()

			stdout, err := kubectl.CommandExec(namespaceFlag, podName, c.Name, execCommand).RunPlain()
			fmt.Printf("===> Pod: %s, Container: %s:\n", pod.GetKey(), c.Name)
			fmt.Println(string(stdout))

			if err != nil {
				return err
			}
		}
	}

	return nil
}
