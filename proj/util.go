package proj

import (
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
)

// Convert a JSON parse into data structures that would be produced by
// encodings/json
func yamlToJson(node yaml.Node) (interface{}, error) {
	var err error
	if scalar, ok := node.(yaml.Scalar); ok {
		return scalar.String(), nil
	} else if list, ok := node.(yaml.List); ok {
		rv := make([]interface{}, list.Len())
		for i, elt := range list {
			rv[i], err = yamlToJson(elt)
			if err != nil {
				return nil, err
			}
		}
		return rv, nil
	} else if hash, ok := node.(yaml.Map); ok {
		rv := make(map[string]interface{})
		for k, elt := range hash {
			rv[k], err = yamlToJson(elt)
			if err != nil {
				return nil, err
			}
		}
		return rv, nil
	} else {
		return nil, fmt.Errorf("invalid data for yamlToJson")
	}
}

// Utility function to get a child of a JSON node, or if the JSON node is not a
// Map, assume that it is the expected value.  This allows things like "cd:
// somedir" as a shorthand for "cd: dir: somedir".
func defaultChild(args interface{}, key string) (interface{}, bool) {
	m, ok := args.(map[string]interface{})
	if !ok {
		return args, true
	} else {
		rv, ok := m[key]
		return rv, ok
	}
}

// Expect a JSON map with a single key and break that out into the key and its
// value.  This is a common structure in YAML files.
func singleKeyMap(input interface{}) (string, interface{}, error) {
	inputMap, ok := input.(map[string]interface{})
	if !ok || len(inputMap) != 1 {
		return "", nil, fmt.Errorf("expected a single-propery object")
	}
	for key, args := range inputMap {
		return key, args, nil
	}
	return "", nil, nil
}
