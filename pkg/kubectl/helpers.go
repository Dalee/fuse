package kubectl

// ToDeploymentList is helper to convert []KubeResourceInterface to []Deployment
func ToDeploymentList(list []KubeResourceInterface) []Deployment {
	dlist := make([]Deployment, 0)
	for _, item := range list {
		d, ok := item.(*Deployment)
		if !ok {
			panic("Unable to typecase to ReplicaSet")
		}

		dlist = append(dlist, *d)
	}

	return dlist
}

// ToReplicaSetList is helper to convert []KubeResourceInterface to []ReplicaSet
func ToReplicaSetList(list []KubeResourceInterface) []ReplicaSet {
	rlist := make([]ReplicaSet, 0)
	for _, item := range list {
		r, ok := item.(*ReplicaSet)
		if !ok {
			panic("Unable to typecase to ReplicaSet")
		}

		rlist = append(rlist, *r)
	}

	return rlist
}

// ToNamespaceList is helper to convert []KubeResourceInterface type to []Namespace
func ToNamespaceList(list []KubeResourceInterface) []Namespace {
	nlist := make([]Namespace, 0)
	for _, item := range list {
		n, ok := item.(*Namespace)
		if !ok {
			panic("Unable to typecase to NameSpace")
		}

		nlist = append(nlist, *n)
	}

	return nlist
}
