package cmd

import (
	"errors"
	"fmt"
	"fuse/pkg/kubectl"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"time"
)

var (
	applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Perform safe deployment to Kubernetes cluster",
		Long:  `Apply new configuration to Kubernetes cluster and monitor release delivery`,
		RunE:  applyCmdHandler,
	}

	configurationYaml  = ""
	clusterTimeoutFlag = 0 // in seconds
	clusterTimeout     time.Duration
)

func init() {
	applyCmd.Flags().StringVarP(&configurationYaml, "configuration", "f", "", "Release configuration yaml file, mandatory")
	applyCmd.Flags().IntVarP(&clusterTimeoutFlag, "release-timeout", "t", 0, "Deploy timeout in seconds, override CLUSTER_RELEASE_TIMEOUT environment")
	RootCmd.AddCommand(applyCmd)
}

// Load and check configuration (yaml file), parse and return all
// deployments defined in configuration file
func initDeploy() (*[]kubectl.Deployment, error) {
	if configurationYaml == "" {
		return nil, errors.New("mandatory configuration parameter is not provided")
	}

	// parsing provided configuration
	fmt.Printf("==> Parsing file: %s\n", configurationYaml)
	fullResourceList, err := kubectl.ParseLocalFile(configurationYaml)
	if err != nil {
		return nil, err
	}

	// filtering deployments, if no deployments found it's an error
	newConfigurationList := fullResourceList.FilteredByKind(kubectl.KindDeployment).ToDeploymentList()
	if len(newConfigurationList) == 0 {
		return nil, errors.New("no Deployment resources found in configuration")
	}

	// loading requested timeout
	clusterTimeout = time.Second * 120 // default is 2m
	if clusterTimeoutFlag == 0 {
		clusterTimeoutEnv := os.Getenv(kubectl.ClusterReleaseTimeoutEnv)
		if parsed, err := strconv.Atoi(clusterTimeoutEnv); err == nil && parsed > 0 {
			clusterTimeout = time.Second * parsed
		}
	}
	if clusterTimeoutFlag > 0 {
		clusterTimeout = time.Second * clusterTimeoutFlag
	}

	return &newConfigurationList, nil
}

// Start deploy process / apply new configuration to cluster and display output
func applyDeploy(list *[]kubectl.Deployment) error {
	// printing extracted resources
	fmt.Println("==> Deployments scheduled for update:")
	for _, resource := range *list {
		fmt.Printf("===> Deployment: %s/%s\n", resource.GetKind(), resource.GetName())
	}

	fmt.Println("==> Applying configuration...")
	stdout, err := kubectl.CommandApply(configurationYaml).RunPlain()
	if err != nil {
		return err
	}

	fmt.Println("==> Cluster apply response:")
	fmt.Println(string(stdout))
	return nil
}

// Monitor configuration delivery, all unavailable replicas of each deployment
// should be 0. Wait until timeout.
func monitorDeploy(list *[]kubectl.Deployment) (bool, error) {
	willExpireAt := time.Now().Add(clusterTimeout * time.Second)
	fmt.Printf("==> Starting release delivery monitoring, timeout is %d seconds\n", clusterTimeout)

	for {
		// make initial delay..
		time.Sleep(5 * time.Second)
		updatedList := make([]kubectl.Deployment, 0)

		// collect cluster configuration
		for _, cfg := range *list {
			cmd := kubectl.CommandDeploymentInfo(cfg.GetNamespace(), cfg.GetName())
			resource, _ := cmd.RunAndParseFirst()
			if resource == nil {
				continue
			}

			deployment, _ := resource.ToDeployment()
			if deployment == nil {
				continue
			}

			updatedList = append(updatedList, deployment)
		}

		// check timeout
		if time.Now().After(willExpireAt) {
			fmt.Println("===> Timeout reached, aborting deploy...")
			break
		}

		// do we get all deployment info back from cluster?
		if len(list) != len(updatedList) {
			fmt.Println("===> Not all deployments registered in cluster, waiting...")
			continue
		}

		// is every deployment successfully delivered?
		isDelivered := true
		for _, cfg := range updatedList {
			fmt.Printf("===> Deployment: %s, %s\n", cfg.GetKey(), cfg.GetStatusString())
			if !cfg.IsReady() {
				isDelivered = false
			}
		}

		// is every deployment delivered?
		if isDelivered {
			return true, nil
		}
	}

	return false, nil
}

// Finalize delivery process, either do nothing or display logs for each pod of each deployment
// in order to have information about broken delivery
func finalizeDeploy(list *[]kubectl.Deployment, isDeployed bool) error {

	// if it's not deployed, display logs by deployment selector
	fmt.Println("==> Fetching logs of failed pods...")
	for _, d := range *list {
		// ok, get list of pods connected to deployment
		resourceList, err := kubectl.CommandPodList(d.GetNamespace(), d.GetSelector()).RunAndParse()
		if err != nil {
			return err
		}

		// display logs from each pod
		podList := resourceList.FilteredByKind(kubectl.KindPod).ToPodList()
		for _, pod := range podList {
			stdout, err := kubectl.CommandPodLogs(pod.GetNamespace(), pod.GetName()).RunPlain()
			if err != nil {
				return err
			}

			fmt.Printf("===> Deployment: %s, Pod: %s - Logs:\n", d.GetKey(), pod.GetKey())
			fmt.Println(string(stdout))
		}
	}

	// if deploy successful do nothing..
	if isDeployed {
		fmt.Println("==> Deploy finished successfully")
		return nil
	}

	// error registered, if deployment has > 1 replica sets, rollback it
	fmt.Println("==> Deploy finished with errors, undoing configuration changes...")
	for _, d := range *list {
		// get list of replica sets connected to deployment
		resourceList, err := kubectl.CommandReplicaSetListBySelector(d.GetNamespace(), d.GetSelector()).RunAndParse()
		if err != nil {
			return err
		}

		// if more than one, we able to perform "rollout undo"
		if len(resourceList) > 1 {
			stdout, err := kubectl.CommandRollback(d.GetNamespace(), d.GetKind(), d.GetName()).RunPlain()
			if err != nil {
				return err
			}

			fmt.Printf("===> Deployment: %s - rolled back to previous release\n", d.GetKey())
			fmt.Println(string(stdout))

		} else {
			fmt.Printf("===> Deployment: %s - no rollback history available", d.GetKey())
		}
	}

	return nil
}

// command handler
func applyCmdHandler(cmd *cobra.Command, args []string) error {
	var list *[]kubectl.Deployment
	var err error
	var deployed bool

	// load and parse configuration
	if list, err = initDeploy(); err != nil {
		return err
	}

	// apply configuration / start deploy
	if err = applyDeploy(list); err != nil {
		return err
	}

	// monitor deploy
	if deployed, err = monitorDeploy(list); err != nil {
		return err
	}

	// finalize deploy
	if err = finalizeDeploy(list, deployed); err != nil {
		return err
	}

	// signalize to CI/CD about final status
	if deployed {
		os.Exit(0)
	} else {
		os.Exit(1)
	}

	return nil
}
