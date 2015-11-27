package proj

import (
	"errors"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
)

// Convert a JSON parse into data structures that would be produced by
// encodings/json
func yaml_to_json(node yaml.Node) interface{} {
	if scalar, ok := node.(yaml.Scalar); ok {
		return scalar.String()
	} else if list, ok := node.(yaml.List); ok {
		rv := make([]interface{}, list.Len())
		for i, elt := range list {
			rv[i] = yaml_to_json(elt)
		}
		return rv
	} else if hash, ok := node.(yaml.Map); ok {
		rv := make(map[string]interface{})
		for k, elt := range hash {
			rv[k] = yaml_to_json(elt)
		}
		return rv
	} else {
		log.Fatal("invalid data for yaml_to_json")
		return nil
	}
}

// Utility function to get a child of a JSON node, or if the JSON node is not a
// Map, assume that it is the expected value.  This allows things like "cd:
// somedir" as a shorthand for "cd: dir: somedir".
func default_child(args interface{}, key string) (interface{}, bool) {
	m, ok := args.(map[string]interface{})
	if !ok {
		return args, true
	} else {
		rv, ok := m[key]
		return rv, ok
	}
}

// Utility function to extract the string value of a JSON node
func node_string(node interface{}) string {
	str, ok := node.(string)
	if !ok {
		log.Fatalf("Expected a string, got %#v", node)
	}
	return str
}

// Expect a JSON map with a single key and break that out into the key and its
// value.  This is a common structure in YAML files.
func singleKeyMap(input interface{}) (string, interface{}, error) {
	input_map, ok := input.(map[string]interface{})
	if !ok || len(input_map) != 1 {
		return "", nil, errors.New("expected a single-propery object")
	}
	for key, args := range input_map {
		return key, args, nil
	}
	return "", nil, nil
}
