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

func (cm *kubeCommandMock) Log() {
}

func (cm *kubeCommandMock) Run() ([]byte, bool) {
	args := cm.Called()
	return args.Get(0).([]byte), args.Bool(1)
}

func (cm *kubeCommandMock) getCommand() *exec.Cmd {
	return nil
}

func (pm *kubeParserMock) parseYaml(data []byte) (ResourceList, error) {
	var parsedResource ResourceList
	args := pm.Called()

	passedResult := args.Get(0)
	if passedResult != nil {
		parsedResource = passedResult.(ResourceList)
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

func TestKubeCall_RunNormal(t *testing.T) {
	cmdMock := new(kubeCommandMock)
	cmdMock.On("Run").Return([]byte(""), true)

	parserMock := new(kubeParserMock)
	parserMock.On("parseYaml").Return(make(ResourceList, 0), nil)

	call := &KubeCall{
		Cmd:    cmdMock,
		Parser: parserMock,
	}

	items, err := call.RunAndParse()
	assert.Nil(t, err)
	assert.Len(t, items, 0)
}

func TestKubeCall_RunAndParseFirst(t *testing.T) {
	cmdMock := new(kubeCommandMock)
	cmdMock.On("Run").Return([]byte(""), true)

	parsedList := make(ResourceList, 0)
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

func TestKubeCall_RunAndParseFirstEmpty(t *testing.T) {
	cmdMock := new(kubeCommandMock)
	cmdMock.On("Run").Return([]byte(""), true)

	parserMock := new(kubeParserMock)
	parserMock.On("parseYaml").Return(make(ResourceList, 0), nil)

	call := &KubeCall{
		Cmd:    cmdMock,
		Parser: parserMock,
	}

	item, err := call.RunAndParseFirst()
	assert.Nil(t, err)
	assert.Nil(t, item)
}

func TestKubeCall_RunAndParseFirstError(t *testing.T) {
	cmdMock := new(kubeCommandMock)
	cmdMock.On("Run").Return([]byte(""), true)

	parserMock := new(kubeParserMock)
	parserMock.On("parseYaml").Return(nil, errors.New("Parser is not available"))

	call := &KubeCall{
		Cmd:    cmdMock,
		Parser: parserMock,
	}

	item, err := call.RunAndParseFirst()
	assert.Error(t, err)
	assert.Nil(t, item)
}

func TestKubeCall_RunCommandFailed(t *testing.T) {
	cmdMock := new(kubeCommandMock)
	cmdMock.On("Run").Return([]byte(""), false)

	parserMock := new(kubeParserMock)
	parserMock.On("parseYaml").Return(make(ResourceList, 0), nil)

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

func TestCommandApply(t *testing.T) {
	cmd := CommandApply("test.yaml")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl apply -f test.yaml -o name", args)
}

func TestCommandRollback(t *testing.T) {
	cmd := CommandRollback("default", "deployment", "example-deployment")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=default rollout undo deployment/example-deployment", args)
}

func TestCommandNamespaceList(t *testing.T) {
	cmd := CommandNamespaceList()

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl get namespaces -o yaml", args)
}

func TestCommandReplicaSetList(t *testing.T) {
	os.Setenv(ClusterContextEnv, "prod")
	cmd := CommandReplicaSetList("kube-system")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --context=prod --namespace=kube-system get replicasets -o yaml", args)
	os.Unsetenv(ClusterContextEnv)
}

func TestCommandReplicaSetBySelector(t *testing.T) {
	cmd := CommandReplicaSetListBySelector("kube-system", []string{"app=example-app"})

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=kube-system get replicasets --selector=app=example-app -o yaml", args)
}

func TestCommandReplicaSetListWithDefaultNamespace(t *testing.T) {
	cmd := CommandReplicaSetList("")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=default get replicasets -o yaml", args)
}

func TestCommandDescribeDeployment(t *testing.T) {
	cmd := CommandDeploymentInfo("sample-namespace", "example")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=sample-namespace get deployment/example -o yaml", args)
}

func TestCommandDeploymentList(t *testing.T) {
	cmd := CommandDeploymentList("kube-system")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=kube-system get deployments -o yaml", args)
}

func TestCommandDeploymentListWithDefaultNamespace(t *testing.T) {
	cmd := CommandDeploymentList("")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=default get deployments -o yaml", args)
}

func TestCommandPodList(t *testing.T) {
	cmd := CommandPodListBySelector("", []string{"app=prod-v1", "name=example"})

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=default get pods --selector=app=prod-v1,name=example -o yaml", args)
}

func TestCommandPodLogs(t *testing.T) {
	cmd := CommandPodLogs("", "pod-123456")

	args := strings.Join(cmd.Cmd.getCommand().Args, " ")
	assert.Equal(t, "kubectl --namespace=default logs pod-123456", args)
}
