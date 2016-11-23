package main

import (
	"os"
	"fmt"
	"time"

	"io/ioutil"
	"path/filepath"
	"os/exec"

	"fuse/lib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fuse file.yml")
		os.Exit(1)
	}

	filename := os.Args[1]
	filename, _ = filepath.Abs(filename)

	fmt.Printf("==> Using file: %s\n", filename)
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	typeList, err := lib.ParseYaml(yamlData)
	if err != nil {
		panic(err)
	}

	// fetching all deployments
	for _, def := range typeList {
		def.UpdateInfo(false, false)
	}

	// updating cluster
	cmd := exec.Command("kubectl", "apply", "-f", filename)
	applyResult, err := lib.RunCmd(cmd)
	fmt.Printf("==> Response from kubectl:\n%s\n", string(applyResult[:]))
	if err != nil {
		panic(err)
	}

	//
	// deploy loop
	//
	expiredAt := time.Now()
	expiredAt = expiredAt.Add(120 * time.Second) // set it to future

	isOk := true
	for {
		fmt.Println("==> ZzzZzzZzz...")
		time.Sleep(5 * time.Second)

		for _, def := range typeList {
			upd, err := def.UpdateInfo(true, true)
			if err != nil {
				continue
			}

			// should wait until new generation is deployed
			if upd.Status.ObservedGeneration == def.Status.ObservedGeneration {
				fmt.Println("==> Notice: generation is not changed, retrying..")
				continue
			}

			// every replica is available?
			if upd.Status.UnavailableReplicas == 0 {
				fmt.Println("==> Notice: no unavailable replicas found, assuming ok")
				def.Deployed = true
				break
			}

			fmt.Printf("==> Still unavailable: %d\n", upd.Status.UnavailableReplicas)
		}

		isOk = true // reset flag, to make sure every deployment is deployed
		for _, def := range typeList {
			isOk = isOk && def.Deployed
		}

		// if it's ok, break current loop
		if isOk {
			fmt.Println("==> Success: All deployments marked as ok..")
			break
		}

		// check for expire hit
		currentTime := time.Now()
		if currentTime.After(expiredAt) {
			fmt.Println("==> Failure: timeout reached, marking deployment as broken")
			break
		}
	}

	//
	// deploy or rollback?
	//
	if isOk == false {
		fmt.Println("==> Error: deploy failed, rolling back deployments..")
		for _, def := range typeList {
			cmd := exec.Command(
				"kubectl",
				"rollout",
				"undo",
				fmt.Sprintf("%s/%s", def.Kind, def.Metadata.Name),
			)

			fmt.Printf("==> Rolling back: %s/%s\n", def.Kind, def.Metadata.Name)
			if _, err := lib.RunCmd(cmd); err != nil {
				fmt.Printf("==> Error: %#v", err)
			}
		}
		os.Exit(127)
	} else {
		fmt.Println("==> Success: deploy sucessfull")
	}
}

