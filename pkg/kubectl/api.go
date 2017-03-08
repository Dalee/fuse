package kubectl

import (
	"errors"
	"fmt"
	"strings"
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
func (c *KubeCall) RunAndParse() (ResourceList, error) {
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

// CommandApply apply new yaml configuration to cluster
func CommandApply(configurationYaml string) *KubeCall {
	p := newParser()
	c := newCommand([]string{
		"apply",
		"-f",
		configurationYaml,
		"-o",
		"name",
	})

	return &KubeCall{
		Cmd:    c,
		Parser: p,
	}
}

// CommandRollback allow to rollback any resource to previous version
func CommandRollback(namespace, kind, name string) *KubeCall {
	p := newParser()
	c := newCommand([]string{
		fmt.Sprintf("--namespace=%s", formatNamespace(namespace)),
		"rollout",
		"undo",
		fmt.Sprintf("%s/%s", kind, name),
	})

	return &KubeCall{
		Cmd:    c,
		Parser: p,
	}
}

// CommandNamespaceList return call which will return list of namespaces registered in kubernetes cluster
func CommandNamespaceList() *KubeCall {
	p := newParser()
	c := newCommand([]string{
		"get",
		"namespaces",
		"-o",
		"yaml",
	})

	return &KubeCall{
		Cmd:    c,
		Parser: p,
	}
}

// CommandReplicaSetList return call which return list of replicasets registered in kubernetes cluster
func CommandReplicaSetList(namespace string) *KubeCall {
	p := newParser()
	c := newCommand([]string{
		fmt.Sprintf("--namespace=%s", formatNamespace(namespace)),
		"get",
		"replicasets",
		"-o",
		"yaml",
	})

	return &KubeCall{
		Cmd:    c,
		Parser: p,
	}
}

// CommandReplicaSetListBySelector get replica set list by selector
func CommandReplicaSetListBySelector(namespace string, selector []string) *KubeCall {
	selectorList := strings.Join(selector, ",")
	p := newParser()
	c := newCommand([]string{
		fmt.Sprintf("--namespace=%s", formatNamespace(namespace)),
		"get",
		"replicasets",
		fmt.Sprintf("--selector=%s", selectorList),
		"-o",
		"yaml",
	})

	return &KubeCall{
		Cmd:    c,
		Parser: p,
	}
}

// CommandDeploymentInfo get information about single deployment
func CommandDeploymentInfo(namespace string, deployment string) *KubeCall {
	p := newParser()
	c := newCommand([]string{
		fmt.Sprintf("--namespace=%s", formatNamespace(namespace)),
		"get",
		fmt.Sprintf("deployment/%s", deployment),
		"-o",
		"yaml",
	})

	return &KubeCall{
		Cmd:    c,
		Parser: p,
	}
}

// CommandDeploymentList return call which return list of deployments registered in kubernetes clusted
func CommandDeploymentList(namespace string) *KubeCall {
	p := newParser()
	c := newCommand([]string{
		fmt.Sprintf("--namespace=%s", formatNamespace(namespace)),
		"get",
		"deployments",
		"-o",
		"yaml",
	})

	return &KubeCall{
		Cmd:    c,
		Parser: p,
	}
}

// CommandPodList return list of pods in namespace with selector
func CommandPodList(namespace string, selector []string) *KubeCall {
	selectorList := strings.Join(selector, ",")
	p := newParser()
	c := newCommand([]string{
		fmt.Sprintf("--namespace=%s", formatNamespace(namespace)),
		"get",
		"pods",
		fmt.Sprintf("--selector=%s", selectorList),
		"-o",
		"yaml",
	})

	return &KubeCall{
		Cmd:    c,
		Parser: p,
	}
}

// CommandPodLogs return logs for pod
func CommandPodLogs(namespace, pod string) *KubeCall {
	p := newParser()
	c := newCommand([]string{
		fmt.Sprintf("--namespace=%s", formatNamespace(namespace)),
		"logs",
		pod,
	})

	return &KubeCall{
		Cmd:    c,
		Parser: p,
	}
}

func formatNamespace(namespace string) string {
	if namespace == "" {
		return "default"
	}
	return namespace
}
