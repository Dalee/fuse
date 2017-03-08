package kubectl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNamespace_Type(t *testing.T) {
	ns := Namespace{
		Kind: "Namespace",
		Metadata: resourceMetadata{
			Name: "default",
		},
	}

	assert.Equal(t, "default", ns.GetName())
	assert.Equal(t, KindNamespace, ns.GetKind())

	_, err := ns.ToDeployment()
	assert.Error(t, err)
}

func TestReplicaSet_Type(t *testing.T) {
	rs := ReplicaSet{
		Kind: "ReplicaSet",
		Metadata: resourceMetadata{
			Name:      "test-replicaset-42",
			Namespace: "default",
		},
		Spec: resourceSpec{
			Template: resourceTemplate{
				Spec: resourceContainerSpec{
					Containers: []resourceContainer{
						{
							Image: "registry.example.com/example/repo:1",
						},
					},
				},
			},
		},
	}

	assert.Equal(t, "test-replicaset-42", rs.GetName())
	assert.Equal(t, KindReplicaSet, rs.GetKind())
	assert.Equal(t, []string{"registry.example.com/example/repo:1"}, rs.GetImages())

	_, err := rs.ToDeployment()
	assert.Error(t, err)
}

func TestDeployment_Type(t *testing.T) {
	d := Deployment{
		Kind: "Deployment",
		Metadata: resourceMetadata{
			Name:      "test-deployment",
			Namespace: "default",
			UID:       "3eb259fc-bc6f-11e6-a342-005056ba5444",
			Labels: map[string]string{
				"sample-label1": "example1",
			},
		},
	}

	assert.Equal(t, KindDeployment, d.GetKind())
	assert.Equal(t, "test-deployment", d.GetName())
	assert.Equal(t, "3eb259fc-bc6f-11e6-a342-005056ba5444", d.GetUUID())
	assert.Equal(t, "default/test-deployment", d.GetKey())
	assert.Equal(t, []string{"sample-label1=example1"}, d.GetSelector())

	converted, err := d.ToDeployment()
	assert.Nil(t, err)
	assert.Equal(t, d, *converted)
}

func TestPod_Type(t *testing.T) {
	p := Pod{
		Kind: "Pod",
		Metadata: resourceMetadata{
			Name:      "test-pod",
			Namespace: "system",
		},
	}

	assert.Equal(t, "test-pod", p.GetName())
	assert.Equal(t, KindPod, p.GetKind())
	assert.Equal(t, "system/test-pod", p.GetKey())

	_, err := p.ToDeployment()
	assert.Error(t, err)
}

func TestKubeResourceList_GetKind(t *testing.T) {
	rl := kubeResourceList{
		Kind: "List",
	}

	assert.Equal(t, KindList, rl.GetKind())
}

func TestKubeResourceList_GetItems(t *testing.T) {
	rl := kubeResourceList{
		Kind: "List",
		Items: []kubeResource{
			{
				Kind: "Deployment",
				Metadata: resourceMetadata{
					Name: "example",
				},
			},
		},
	}

	items := rl.GetItems()
	assert.Len(t, items, 1)
	assert.Equal(t, KindDeployment, items[0].GetKind())
	assert.Equal(t, "example", items[0].GetName())

	_, err := items[0].ToDeployment()
	assert.Error(t, err)
}

func TestResourceList_FilteredByKind(t *testing.T) {
	rl := ResourceList{
		&Pod{
			Kind: "Pod",
			Metadata: resourceMetadata{
				Name:      "test-pod",
				Namespace: "system",
			},
		},
		&Deployment{
			Kind: "Deployment",
			Metadata: resourceMetadata{
				Name:      "test-deployment",
				Namespace: "default",
			},
		},
	}

	items := rl.FilteredByKind(KindDeployment)

	assert.Len(t, items, 1)
	assert.Equal(t, KindDeployment, items[0].GetKind())
}
