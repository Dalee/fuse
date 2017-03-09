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
		parseYaml(data []byte) (ResourceList, error)
	}

	kubeResourceParser struct {
	}
)

func newParser() kubeResourceParserInterface {
	return &kubeResourceParser{}
}

// ParseLocalFile will allow to parse local file and fetch all resources defined there
func ParseLocalFile(filename string) (ResourceList, error) {
	file, _ := filepath.Abs(filename)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return newParser().parseYaml(data)
}

// parse whole kubectl answer into list of objects
func (p *kubeResourceParser) parseYaml(data []byte) (ResourceList, error) {
	typeList := make(ResourceList, 0)
	maxBufferSize := 1024 * 1024 * 200 // should be enough

	binaryReader := bytes.NewReader(data)
	scanner := bufio.NewScanner(binaryReader)

	scanner.Buffer(data, maxBufferSize)
	scanner.Split(splitYAMLDocument)

	for scanner.Scan() {
		resource := &kubeResource{}
		chunkData := scanner.Bytes()

		if err := yaml.Unmarshal(chunkData, resource); err != nil {
			return nil, err
		}

		resourceList, _ := parseKubeResource(chunkData, resource)
		typeList = append(typeList, resourceList...)
	}

	// if document is tooo big, even bigger than MaxBuffSize
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return typeList, nil
}

// transform KubeResourceInterface/KubeResource into concrete class
func parseKubeResource(data []byte, resource KubeResourceInterface) (ResourceList, error) {
	var err error
	var object KubeResourceInterface

	typeList := make(ResourceList, 0)

	// TODO: i think, i should refactor this..
	switch resource.GetKind() {
	case KindList: // parse list is little bit tricky..
		listObject := &kubeResourceList{}
		yaml.Unmarshal(data, listObject)

		// but, we know that list can't have mixed resource type
		// packed in single list, so, check first item kind and create
		// appropriate storage type
		if len(listObject.Items) > 0 {
			var resourceList kubeResourceListInterface

			switch listObject.Items[0].GetKind() {

			case KindPod:
				resourceList = &podList{}
				break

			case KindNamespace:
				resourceList = &namespaceList{}
				break

			case KindDeployment:
				resourceList = &deploymentList{}
				break

			case KindReplicaSet:
				resourceList = &replicaSetList{}
				break
			}

			if resourceList != nil {
				yaml.Unmarshal(data, resourceList)
				for _, item := range resourceList.GetItems() {
					typeList = append(typeList, item)
				}
			}
		}
		break

	case KindPod:
		object = &Pod{}
		break

	case KindDeployment: // parse deployment
		object = &Deployment{}
		break

	case KindReplicaSet: // parse replicaset object
		object = &ReplicaSet{}
		break

	case KindNamespace: // parse namespace object
		object = &Namespace{}
		break
	}

	if object != nil {
		yaml.Unmarshal(data, object)
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
