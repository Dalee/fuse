package kubectl

import (
	"github.com/stretchr/testify/assert"
	"reflect"
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
			Name:       "test-deployment",
			Namespace:  "default",
			UID:        "3eb259fc-bc6f-11e6-a342-005056ba5444",
			Generation: 12,
			Labels: map[string]string{
				"sample-label1": "example1",
			},
		},
		Spec: resourceSpec{
			Replicas: 1,
			Template: resourceTemplate{
				Metadata: resourceMetadataSpec{
					Labels: map[string]string{
						"project":       "test-project-2",
						"project_build": "12",
					},
				},
			},
			Strategy: resourceStrategy{
				Type: strategyTypeRollingUpdate,
				RollingUpdate: resourceStrategyRolling{
					MaxUnavailable: 0,
				},
			},
		},
		Status: resourceStatus{
			UnavailableReplicas: 1,
			ObservedGeneration:  0,
		},
	}

	assert.Equal(t, KindDeployment, d.GetKind())
	assert.Equal(t, "test-deployment", d.GetName())
	assert.Equal(t, "3eb259fc-bc6f-11e6-a342-005056ba5444", d.GetUUID())
	assert.Equal(t, "default/test-deployment", d.GetKey())
	assert.True(t, reflect.DeepEqual([]string{"sample-label1=example1"}, d.GetSelector()))
	assert.True(t, reflect.DeepEqual([]string{"project=test-project-2", "project_build=12"}, d.GetPodSelector()))
	assert.False(t, d.IsReady())
	assert.Equal(t, "Ready: false, Generation: meta=12 observed=0, Replicas: s=1, u=0, a=0, na=1", d.GetStatusString())

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

func TestDeployment_IsReady(t *testing.T) {
	// initial deployment, apply command issued
	d := Deployment{
		Kind: "Deployment",
		Metadata: resourceMetadata{
			Name:       "test-deployment",
			Namespace:  "default",
			Generation: 12,
		},
		Spec: resourceSpec{
			Replicas: 1,
			Strategy: resourceStrategy{
				Type: strategyTypeRollingUpdate,
				RollingUpdate: resourceStrategyRolling{
					MaxUnavailable: 0,
				},
			},
		},
		Status: resourceStatus{
			UpdatedReplicas:     0,
			AvailableReplicas:   0,
			UnavailableReplicas: 0,
			ObservedGeneration:  11,
		},
	}

	// initial check
	assert.False(t, d.IsReady())
	assert.Equal(t, "Ready: false, Generation: meta=12 observed=11, Replicas: s=1, u=0, a=0, na=0", d.GetStatusString())
	assert.Empty(t, d.GetPodSelector()) // since no labels defined in Spec.Metadata

	// 1) k8s started rollout
	d.Status.ObservedGeneration = 12
	d.Status.UpdatedReplicas = 0
	d.Status.AvailableReplicas = 0
	d.Status.UnavailableReplicas = 1
	assert.False(t, d.IsReady())
	assert.Equal(t, "Ready: false, Generation: meta=12 observed=12, Replicas: s=1, u=0, a=0, na=1", d.GetStatusString())

	// 2) k8s changed updated replica count
	d.Status.UpdatedReplicas = 1
	assert.False(t, d.IsReady())
	assert.Equal(t, "Ready: false, Generation: meta=12 observed=12, Replicas: s=1, u=1, a=0, na=1", d.GetStatusString())

	// 3) k8s changed available replica count
	d.Status.AvailableReplicas = 1
	assert.False(t, d.IsReady())
	assert.Equal(t, "Ready: false, Generation: meta=12 observed=12, Replicas: s=1, u=1, a=1, na=1", d.GetStatusString())

	// 4) finally, k8s changed unavailable replica count
	d.Status.UnavailableReplicas = 0
	assert.True(t, d.IsReady())
	assert.Equal(t, "Ready: true, Generation: meta=12 observed=12, Replicas: s=1, u=1, a=1, na=0", d.GetStatusString())

	// 5) rollout done..
}
