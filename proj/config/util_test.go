package config

import (
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"testing"
)

type yamlToJsonTest struct {
	input  string
	output string
	err    error
}

var yamlToJsonTests = []yamlToJsonTest{
	{"x: y", "map[\"x\":\"y\"]", nil},
	{"- y\n- z", "[\"y\" \"z\"]", nil},
	{"- 1\n- 2", "[\"1\" \"2\"]", nil}, // all strings
}

func TestYamlToJson(t *testing.T) {
	for _, y2jt := range yamlToJsonTests {
		input := yaml.Config(y2jt.input).Root
		res, err := yamlToJson(input)
		if err != nil {
			if err != y2jt.err {
				t.Errorf("%s: got error %q, expected error %q",
					y2jt.input, err, y2jt.err)
			}
		} else {
			output := fmt.Sprintf("%q", res)
			if output != y2jt.output {
				t.Errorf("%s: got output %s, expected output %s",
					y2jt.input, output, y2jt.output)
			}
		}
	}
}
