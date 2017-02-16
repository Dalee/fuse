package kubectl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// ensure unknown type is not parsed
func TestParseUnknown(t *testing.T) {
	rawYamlString := `---
apiVersion: extensions/v1beta1
kind: Unknown
metadata:
  name: test-unknown
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: example.com/image:1
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 0)
}

// ensure wrong yaml is not firing any error
func TestParseRandom(t *testing.T) {
	rawYamlString := `---
just: "a"
random: "string"
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 0)
}

// ensure parse deployments works as expected
func TestParseDeployment(t *testing.T) {
	rawYamlString := `---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: test-deployment
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: example.com/image:1
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 1)

	dlist := ToDeploymentList(result)
	assert.Len(t, dlist, 1)

	d := dlist[0]
	assert.Equal(t, "deployment", d.GetKind())
	assert.Equal(t, "test-deployment", d.Metadata.Name)
	assert.Equal(t, "default", d.Metadata.Namespace)
	assert.Len(t, d.Spec.Template.Spec.Containers, 1)

	container := d.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "example.com/image:1", container.Image)
}

// ensure parse replicaset works as expected
func TestParseReplicaSet(t *testing.T) {
	rawYamlString := `---
apiVersion: extensions/v1beta1
kind: ReplicaSet
metadata:
  name: test-replicaset
  namespace: kube-system
spec:
  template:
    metadata:
    spec:
      containers:
      - image: example.com/image:1
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 1)

	rlist := ToReplicaSetList(result)
	assert.Len(t, rlist, 1)

	r := rlist[0]
	assert.Equal(t, "replicaset", r.GetKind())
	assert.Equal(t, "test-replicaset", r.Metadata.Name)
	assert.Equal(t, "kube-system", r.Metadata.Namespace)
	assert.Len(t, r.Spec.Template.Spec.Containers, 1)

	container := r.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "example.com/image:1", container.Image)
}

// ensure namespace parsing is ok
func TestParseNamespace(t *testing.T) {
	rawYamlString := `---
apiVersion: v1
kind: Namespace
metadata:
  name: kube-system
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 1)

	nlist := ToNamespaceList(result)
	assert.Len(t, nlist, 1)

	n := nlist[0]
	assert.Equal(t, "namespace", n.GetKind())
	assert.Equal(t, "kube-system", n.GetName())
}

// ensure namespace parsing is ok
func TestParseNamespaceMultipleDefinitions(t *testing.T) {
	rawYamlString := `
apiVersion: v1
kind: Namespace
metadata:
  name: kube-system
---
apiVersion: v1
kind: Namespace
metadata:
  name: my-new-shiny-namespace
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 2)

	nlist := ToNamespaceList(result)
	assert.Len(t, nlist, 2)

	n1 := nlist[0]
	assert.Equal(t, "namespace", n1.GetKind())
	assert.Equal(t, "kube-system", n1.GetName())

	n2 := nlist[1]
	assert.Equal(t, "namespace", n2.GetKind())
	assert.Equal(t, "my-new-shiny-namespace", n2.GetName())
}

// ensure parse list with namespace items works as expected
func TestParseNamespaceList(t *testing.T) {
	rawYamlString := `---
apiVersion: v1
items:
- apiVersion: v1
  kind: Namespace
  metadata:
    name: default
- apiVersion: v1
  kind: Namespace
  metadata:
    name: kube-system
kind: List
metadata: {}
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 2)

	nlist := ToNamespaceList(result)
	assert.Len(t, nlist, 2)

	// check the first
	n1 := nlist[0]
	assert.Equal(t, "namespace", n1.GetKind())
	assert.Equal(t, "default", n1.Metadata.Name)

	// check the second
	n2 := nlist[1]
	assert.Equal(t, "namespace", n2.GetKind())
	assert.Equal(t, "kube-system", n2.Metadata.Name)
}

// ensure parse list with namespace items works as expected
func TestParseDeploymentList(t *testing.T) {
	rawYamlString := `apiVersion: v1
items:
- apiVersion: extensions/v1beta1
  kind: Deployment
  metadata:
    name: test-deployment1
    namespace: default
  spec:
    template:
      spec:
        containers:
        - image: example.com/image:1
- apiVersion: extensions/v1beta1
  kind: Deployment
  metadata:
    name: test-deployment2
    namespace: kube-system
  spec:
    template:
      spec:
        containers:
        - image: example.com/image:2
kind: List
metadata: {}
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 2)

	dlist := ToDeploymentList(result)
	assert.Len(t, dlist, 2)

	// check the first
	d1 := dlist[0]
	assert.Equal(t, "deployment", d1.GetKind())
	assert.Equal(t, "default", d1.Metadata.Namespace)
	assert.Equal(t, "test-deployment1", d1.Metadata.Name)
	assert.Len(t, d1.Spec.Template.Spec.Containers, 1)

	// check the second
	d2 := dlist[1]
	assert.Equal(t, "deployment", d2.GetKind())
	assert.Equal(t, "kube-system", d2.Metadata.Namespace)
	assert.Equal(t, "test-deployment2", d2.Metadata.Name)
	assert.Len(t, d2.Spec.Template.Spec.Containers, 1)
}

// ensure parse replicaset works as expected
func TestParseReplicaSetList(t *testing.T) {
	rawYamlString := `
apiVersion: v1
items:
- apiVersion: extensions/v1beta1
  kind: ReplicaSet
  metadata:
    name: test-replicaset1
    namespace: kube-system
  spec:
    template:
      metadata:
      spec:
        containers:
        - image: example.com/image:1
- apiVersion: extensions/v1beta1
  kind: ReplicaSet
  metadata:
    name: test-replicaset2
    namespace: default
  spec:
    template:
      metadata:
      spec:
        containers:
        - image: example.com/image:2
kind: List
metadata: {}
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 2)

	rlist := ToReplicaSetList(result)
	assert.Len(t, rlist, 2)

	// first
	r := rlist[0]
	assert.Equal(t, "replicaset", r.GetKind())
	assert.Equal(t, "test-replicaset1", r.Metadata.Name)
	assert.Equal(t, "kube-system", r.Metadata.Namespace)
	assert.Len(t, r.Spec.Template.Spec.Containers, 1)

	container := r.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "example.com/image:1", container.Image)

	// second
	r2 := rlist[1]
	assert.Equal(t, "replicaset", r2.GetKind())
	assert.Equal(t, "test-replicaset2", r2.Metadata.Name)
	assert.Equal(t, "default", r2.Metadata.Namespace)
	assert.Len(t, r2.Spec.Template.Spec.Containers, 1)

	container2 := r2.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "example.com/image:2", container2.Image)
}
