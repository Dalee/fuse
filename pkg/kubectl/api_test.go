package kubectl

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type (
	kubeCommandMock struct {
		mock.Mock
	}

	kubeParserMock struct {
		mock.Mock
	}
)

// implement interface for kubeCommandMock
func (cm *kubeCommandMock) Log() {
}

func (cm *kubeCommandMock) Run() ([]byte, bool) {
	args := cm.Called()
	return args.Get(0).([]byte), args.Bool(1)
}

func (cm *kubeCommandMock) getCommand() *exec.Cmd {
	return nil
}

// implement interface for kubeParserMock
func (pm *kubeParserMock) parseYaml(data []byte) ([]KubeResourceInterface, error) {
	var parsedResource []KubeResourceInterface
	args := pm.Called()

	passedResult := args.Get(0)
	if passedResult != nil {
		parsedResource = passedResult.([]KubeResourceInterface)
	}

	return parsedResource, args.Error(1)
}

func TestKubeCall_RunPlain(t *testing.T) {
	cmdMock := new(kubeCommandMock)
	cmdMock.On("Run").Return([]byte("Hello world"), true)

	call := &KubeCall{
		Cmd:    cmdMock,
		Parser: nil,
	}

	output, err := call.RunPlain()
	assert.Nil(t, err)
	assert.Equal(t, []byte("Hello world"), output)
}

//
func TestKubeCall_RunNormal(t *testing.T) {
	cmdMock := new(kubeCommandMock)
	cmdMock.On("Run").Return([]byte(""), true)

	parserMock := new(kubeParserMock)
	parserMock.On("parseYaml").Return(make([]KubeResourceInterface, 0), nil)

	call := &KubeCall{
		Cmd:    cmdMock,
		Parser: parserMock,
	}

	items, err := call.RunAndParse()
	assert.Nil(t, err)
	assert.Len(t, items, 0)
}

//
func TestKubeCall_RunAndParseFirst(t *testing.T) {
	cmdMock := new(kubeCommandMock)
	cmdMock.On("Run").Return([]byte(""), true)

	parsedList := make([]KubeResourceInterface, 0)
	parsedList = append(parsedList, &kubeResource{
		Kind: "Namespace",
	})

	parserMock := new(kubeParserMock)
	parserMock.On("parseYaml").Return(parsedList, nil)

	call := &KubeCall{
		Cmd:    cmdMock,
		Parser: parserMock,
	}

	item, err := call.RunAndParseFirst()
	assert.Nil(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "namespace", item.GetKind())
}

//
func TestKubeCall_RunCommandFailed(t *testing.T) {
	cmdMock := new(kubeCommandMock)
	cmdMock.On("Run").Return([]byte(""), false)

	parserMock := new(kubeParserMock)
	parserMock.On("parseYaml").Return(make([]KubeResourceInterface, 0), nil)

	call := &KubeCall{
		Cmd:    cmdMock,
		Parser: parserMock,
	}

	items, err := call.RunAndParse()
	assert.Error(t, err, "Command exited with non-zero status")
	assert.Nil(t, items)
}

func TestKubeCall_RunParserFailed(t *testing.T) {
	cmdMock := new(kubeCommandMock)
	cmdMock.On("Run").Return([]byte(""), false)

	parserMock := new(kubeParserMock)
	parserMock.On("parseYaml").Return(nil, errors.New("Something wrong with parser"))

	call := &KubeCall{
		Cmd:    cmdMock,
		Parser: parserMock,
	}

	items, err := call.RunAndParse()
	assert.Error(t, err)
	assert.Nil(t, items)
}

// get namespace list
func TestCommandNamespaceList(t *testing.T) {
	cmd := CommandNamespaceList()

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl get namespaces -o yaml", args)
}

// get replicaset list
func TestCommandReplicaSetList(t *testing.T) {
	os.Setenv("CLUSTER_CONTEXT", "prod")
	cmd := CommandReplicaSetList("kube-system")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --context=prod --namespace=kube-system get replicasets -o yaml", args)
	os.Unsetenv("CLUSTER_CONTEXT")
}

//
func TestCommandReplicaSetListWithDefaultNamespace(t *testing.T) {
	cmd := CommandReplicaSetList("")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=default get replicasets -o yaml", args)
}

// get deployments in namespace
func TestCommandDeploymentList(t *testing.T) {
	cmd := CommandDeploymentList("kube-system")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=kube-system get deployments -o yaml", args)
}

// get deployments in namespace
func TestCommandDeploymentListWithDefaultNamespace(t *testing.T) {
	cmd := CommandDeploymentList("")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=default get deployments -o yaml", args)
}

// test get logs from pods
func TestCommandPodLogs(t *testing.T) {
	cmd := CommandPodLogs("", "pod-123456")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=default logs pod-123456", args)
}

// test get pod list by selector
func TestCommandPodList(t *testing.T) {
	cmd := CommandPodList("", []string{"app=prod-v1", "name=example"})

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=default get pods --selector=app=prod-v1,name=example -o yaml", args)
}
