package config

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
