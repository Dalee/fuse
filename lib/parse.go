package lib

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/ghodss/yaml"
)

func ParseYaml(data []byte) ([]*KubeType, error) {
	var err error

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(splitYAMLDocument)

	typeList := make([]*KubeType, 0)
	err = nil

	for scanner.Scan() {
		var config KubeType

		data := scanner.Bytes()
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			break
		}

		config.Kind = strings.ToLower(config.Kind)
		config.Deployed = false

		switch config.Kind {
		case "deployment":
			typeList = append(typeList, &config)
			break
		}
	}

	return typeList, err
}

//
// https://github.com/kubernetes/kubernetes/blob/b359034817685a8d25bb51bae765308d9d200da0/pkg/util/yaml/decoder.go#L143
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
				return len(data), data[:len(data) - sep], nil
			}
			return 0, nil, nil
		}
		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i - sep], nil
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
