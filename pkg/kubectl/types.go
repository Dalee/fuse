package kubectl

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// KindDeployment name of Deployment resource type
	KindDeployment = "deployment"

	// KindReplicaSet name of ReplicaSet resource type
	KindReplicaSet = "replicaset"

	// KindNamespace name of Namespace resource type
	KindNamespace = "namespace"

	// KindPod name of Pod resource type
	KindPod = "pod"

	// KindList name of List resource type
	KindList = "list"

	// ClusterContextEnv is the name of environment variable to get kubectl requested context
	ClusterContextEnv = "CLUSTER_CONTEXT"

	// ClusterReleaseTimeoutEnv is the name of environment variable for apply command
	ClusterReleaseTimeoutEnv = "CLUSTER_RELEASE_TIMEOUT"

	// rolling update strategy type
	strategyTypeRollingUpdate = "RollingUpdate"
)

type (
	resourceMetadata struct {
		Name       string            `yaml:"name"`
		Namespace  string            `yaml:"namespace"`
		Labels     map[string]string `yaml:"labels"`
		Generation int               `yaml:"generation"`
		UID        string            `yaml:"uid"`
	}

	resourceStatus struct {
		AvailableReplicas   int `yaml:"availableReplicas"`   // total number of available instances
		ObservedGeneration  int `yaml:"observedGeneration"`  // current generation value
		Replicas            int `yaml:"replicas"`            // requested number of instances
		UpdatedReplicas     int `yaml:"updatedReplicas"`     // up-to-date instances
		UnavailableReplicas int `yaml:"unavailableReplicas"` // total number of unavailable instances
	}

	resourceContainer struct {
		Image string `yaml:"image"` // example.com:80/dalee/image:34
	}

	resourceContainerSpec struct {
		Containers []resourceContainer `yaml:"containers"`
	}

	resourceTemplate struct {
		Spec resourceContainerSpec `yaml:"spec"`
	}

	resourceStrategyRolling struct {
		MaxSurge       int `yaml:"maxSurge"`
		MaxUnavailable int `yaml:"maxUnavailable"`
	}

	resourceStrategy struct {
		Type          string                  `yaml:"type"`
		RollingUpdate resourceStrategyRolling `yaml:"rollingUpdate"`
	}

	resourceSpec struct {
		Replicas int              `yaml:"replicas"`
		Template resourceTemplate `yaml:"template"`
		Strategy resourceStrategy `yaml:"strategy"`
	}

	kubeResourceList struct {
		Kind  string         `yaml:"kind"`
		Items []kubeResource `yaml:"items"`
	}

	kubeResource struct {
		Kind     string           `yaml:"kind"`
		Metadata resourceMetadata `yaml:"metadata"`
		Spec     resourceSpec     `yaml:"spec"`
		Status   resourceStatus   `yaml:"status"`
	}

	kubeResourceListInterface interface {
		GetItems() ResourceList
	}

	// List kind unpacking structures
	podList struct {
		Items []Pod `yaml:"items"`
	}

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
		GetName() string
		ToDeployment() (*Deployment, error)
	}

	// ResourceList is an alias for []KubeResourceInterface
	ResourceList []KubeResourceInterface

	// Pod is k8s Pod resource
	Pod struct {
		Kind     string                `yaml:"kind"`
		Metadata resourceMetadata      `yaml:"metadata"`
		Spec     resourceContainerSpec `yaml:"spec"`
		Status   resourceStatus        `yaml:"status"`
	}

	// Deployment is k8s Deployment resource
	Deployment struct {
		Kind     string           `yaml:"kind"`
		Metadata resourceMetadata `yaml:"metadata"`
		Spec     resourceSpec     `yaml:"spec"`
		Status   resourceStatus   `yaml:"status"`
	}

	// ReplicaSet is k8s ReplicaSet resource
	ReplicaSet struct {
		Kind     string           `yaml:"kind"`
		Metadata resourceMetadata `yaml:"metadata"`
		Spec     resourceSpec     `yaml:"spec"`
		Status   resourceStatus   `yaml:"status"`
	}

	// Namespace is k8s Namespace resource
	Namespace struct {
		Kind     string           `yaml:"kind"`
		Metadata resourceMetadata `yaml:"metadata"`
	}
)

// FilteredByKind return filtered slice of resources by kind
func (rl ResourceList) FilteredByKind(kind string) ResourceList {
	result := make(ResourceList, 0)
	for _, obj := range rl {
		if obj.GetKind() == kind {
			result = append(result, obj)
		}
	}

	return result
}

// ToDeploymentList is helper to convert []KubeResourceInterface to []Deployment
func (rl ResourceList) ToDeploymentList() []Deployment {
	dlist := make([]Deployment, 0)
	for _, obj := range rl {
		if obj.GetKind() == KindDeployment {
			d, _ := obj.(*Deployment)
			dlist = append(dlist, *d)
		}
	}

	return dlist
}

// ToReplicaSetList is helper to convert []KubeResourceInterface to []ReplicaSet
func (rl ResourceList) ToReplicaSetList() []ReplicaSet {
	rlist := make([]ReplicaSet, 0)
	for _, obj := range rl {
		if obj.GetKind() == KindReplicaSet {
			r, _ := obj.(*ReplicaSet)
			rlist = append(rlist, *r)
		}
	}

	return rlist
}

// ToNamespaceList is helper to convert []KubeResourceInterface type to []Namespace
func (rl ResourceList) ToNamespaceList() []Namespace {
	nlist := make([]Namespace, 0)
	for _, obj := range rl {
		if obj.GetKind() == KindNamespace {
			n, _ := obj.(*Namespace)
			nlist = append(nlist, *n)
		}
	}

	return nlist
}

// ToPodList is helper to convert []KubeResourceInterface type to []Pod
func (rl ResourceList) ToPodList() []Pod {
	plist := make([]Pod, 0)
	for _, obj := range rl {
		if obj.GetKind() == KindPod {
			p, _ := obj.(*Pod)
			plist = append(plist, *p)
		}
	}

	return plist
}

// generic type interface support
func (k *kubeResource) GetKind() string {
	return strings.ToLower(k.Kind)
}

// interface support method
func (k *kubeResource) GetName() string {
	return k.Metadata.Name
}

func (k *kubeResource) ToDeployment() (*Deployment, error) {
	return nil, errors.New("kubeResource can't be transformed to deployment")
}

// interface support method
func (d *kubeResourceList) GetKind() string {
	return strings.ToLower(d.Kind)
}

func (d *kubeResourceList) GetItems() ResourceList {
	r := make([]KubeResourceInterface, 0)
	for i := range d.Items {
		r = append(r, &d.Items[i])
	}
	return r
}

// GetKind interface method support, returns string "deployment"
func (d *Deployment) GetKind() string {
	return strings.ToLower(d.Kind)
}

// GetName return name of Deployment
func (d *Deployment) GetName() string {
	return d.Metadata.Name
}

// GetNamespace return deployment namespace
func (d *Deployment) GetNamespace() string {
	return formatNamespace(d.Metadata.Namespace)
}

// GetUUID return UUID of deployment
func (d *Deployment) GetUUID() string {
	return d.Metadata.UID
}

// GetKey will return unique name within a cluster
func (d *Deployment) GetKey() string {
	return fmt.Sprintf("%s/%s", d.GetNamespace(), d.GetName())
}

// GetGeneration will return current deployment generation
func (d *Deployment) GetGeneration() int {
	return d.Metadata.Generation
}

// GetSelector return slice of selectors associated with Deployment
func (d *Deployment) GetSelector() []string {
	selectorList := make([]string, 0)
	for key, value := range d.Metadata.Labels {
		selectorList = append(selectorList, fmt.Sprintf("%s=%s", key, value))
	}

	return selectorList
}

// ToDeployment interface method
func (d *Deployment) ToDeployment() (*Deployment, error) {
	return d, nil
}

// IsReady check deploy has been rolled out
// @see https://kubernetes.io/docs/user-guide/deployments/#the-status-of-a-deployment
func (d *Deployment) IsReady() bool {
	isReady := d.Status.ObservedGeneration >= d.Metadata.Generation
	isReady = isReady && (d.Status.UpdatedReplicas >= d.Spec.Replicas)
	isReady = isReady && (d.Status.UnavailableReplicas == 0)

	if (d.Spec.Replicas != 0) && (d.Spec.Strategy.Type == strategyTypeRollingUpdate) {
		replicaMinRequired := d.Spec.Replicas - d.Spec.Strategy.RollingUpdate.MaxUnavailable
		isReady = isReady && (d.Status.AvailableReplicas >= replicaMinRequired)
	}

	return isReady
}

// GetStatusString return deployment status message
func (d *Deployment) GetStatusString() string {
	return fmt.Sprintf(
		"Ready: %v, Generation: meta=%d observed=%d, Replicas: s=%d, u=%d, a=%d, na=%d",
		d.IsReady(),
		d.Metadata.Generation,
		d.Status.ObservedGeneration,
		d.Spec.Replicas,
		d.Status.UpdatedReplicas,
		d.Status.AvailableReplicas,
		d.Status.UnavailableReplicas,
	)
}

// GetItems is an interface support method
func (dl *deploymentList) GetItems() ResourceList {
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

// ToDeployment interface method
func (r *ReplicaSet) ToDeployment() (*Deployment, error) {
	return nil, errors.New("ReplicaSet can't be transformed to deployment")
}

// GetItems interface method
func (rl *replicaSetList) GetItems() ResourceList {
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

// ToDeployment interface method
func (n *Namespace) ToDeployment() (*Deployment, error) {
	return nil, errors.New("Namespace can't be transformed to deployment")
}

// GetItems is an interface support method
func (nl *namespaceList) GetItems() ResourceList {
	r := make([]KubeResourceInterface, 0)
	for i := range nl.Items {
		r = append(r, &nl.Items[i])
	}
	return r
}

// GetKind is an interface method
func (p *Pod) GetKind() string {
	return strings.ToLower(p.Kind)
}

// GetName is an interface method
func (p *Pod) GetName() string {
	return p.Metadata.Name
}

// GetNamespace return namespace name for pod
func (p *Pod) GetNamespace() string {
	return p.Metadata.Namespace
}

// GetKey will return unique name within a cluster
func (p *Pod) GetKey() string {
	return fmt.Sprintf("%s/%s", p.GetNamespace(), p.GetName())
}

// ToDeployment interface method
func (p *Pod) ToDeployment() (*Deployment, error) {
	return nil, errors.New("Pod can't be transformed to deployment")
}

// GetItems interface method
func (pl *podList) GetItems() ResourceList {
	r := make([]KubeResourceInterface, 0)
	for i := range pl.Items {
		r = append(r, &pl.Items[i])
	}
	return r
}
