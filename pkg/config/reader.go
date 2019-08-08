package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/ppacher/webthings-mqtt-gateway/pkg/spec"
)

// ThingFromFile reads a spec.Thing definition in YAML format
// from the given file
func ThingFromFile(fileName string) (*spec.Thing, error) {
	blob, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var t spec.Thing
	if err := yaml.Unmarshal(blob, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

// ReadThingsFromDirectory reads all thing defintion files from
// a given directory
func ReadThingsFromDirectory(dir string) ([]*spec.Thing, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var things []*spec.Thing
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		thing, err := ThingFromFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("%s: %s", file.Name(), err.Error())
		}

		things = append(things, thing)
	}

	return things, nil
}
