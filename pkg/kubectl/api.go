package kubectl

import (
	"errors"
	"fmt"
)

type (
	// KubeCall is kubectl wrapper combined with parser
	KubeCall struct {
		Cmd    kubeCommandInterface
		Parser kubeResourceParserInterface
	}
)

// RunPlain run command and just return command output
func (c *KubeCall) RunPlain() ([]byte, error) {
	c.Cmd.Log()

	output, success := c.Cmd.Run()
	if success != true {
		return output, errors.New(string(output))
	}

	return output, nil
}

// RunAndParse run command and try to parse output with provided parser
func (c *KubeCall) RunAndParse() ([]KubeResourceInterface, error) {
	output, err := c.RunPlain()
	if err != nil {
		return nil, err
	}

	items, err := c.Parser.parseYaml(output)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// RunAndParseFirst run command, parse output and return first element of decoded items
func (c *KubeCall) RunAndParseFirst() (KubeResourceInterface, error) {
	items, err := c.RunAndParse()
	if err != nil {
		return nil, err
	}

	if len(items) > 0 {
		return items[0], nil
	}

	return nil, nil
}

// CommandNamespaceList return call which will return list of namespaces registered in kubernetes cluster
func CommandNamespaceList() (*KubeCall) {
	p := newParser()
	c := newCommand([]string{
		"get",
		"namespaces",
		"-o",
		"yaml",
	})

	return &KubeCall{
		Cmd: c,
		Parser: p,
	}
}

// CommandReplicaSetList return call which return list of replicasets registered in kubernetes cluster
func CommandReplicaSetList(namespace string) (*KubeCall) {
	if namespace == "" {
		namespace = "default"
	}

	p := newParser()
	c := newCommand([]string{
		fmt.Sprintf("--namespace=%s", namespace),
		"get",
		"replicasets",
		"-o",
		"yaml",
	})

	return &KubeCall{
		Cmd: c,
		Parser: p,
	}
}

// CommandDeploymentList return call which return list of deployments registered in kubernetes clusted
func CommandDeploymentList(namespace string) (*KubeCall) {
	if namespace == "" {
		namespace = "default"
	}

	p := newParser()
	c := newCommand([]string{
		fmt.Sprintf("--namespace=%s", namespace),
		"get",
		"deployments",
		"-o",
		"yaml",
	})

	return &KubeCall{
		Cmd: c,
		Parser: p,
	}
}
