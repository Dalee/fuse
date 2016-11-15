package lib

import (
	"os/exec"
	"fmt"
)

type KubeTypeMetadata struct {
	Name string `yaml:"name"`
}

type KubeTypeStatus struct {
	AvailableReplicas   int `yaml:"availableReplicas"`
	ObservedGeneration  int `yaml:"observedGeneration"`
	Replicas            int `yaml:"replicas"`
	UpdatedReplicas     int `yaml:"updatedReplicas"`
	UnavailableReplicas int `yaml:"unavailableReplicas"`
}

type KubeType struct {
	Kind     string `yaml:"kind"`
	Metadata KubeTypeMetadata `yaml:"metadata"`
	Status   KubeTypeStatus `yaml:"status"`
	Deployed bool
}

func (def *KubeType) UpdateInfo(new bool) (*KubeType, error) {
	cmd := exec.Command(
		"kubectl",
		"get",
		fmt.Sprintf("%s/%s", def.Kind, def.Metadata.Name),
		"-o",
		"yaml",
	)

	data, err := RunCmd(cmd)
	if err == nil {
		typeList, err := ParseYaml(data)
		if err != nil {
			panic(err)
		}

		if len(typeList) == 0 {
			panic("No information fetched!")
		}

		right := typeList[0]
		if new == true {
			return right, nil
		}

		fmt.Println("==> New is not requested, updating current type")
		def.Status.AvailableReplicas = right.Status.AvailableReplicas
		def.Status.ObservedGeneration = right.Status.ObservedGeneration
		def.Status.Replicas = right.Status.Replicas
		def.Status.UpdatedReplicas = right.Status.UpdatedReplicas
		def.Status.UnavailableReplicas = right.Status.UnavailableReplicas

	} else {
		if new == true {
			fmt.Printf(
				"==> Error: can't fetch new data about %s/%s",
				def.Kind,
				def.Metadata.Name,
			)
			return nil, err
		}

		fmt.Printf("==> Notice: %s/%s is not registered\n", def.Kind, def.Metadata.Name)
		def.Status.ObservedGeneration = 1
		def.Status.UnavailableReplicas = 0
	}

	return def, nil
}
