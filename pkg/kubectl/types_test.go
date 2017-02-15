package kubectl

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestReplicaSet_Types(t *testing.T) {
	rs := ReplicaSet{
		Kind: "ReplicaSet",
		Metadata: kubeTypeMetadata{
			Name: "test-replicaset-42",
			Namespace: "default",
		},
		Spec: kubeTypeResourceSpec{
			Template: kubeTypeTemplate{
				Spec: kubeTypeSpec{
					Containers: []kubeContainerSpec{
						{
							Image: "registry.example.com/example/repo:1",
						},
					},
				},
			},
		},
	}

	assert.Equal(t, "test-replicaset-42", rs.GetName())
	assert.Equal(t, "replicaset", rs.GetKind())
	assert.Equal(t, []string{"registry.example.com/example/repo:1"}, rs.GetImages())
}

func TestDeployment_Types(t *testing.T) {
	d := Deployment{
		Kind: "Deployment",
		Metadata: kubeTypeMetadata{
			Name: "test-deployment",
			Namespace: "default",
		},
	}

	assert.Equal(t, "test-deployment", d.GetName())
	assert.Equal(t, "deployment", d.GetKind())
}

func TestKubeResourceList_Types(t *testing.T) {
	rl := kubeResourceList{
		Kind: "List",
	}

	assert.Equal(t, "list", rl.GetKind())
}
