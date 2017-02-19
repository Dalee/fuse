package kubectl

// ToDeploymentList is helper to convert []KubeResourceInterface to []Deployment
func ToDeploymentList(list []KubeResourceInterface) []Deployment {
	dlist := make([]Deployment, 0)
	for _, item := range list {
		d, _ := item.(*Deployment)
		dlist = append(dlist, *d)
	}

	return dlist
}

// ToReplicaSetList is helper to convert []KubeResourceInterface to []ReplicaSet
func ToReplicaSetList(list []KubeResourceInterface) []ReplicaSet {
	rlist := make([]ReplicaSet, 0)
	for _, item := range list {
		r, _ := item.(*ReplicaSet)
		rlist = append(rlist, *r)
	}

	return rlist
}

// ToNamespaceList is helper to convert []KubeResourceInterface type to []Namespace
func ToNamespaceList(list []KubeResourceInterface) []Namespace {
	nlist := make([]Namespace, 0)
	for _, item := range list {
		n, _ := item.(*Namespace)
		nlist = append(nlist, *n)
	}

	return nlist
}
