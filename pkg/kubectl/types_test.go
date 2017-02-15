package kubectl

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestReplicaSet_GetImages(t *testing.T) {
	rs := ReplicaSet{
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

	assert.Equal(t, []string{"registry.example.com/example/repo:1"}, rs.GetImages())
}
