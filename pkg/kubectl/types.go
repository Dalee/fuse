package kubectl

import (
	"strings"
)

type (
	kubeTypeMetadata struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	}

	kubeTypeStatus struct {
		AvailableReplicas   int `yaml:"availableReplicas"`   // status: total amount of available instances
		ObservedGeneration  int `yaml:"observedGeneration"`  // status: current generation value
		Replicas            int `yaml:"replicas"`            // status: requested amount of instances
		UpdatedReplicas     int `yaml:"updatedReplicas"`     // status: up-to-date instances
		UnavailableReplicas int `yaml:"unavailableReplicas"` // status: total amount of unavailable instances (missed after ok deploy)
	}

	kubeContainerSpec struct {
		Image string `yaml:"image"` // example.com:80/dalee/image:34
	}

	kubeTypeSpec struct {
		Containers []kubeContainerSpec `yaml:"containers"`
	}

	kubeTypeTemplate struct {
		Spec kubeTypeSpec `yaml:"spec"`
	}

	kubeTypeResourceSpec struct {
		Template kubeTypeTemplate `yaml:"template"`
	}

	kubeResourceList struct {
		Kind  string         `yaml:"kind"`
		Items []kubeResource `yaml:"items"`
	}

	kubeResource struct {
		Kind     string               `yaml:"kind"`
		Metadata kubeTypeMetadata     `yaml:"metadata"`
		Spec     kubeTypeResourceSpec `yaml:"spec"`
		Status   kubeTypeStatus       `yaml:"status"`
	}

	kubeResourceListInterface interface {
		GetItems() []KubeResourceInterface
	}

	// List kind unpacking structures
	deploymentList struct {
		Items []Deployment `yaml:"items"`
	}

	replicaSetList struct {
		Items []ReplicaSet `yaml:"items"`
	}

	namespaceList struct {
		Items []Namespace `yaml:"items"`
	}

	// KubeResourceInterface is common interface to all k8s resource types
	KubeResourceInterface interface {
		GetKind() string
	}

	// Deployment is k8s Deployment resource
	Deployment struct {
		Kind     string               `yaml:"kind"`
		Metadata kubeTypeMetadata     `yaml:"metadata"`
		Spec     kubeTypeResourceSpec `yaml:"spec"`
		Status   kubeTypeStatus       `yaml:"status"`
	}

	// ReplicaSet is k8s ReplicaSet resource
	ReplicaSet struct {
		Kind     string               `yaml:"kind"`
		Metadata kubeTypeMetadata     `yaml:"metadata"`
		Spec     kubeTypeResourceSpec `yaml:"spec"`
		Status   kubeTypeStatus       `yaml:"status"`
	}

	// Namespace is k8s Namespace resource
	Namespace struct {
		Kind     string           `yaml:"kind"`
		Metadata kubeTypeMetadata `yaml:"metadata"`
	}
)

// generic type interface support
func (k *kubeResource) GetKind() string {
	return strings.ToLower(k.Kind)
}

// interface support method
func (d *kubeResourceList) GetKind() string {
	return strings.ToLower(d.Kind)
}

// GetKind interface method support, returns string "deployment"
func (d *Deployment) GetKind() string {
	return strings.ToLower(d.Kind)
}

// GetName return name of Deployment
func (d *Deployment) GetName() string {
	return d.Metadata.Name
}

// interface support method
func (dl *deploymentList) GetItems() []KubeResourceInterface {
	r := make([]KubeResourceInterface, 0)
	for i := range dl.Items {
		r = append(r, &dl.Items[i])
	}
	return r
}

// GetKind interface method support, returns string "replicaset"
func (r *ReplicaSet) GetKind() string {
	return strings.ToLower(r.Kind)
}

// GetName return name of ReplicaSet
func (r *ReplicaSet) GetName() string {
	return r.Metadata.Name
}

// GetImages return list of docker images registered in ReplicaSet
func (r *ReplicaSet) GetImages() []string {
	items := make([]string, 0)
	for _, c := range r.Spec.Template.Spec.Containers {
		items = append(items, c.Image)
	}
	return items
}

// interface support method
func (rl *replicaSetList) GetItems() []KubeResourceInterface {
	r := make([]KubeResourceInterface, 0)
	for i := range rl.Items {
		r = append(r, &rl.Items[i])
	}
	return r
}

// GetKind interface method support, returns string "namespace"
func (n *Namespace) GetKind() string {
	return strings.ToLower(n.Kind)
}

// GetName return name of Namespace
func (n *Namespace) GetName() string {
	return n.Metadata.Name
}

// interface support method
func (nl *namespaceList) GetItems() []KubeResourceInterface {
	r := make([]KubeResourceInterface, 0)
	for i := range nl.Items {
		r = append(r, &nl.Items[i])
	}
	return r
}
