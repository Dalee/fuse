package kubectl

import (
	"bufio"
	"bytes"

	"github.com/ghodss/yaml"
	"io/ioutil"
	"path/filepath"
)

type (
	kubeResourceParserInterface interface {
		parseYaml(data []byte) ([]KubeResourceInterface, error)
	}

	kubeResourceParser struct {
	}
)

func newParser() kubeResourceParserInterface {
	return &kubeResourceParser{}
}

// ParseLocalFile will allow to parse local file and fetch all resources defined there
func ParseLocalFile(filename string) ([]KubeResourceInterface, error) {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return newParser().parseYaml(data)
}

// parse whole kubectl answer into list of objects
func (p *kubeResourceParser) parseYaml(data []byte) ([]KubeResourceInterface, error) {
	var err error
	typeList := make([]KubeResourceInterface, 0)
	maxBuffSize := 1024 * 1024 * 200 // 200MB

	breader := bytes.NewReader(data)
	scanner := bufio.NewScanner(breader)

	scanner.Buffer(data, maxBuffSize)
	scanner.Split(splitYAMLDocument)

	for scanner.Scan() {
		resource := &kubeResource{}

		chunkData := scanner.Bytes()
		if err := yaml.Unmarshal(chunkData, resource); err != nil {
			return nil, err
		}

		resourceList, err := parseKubeResource(chunkData, resource)
		if err != nil {
			return nil, err
		}

		typeList = append(typeList, resourceList...)
	}

	// if document is tooo big
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return typeList, err
}

// transform KubeResourceInterface/KubeResource into concrete class
func parseKubeResource(data []byte, resource KubeResourceInterface) ([]KubeResourceInterface, error) {
	var err error
	var object KubeResourceInterface

	typeList := make([]KubeResourceInterface, 0)

	// TODO: i think, i should need to refactor this..
	switch resource.GetKind() {
	case KIND_LIST: // parse list is little bit tricky..
		listObject := &kubeResourceList{}
		if err = yaml.Unmarshal(data, listObject); err != nil {
			return nil, err
		}
		if len(listObject.Items) > 0 {
			var list kubeResourceListInterface

			switch listObject.Items[0].GetKind() {
			case KIND_NAMESPACE:
				list = &namespaceList{}
				break

			case KIND_DEPLOYMENT:
				list = &deploymentList{}
				break

			case KIND_REPLICASET:
				list = &replicaSetList{}
				break
			}

			if list != nil {
				yaml.Unmarshal(data, list)
				for _, item := range list.GetItems() {
					typeList = append(typeList, item)
				}
			}
		}
		break

	case KIND_DEPLOYMENT: // parse deployment
		object = &Deployment{}
		break

	case KIND_REPLICASET: // parse replicaset object
		object = &ReplicaSet{}
		break

	case KIND_NAMESPACE: // parse namespace object
		object = &Namespace{}
		break
	}

	if object != nil {
		if err = yaml.Unmarshal(data, object); err != nil {
			return nil, err
		}
		typeList = append(typeList, object)
	}

	return typeList, err
}

//
// This piece of code is taken from Kubernetes project:
// https://github.com/kubernetes/kubernetes/blob/b359034817685a8d25bb51bae765308d9d200da0/pkg/util/yaml/decoder.go#L143
// all credits should go to Kubernetes authors
//
const yamlSeparator = "\n---"

func splitYAMLDocument(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	sep := len([]byte(yamlSeparator))
	if i := bytes.Index(data, []byte(yamlSeparator)); i >= 0 {
		// We have a potential document terminator
		i += sep
		after := data[i:]
		if len(after) == 0 {
			// we can't read any more characters
			if atEOF {
				return len(data), data[:len(data)-sep], nil
			}
			return 0, nil, nil
		}
		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i-sep], nil
		}
		return 0, nil, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
