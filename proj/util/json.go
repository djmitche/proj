package util

import (
	"fmt"
	"github.com/spf13/cast"
)

// Utility function to get a child of a JSON node, or if the JSON node is not a
// Map, assume that it is the expected value.  This allows things like "cd:
// somedir" as a shorthand for "cd: dir: somedir".
func DefaultChild(args interface{}, key string) (interface{}, bool) {
	m, ok := args.(map[string]interface{})
	if !ok {
		return args, true
	} else {
		rv, ok := m[key]
		return rv, ok
	}
}

// Expect a Viper map with a single key and break that out into the key and its
// value.  This is a common structure in YAML files.
func SingleKeyMap(input interface{}) (string, interface{}, error) {
	inputMap, err := cast.ToStringMapE(input)
	if err != nil {
		return "", nil, fmt.Errorf("expected a single-key map: %s", err)
	}
	if len(inputMap) != 1 {
		return "", nil, fmt.Errorf("expected a single-key map; got %d keys", len(inputMap))
	}
	for key, args := range inputMap {
		return key, args, nil
	}
	return "", nil, nil
}
