package proj

import (
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
)

// Convert a YAML parse into data structures that would be produced by
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

// utility function to get a child of a YAML node, or if the YAML node is not a
// Map, assume that it is the expected value.  This allows things like "cd:
// somedir" as a shorthand for "cd: dir: somedir".
func default_child(args yaml.Node, key string) (yaml.Node, error) {
	_, ok := args.(yaml.Map)
	if !ok {
		return args, nil
	} else {
		return yaml.Child(args, key)
	}
}

// Utility function to extract the string value of a YAML node
func node_string(node yaml.Node) string {
	scalar, ok := node.(yaml.Scalar)
	if !ok {
		log.Fatalf("Expected a string, got %#v", node)
	}
	return scalar.String()
}
