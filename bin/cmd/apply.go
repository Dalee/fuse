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
func initRollOut() (*[]kubectl.Deployment, error) {
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
		if timeoutEnv, err := strconv.Atoi(clusterTimeoutEnv); err == nil && timeoutEnv > 0 {
			clusterTimeout = time.Duration(timeoutEnv) * time.Second
		}
	}
	if clusterTimeoutFlag > 0 {
		clusterTimeout = time.Duration(clusterTimeoutFlag) * time.Second
	}

	return &newConfigurationList, nil
}

// Start deploy process / apply new configuration to cluster and display output
func applyRollOut(specList *[]kubectl.Deployment) error {
	fmt.Println("==> Deployments scheduled for update:")
	for _, resource := range *specList {
		fmt.Printf("===> %s\n", resource.GetKey())
	}

	fmt.Println("==> Applying configuration...")
	stdout, err := kubectl.CommandApply(configurationYaml).RunPlain()

	fmt.Println("==> Cluster apply response:")
	fmt.Println(string(stdout)) // in case of error, display output
	if err != nil {
		return err
	}

	return nil
}

// Monitor configuration delivery, all unavailable replicas of each deployment
// should be 0. Wait until timeout.
func monitorRollOut(specList *[]kubectl.Deployment) (bool, error) {
	willExpireAt := time.Now().Add(clusterTimeout * time.Second)
	fmt.Printf("==> Starting rollout monitoring, timeout is %d seconds\n", clusterTimeout)

	for {
		// make initial delay..
		time.Sleep(5 * time.Second)
		rolledList := make([]kubectl.Deployment, 0)

		// fetch updated deployment configuration
		for _, spec := range *specList {
			r, _ := kubectl.CommandDeploymentInfo(spec.GetNamespace(), spec.GetName()).RunAndParseFirst()
			if r == nil {
				continue
			}

			d, _ := r.ToDeployment()
			if d == nil {
				continue
			}

			rolledList = append(rolledList, *d)
		}

		// 1) timeout reached?
		if time.Now().After(willExpireAt) {
			fmt.Println("===> Timeout reached, aborting rollout...")
			break
		}

		// 2) every deployment defined in spec registered in cluster?
		if len(*specList) != len(rolledList) {
			fmt.Printf("===> Waiting for deployment registration, %d to go...\n", len(*specList)-len(rolledList))
			continue
		}

		// 3) every deployment is rolled out?
		isRolledOut := true
		for _, d := range rolledList {
			fmt.Printf("===> %s, %s\n", d.GetKey(), d.GetStatusString())
			if !d.IsReady() {
				isRolledOut = false
			}
		}

		// 4) if rolled out, stop..
		if isRolledOut {
			return true, nil
		}
	}

	// timeout reached, aborting...
	return false, nil
}

// Finalize delivery process, either do nothing or display logs for each pod of each deployment
// in order to have information about broken delivery
func finalizeRollOut(specList *[]kubectl.Deployment, isRolledOut bool) error {

	// display logs for each pod attached to deployment list
	fmt.Println("==> Fetching logs...")
	for _, d := range *specList {
		// get list of pods connected to deployment
		rlist, err := kubectl.CommandPodListBySelector(d.GetNamespace(), d.GetSelector()).RunAndParse()
		if err != nil {
			return err
		}

		// display logs for each pod
		plist := rlist.FilteredByKind(kubectl.KindPod).ToPodList()
		for _, pod := range plist {
			stdout, err := kubectl.CommandPodLogs(pod.GetNamespace(), pod.GetName()).RunPlain()
			fmt.Printf("===> Deployment: %s, Pod: %s:\n", d.GetKey(), pod.GetKey())
			fmt.Println(string(stdout))

			if err != nil {
				return err
			}
		}
	}

	// if deploy successful do nothing..
	if isRolledOut {
		fmt.Println("==> Rollout finished successfully")
		return nil
	}

	// error registered, if deployment has > 1 replica sets, rollback it
	fmt.Println("==> Rollout failed, starting undo process...")
	for _, d := range *specList {
		// get list of replica sets connected to deployment
		rlist, err := kubectl.CommandReplicaSetListBySelector(d.GetNamespace(), d.GetSelector()).RunAndParse()
		if err != nil {
			return err
		}

		// if deployment has previous configuration - perform "rollout undo"
		// otherwise - do nothing...
		if len(rlist) > 1 {
			stdout, err := kubectl.CommandRollback(d.GetNamespace(), d.GetKind(), d.GetName()).RunPlain()
			fmt.Printf("===> Deployment: %s - rolled back to previous release\n", d.GetKey())
			fmt.Println(string(stdout))
			if err != nil {
				return err
			}

		} else {
			fmt.Printf("===> Deployment: %s - no rollback history available", d.GetKey())
		}
	}

	return nil
}

// command handler
func applyCmdHandler(cmd *cobra.Command, args []string) error {
	var specList *[]kubectl.Deployment
	var err error
	var isRolledOut bool

	// load and parse configuration spec
	if specList, err = initRollOut(); err != nil {
		return err
	}

	// apply configuration / start rollout
	if err = applyRollOut(specList); err != nil {
		return err
	}

	// monitor rollout
	if isRolledOut, err = monitorRollOut(specList); err != nil {
		return err
	}

	// finalize deploy
	if err = finalizeRollOut(specList, isRolledOut); err != nil {
		return err
	}

	// signalize to CI/CD about final status
	if isRolledOut {
		os.Exit(0)
	} else {
		os.Exit(1)
	}

	return nil
}
