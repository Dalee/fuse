package main

import (
	"os"
	"fmt"
	"time"
	"errors"

	"io/ioutil"
	"path/filepath"

	"fuse/lib"
)

var clusterContext string
var deployFilename string

//
// get cluster context in which we are working
// and prepare passed filename to be read
//
func loadDefaults([]string) {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fuse file.yml")
		os.Exit(1)
	}

	deployFilename = os.Args[1]
	deployFilename, _ = filepath.Abs(deployFilename)

	if deployFilename == "" {
		panic(errors.New("Empty filename, did you forget about passing kubernetes.yml?"))
	}

	clusterContext = os.Getenv("CLUSTER_CONTEXT")
}

//
// Entry point
//
func main() {
	loadDefaults(os.Args)

	//
	// parsing yml file and fetch information
	// about deployments described in file
	//
	fmt.Printf("==> Using file: %s\n", deployFilename)
	yamlData, err := ioutil.ReadFile(deployFilename)
	if err != nil {
		panic(err)
	}

	typeList, err := lib.ParseYaml(string(yamlData[:]))
	if err != nil {
		panic(err)
	}

	for _, def := range typeList {
		def.UpdateInfo(clusterContext, false, false)
	}

	//
	// update cluster with new yml definition
	//
	cmd := lib.CommandFactory(
		clusterContext,
		[]string{
			"apply",
			"-f",
			deployFilename,
		},
	)

	output, success := lib.RunCmd(cmd)
	fmt.Printf("==> Response from kubectl:\n%s\n", output)
	if success == false {
		os.Exit(127)
	}

	//
	// deploy loop
	//
	expiredAt := time.Now()
	expiredAt = expiredAt.Add(120 * time.Second) // set it to future

	isOk := true
	for {
		//
		// give k8s some time after apply to create new replica sets
		// for each deployment
		//
		fmt.Println("==> ZzzZzzZzz...")
		time.Sleep(5 * time.Second)

		for _, def := range typeList {
			//
			// update current deployment info
			//
			upd, err := def.UpdateInfo(clusterContext, true, true)
			if err != nil {
				continue
			}

			//
			// is generation changed? if not, there can be two options:
			// 1) we are trying to deploy same generation (which should'nt happen)
			// 2) deployment is not yet created, need to wait some time
			//
			if upd.Status.ObservedGeneration == def.Status.ObservedGeneration {
				fmt.Println("==> Notice: generation is not changed yet, skipping...")
				fmt.Println("==> Notice: probably you try to re-deploy same generation (which is error)")
				continue
			}

			//
			// does deployment have any unavailable replicas?
			// if it does, deployment is not yet deployed..
			//
			if upd.Status.UnavailableReplicas == 0 {
				fmt.Println("==> Notice: no unavailable replicas found, assuming ok")
				def.Deployed = true
				continue
			}

			fmt.Printf("==> Unavailable replica sets: %d, waiting...\n", upd.Status.UnavailableReplicas)
		}

		//
		// checking that every deployment defined in yml file
		// is deployed, and no errors registered during loop
		//
		isOk = true
		for _, def := range typeList {
			isOk = isOk && def.Deployed
		}

		//
		// if every deployment defined in yml files
		// is deployed, then, our job is done.
		//
		if isOk {
			fmt.Println("==> Success: All deployments marked as ok..")
			break
		}

		//
		// check for expire
		// if expire time is reached, and nothing is deployed
		// should undo all deployments and revert them to
		// previous replica version
		//
		currentTime := time.Now()
		if currentTime.After(expiredAt) {
			fmt.Println("==> Failure: timeout reached, marking deployment as broken")
			break
		}
	}

	//
	// So, main cycle is finished, does all deployments are deployed?
	// if not, should rollback, otherwise, everything is ok
	//
	if isOk == false {
		fmt.Println("==> Error: deploy failed, rolling back deployments..")
		for _, def := range typeList {
			cmd := lib.CommandFactory(
				clusterContext,
				[]string{
					"rollout",
					"undo",
					fmt.Sprintf("%s/%s", def.Kind, def.Metadata.Name),
				},
			)

			fmt.Printf("==> Rolling back: %s/%s\n", def.Kind, def.Metadata.Name)
			output, _ := lib.RunCmd(cmd)
			fmt.Printf("==> Response from kubectl: %s\n", output)
		}
		os.Exit(127)

	} else {
		fmt.Println("==> Success: deploy successfull")
		os.Exit(0)
	}
}

