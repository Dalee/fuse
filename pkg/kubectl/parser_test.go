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
  generation: 42
  labels:
    app: test-label
spec:
  template:
    metadata:
      labels:
        project: "test"
    spec:
      containers:
      - image: example.com/image:1
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 1)

	dlist := result.ToDeploymentList()
	assert.Len(t, dlist, 1)

	d := dlist[0]
	assert.Equal(t, KindDeployment, d.GetKind())
	assert.Equal(t, "test-deployment", d.GetName())
	assert.Equal(t, "default", d.Metadata.Namespace)
	assert.Len(t, d.Spec.Template.Spec.Containers, 1)

	container := d.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "example.com/image:1", container.Image)

	assert.Len(t, d.Metadata.Labels, 1)
	assert.Equal(t, d.Metadata.Labels["app"], "test-label")
	assert.Equal(t, []string{"project=test"}, d.GetPodSelector())

	assert.Equal(t, d.Metadata.Generation, 42)
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

	rlist := result.ToReplicaSetList()
	assert.Len(t, rlist, 1)

	r := rlist[0]
	assert.Equal(t, KindReplicaSet, r.GetKind())
	assert.Equal(t, "test-replicaset", r.GetName())
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

	nlist := result.ToNamespaceList()
	assert.Len(t, nlist, 1)

	n := nlist[0]
	assert.Equal(t, KindNamespace, n.GetKind())
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

	nlist := result.ToNamespaceList()
	assert.Len(t, nlist, 2)

	n1 := nlist[0]
	assert.Equal(t, KindNamespace, n1.GetKind())
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

	nlist := result.ToNamespaceList()
	assert.Len(t, nlist, 2)

	// check the first
	n1 := nlist[0]
	assert.Equal(t, KindNamespace, n1.GetKind())
	assert.Equal(t, "default", n1.GetName())

	// check the second
	n2 := nlist[1]
	assert.Equal(t, KindNamespace, n2.GetKind())
	assert.Equal(t, "kube-system", n2.GetName())
}

// ensure parse list with pod items works as expected
func TestParsePodList(t *testing.T) {
	rawYamlString := `---
apiVersion: v1
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: example-1-pod
    namespace: kube-system
- apiVersion: v1
  kind: Pod
  metadata:
    name: example-2-pod
    namespace: default
  status:
    phase: Running
kind: List
metadata: {}
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 2)

	plist := result.ToPodList()
	assert.Len(t, plist, 2)

	// check the first
	p1 := plist[0]
	assert.Equal(t, KindPod, p1.GetKind())
	assert.Equal(t, "example-1-pod", p1.GetName())

	// check the second
	p2 := plist[1]
	assert.Equal(t, KindPod, p2.GetKind())
	assert.Equal(t, "example-2-pod", p2.GetName())
	assert.Equal(t, PodStatusRunning, p2.Status.Phase)
}

func TestParsePod(t *testing.T) {
	rawYamlString := `---
apiVersion: extensions/v1beta1
kind: Pod
metadata:
  name: test-pod
  namespace: default
  labels:
    app: test-label
spec:
  containers:
    - image: example.com/image:1
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 1)

	plist := result.ToPodList()
	assert.Len(t, plist, 1)

	pod := plist[0]
	assert.Equal(t, KindPod, pod.GetKind())
	assert.Equal(t, "test-pod", pod.GetName())
	assert.Equal(t, "default", pod.Metadata.Namespace)
	assert.Len(t, pod.Spec.Containers, 1)

	container := pod.Spec.Containers[0]
	assert.Equal(t, "example.com/image:1", container.Image)

	assert.Len(t, pod.Metadata.Labels, 1)
	assert.Equal(t, pod.Metadata.Labels["app"], "test-label")
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
    uid: 3eb259fc-bc6f-11e6-a342-005056ba5444
    generation: 42
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
    uid: 07abb126-0266-11e7-931c-005056ba5444
    generation: 15
  spec:
    template:
      spec:
        containers:
        - image: example.com/image:2
kind: List
metadata: {}
---
`
	p := newParser()
	result, err := p.parseYaml([]byte(rawYamlString))
	assert.Nil(t, err)
	assert.Len(t, result, 2)

	dlist := result.ToDeploymentList()
	assert.Len(t, dlist, 2)

	// check the first
	d1 := dlist[0]
	assert.Equal(t, KindDeployment, d1.GetKind())
	assert.Equal(t, "3eb259fc-bc6f-11e6-a342-005056ba5444", d1.GetUUID())
	assert.Equal(t, "test-deployment1", d1.GetName())
	assert.Equal(t, "default", d1.GetNamespace())
	assert.Equal(t, 42, d1.GetGeneration())
	assert.Len(t, d1.Spec.Template.Spec.Containers, 1)

	// check the second
	d2 := dlist[1]
	assert.Equal(t, KindDeployment, d2.GetKind())
	assert.Equal(t, "07abb126-0266-11e7-931c-005056ba5444", d2.GetUUID())
	assert.Equal(t, "test-deployment2", d2.GetName())
	assert.Equal(t, "kube-system", d2.GetNamespace())
	assert.Equal(t, 15, d2.GetGeneration())
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

	rlist := result.ToReplicaSetList()
	assert.Len(t, rlist, 2)

	// first
	r := rlist[0]
	assert.Equal(t, KindReplicaSet, r.GetKind())
	assert.Equal(t, "test-replicaset1", r.GetName())
	assert.Equal(t, "kube-system", r.Metadata.Namespace)
	assert.Len(t, r.Spec.Template.Spec.Containers, 1)

	container := r.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "example.com/image:1", container.Image)

	// second
	r2 := rlist[1]
	assert.Equal(t, KindReplicaSet, r2.GetKind())
	assert.Equal(t, "test-replicaset2", r2.Metadata.Name)
	assert.Equal(t, "default", r2.Metadata.Namespace)
	assert.Len(t, r2.Spec.Template.Spec.Containers, 1)

	container2 := r2.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "example.com/image:2", container2.Image)
}

// ensure local file parsing is supported
func TestParseLocalFile_Deployment(t *testing.T) {
	result, err := ParseLocalFile("./testdata/parser_test1.yml")
	assert.Nil(t, err)
	assert.Len(t, result, 1)

	dlist := result.ToDeploymentList()
	assert.Len(t, dlist, 1)

	d := dlist[0]
	assert.Equal(t, KindDeployment, d.GetKind())
	assert.Equal(t, "test-deployment", d.GetName())
	assert.Equal(t, "default", d.Metadata.Namespace)
	assert.Len(t, d.Spec.Template.Spec.Containers, 1)

	container := d.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "example.com/image:1", container.Image)
}

// ensure local file with absent file will raise an error
func TestParseLocalFile_DeploymentAbsentFile(t *testing.T) {
	result, err := ParseLocalFile("./testdata/__not_exist__")
	assert.Nil(t, result)
	assert.Error(t, err)
}

// ensure local file parsing is supported
func TestParseLocalFile_AbsFailed(t *testing.T) {
	result, err := ParseLocalFile("\\testdata\\parser_test1.yml")
	assert.Nil(t, result)
	assert.Error(t, err)
}
