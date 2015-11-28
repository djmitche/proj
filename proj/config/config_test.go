package config

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func makeFile(content string) (filename string, cleanup func()) {
	dir, err := ioutil.TempDir("", "proj-test")
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(filepath.Join(dir, "config-test.yml"),
		os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		panic(err)
	}

	_, err = f.Write([]byte(content))
	if err != nil {
		panic(err)
	}
	f.Close()

	return f.Name(), func() {
		err := os.RemoveAll(dir)
		if err != nil {
			panic(err)
		}
	}
}

func TestLoadProjectConfiFromFileEmpty(t *testing.T) {
	filename, cleanup := makeFile("")
	defer cleanup()

	config, err := loadProjectConfigFromFile(filename)
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, len(config.Children), 0, "no children")
	assert.Equal(t, len(config.Modifiers), 0, "no children")
}

func TestLoadProjectConfiFromFileExample(t *testing.T) {
	filename, cleanup := makeFile(`
children:
    proj1:  # note that spaces are important here, apparently
        cd: code/proj1
    proj2:
        cd:
            dir: code/proj2

# shell modifiers
shell:
  - mod1: a
  - mod2: 10
`)
	defer cleanup()

	config, err := loadProjectConfigFromFile(filename)
	if err != nil {
		t.Error(err)
		return
	}

	assert := assert.New(t)

	if assert.Equal(len(config.Children), 2, "two children") {
		assert.Equal(config.Children["proj1"].Type, "cd", "first child is cd")
		assert.Equal(config.Children["proj1"].Args,
			interface{}("code/proj1"),
			"first child arg raw dir")
		assert.Equal(config.Children["proj2"].Type, "cd", "second child is cd")
		assert.Equal(config.Children["proj2"].Args,
			interface{}(map[interface{}]interface{}{"dir": "code/proj2"}),
			"second child arg dir arg")
	}

	if assert.Equal(len(config.Modifiers), 2, "two children") {
		assert.Equal(config.Modifiers[0],
			interface{}(map[interface{}]interface{}{"mod1": "a"}),
			"first modifier maps mod1 to a")
		assert.Equal(config.Modifiers[1],
			interface{}(map[interface{}]interface{}{"mod2": 10}),
			"second modifier maps mod2 to 10 (int)")
	}
}
