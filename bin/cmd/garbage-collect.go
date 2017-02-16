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

func init() {
	gcCommand.Flags().Bool("dry-run", false, "Do not try to execute destructive actions (default \"false\")")
	gcCommand.Flags().String("registry-url", "", "Registry URL to use (e.g. https://example.com:5000/)")
	gcCommand.Flags().String("namespace", "default", "Namespace to fetch ReplicaSet")
	gcCommand.Flags().Bool("ignore-missing", false, "Ignore/Skip missing images in Registry (default \"false\")")
	RootCmd.AddCommand(gcCommand)
}

var gcCommand = &cobra.Command{
	Use:   "garbage-collect",
	Short: "Remove tags from registry not registered within any Kubernetes ReplicaSet",
	Long:  ``,
	RunE:  garbageCollectHandler,
}

//
//
func garbageCollectHandler(cmd *cobra.Command, args []string) error {
	registryURL, err := cmd.Flags().GetString("registry-url")
	if err != nil {
		return err
	}

	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	// get dry-run flag
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	// constructing registry client
	if registryURL == "" {
		return errors.New("registry-url is a mandatory parameter")
	}
	registryClient := registry.New(registryURL)
	if registryClient.IsValidUrl() == false {
		return fmt.Errorf("Request to %s/v2/ failed, is URL pointed to Docker Registry?", registryURL)
	}

	// get all ReplicaSets available in given namespace
	fmt.Printf("==> Using namespace: %s\n", namespace)
	replicaList, err := kubectl.CommandReplicaSetList(namespace).RunAndParse()
	if err != nil {
		return err
	}

	// collect all images defined in every ReplicaSet
	cnList := make([]string, 0)
	rsList := kubectl.ToReplicaSetList(replicaList)

	fmt.Printf("==> Found: %d ReplicaSets\n", len(rsList))
	for _, rs := range rsList {
		for _, image := range rs.GetImages() {
			cnList = append(cnList, image)
		}
	}

	// perform detection
	fmt.Printf("==> Detecting garbage, dry-run is: %v\n", dryRun)
	garbageInfo, err := reference.DetectGarbage(cnList, registryClient, false)
	if err != nil {
		return err
	}

	// print report
	for _, item := range garbageInfo.Items {
		fmt.Printf("==> Repository: %s\n", item.Repository)
		fmt.Printf("Deployed: %v\n", item.DeployedTagList)
		fmt.Printf("Garbage: %v\n", item.GarbageTagList)
		for _, digest := range item.GarbageDigestList {
			fmt.Printf("%s\n", digest)
		}
	}

	// c'mon, let's do it!
	if dryRun == false {
		fmt.Println("==> Dry run is not set, performing deletion..")
		for _, item := range garbageInfo.Items {

			fmt.Printf("==> %s\n", item.Repository)
			for _, digest := range item.GarbageDigestList {

				// see ya..
				err := registryClient.DeleteImageDigest(item.Repository, digest)
				if err != nil {
					return err
				}

				// well.. we'll miss you of course
				fmt.Printf("Deleted: %s\n", digest)

				// do not throttle registry
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	fmt.Println("Done, have a nice day!")
	return nil
}
