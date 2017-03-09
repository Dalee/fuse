package cmd

import (
	"errors"
	"fmt"

	"fuse/pkg/kubectl"
	"fuse/pkg/reference"

	"github.com/Dalee/hitman/pkg/registry"
	"github.com/spf13/cobra"
	"time"
)

var (
	// command itself
	garbageCollectCmd = &cobra.Command{
		Use:   "garbage-collect",
		Short: "Remove tags from registry not registered within Kubernetes ReplicaSet",
		Long:  ``,
		RunE:  garbageCollectCmdHandler,
	}

	// Flags
	dryRunFlag        = false
	ignoreMissingFlag = false
	registryURLFlag   = ""
	namespaceFlag     = ""

	// Docker Distribution client
	hitmanClient *registry.Registry
)

// register all flags
func init() {
	garbageCollectCmd.Flags().StringVarP(&namespaceFlag, "namespace", "n", "default", "Kubernetes namespace to use")
	garbageCollectCmd.Flags().BoolVarP(&dryRunFlag, "dry-run", "d", false, "Do not execute destructive actions (default \"false\")")
	garbageCollectCmd.Flags().StringVarP(&registryURLFlag, "registry-url", "r", "", "Registry URL (e.g. \"https://registry.example.com:5000/\")")
	garbageCollectCmd.Flags().BoolVarP(&ignoreMissingFlag, "ignore-missing", "i", false, "Skip missing images in Registry (default \"false\")")
	RootCmd.AddCommand(garbageCollectCmd)
}

// get garbage from docker distribution, list of repositories from kubernetes replica sets
func getGarbage() (*reference.GarbageDetectInfo, error) {
	fmt.Println("==> Fetching repository info...")
	resourceList, err := kubectl.CommandReplicaSetList(namespaceFlag).RunAndParse()
	if err != nil {
		return nil, err
	}

	// collect all images defined in every ReplicaSet
	cnList := make([]string, 0)
	rsList := resourceList.ToReplicaSetList()

	for _, rs := range rsList {
		for _, image := range rs.GetImages() {
			cnList = append(cnList, image)
		}
	}

	// perform detection
	garbageInfo, err := reference.DetectGarbage(cnList, hitmanClient, ignoreMissingFlag)
	if err != nil {
		return nil, err
	}

	return garbageInfo, nil
}

// printing report
func printGarbage(garbageInfo *reference.GarbageDetectInfo) error {
	fmt.Printf("==> Found %d repositories\n", len(garbageInfo.Items))
	for _, item := range garbageInfo.Items {
		fmt.Printf("===> Repository: %s\n", item.Repository)
		fmt.Printf("===> Deployed tags: %v\n", item.DeployedTagList)
		fmt.Printf("===> Garbage tags: %v\n", item.GarbageTagList)
	}
	return nil
}

// delete garbage from docker distribution
func deleteGarbage(garbageInfo *reference.GarbageDetectInfo) error {
	fmt.Println("==> Clean up")
	for _, item := range garbageInfo.Items {

		fmt.Printf("===> Clearing up repository: %s\n", item.Repository)
		for _, digest := range item.GarbageDigestList {

			err := hitmanClient.DeleteImageDigest(item.Repository, digest)
			if err != nil {
				return err
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

// command handler
func garbageCollectCmdHandler(cmd *cobra.Command, args []string) error {
	var err error

	// build registry
	if registryURLFlag == "" {
		return errors.New("registry-url is a mandatory parameter")
	}

	hitmanClient = registry.New(registryURLFlag)
	if hitmanClient.IsValidURL() == false {
		return fmt.Errorf("Request to %s/v2/ failed, is URL pointed to Docker Registry?", registryURLFlag)
	}

	// detect garbage
	garbageInfo, err := getGarbage()
	if err != nil {
		return err
	}

	// print report
	err = printGarbage(garbageInfo)
	if err != nil {
		return err
	}

	// clearing up if not dry-run
	if dryRunFlag == false {
		err = deleteGarbage(garbageInfo)
		if err != nil {
			return err
		}
	}

	return nil
}
