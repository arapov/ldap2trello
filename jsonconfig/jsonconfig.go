// Package jsonconfig handles loading a JSON file into a struct.
package jsonconfig

import (
	"io/ioutil"
)

// Parser must implement ParseJSON.
type Parser interface {
	ParseJSON([]byte) error
}

// Load the JSON config file.
func Load(configFile string, p Parser) error {
	// Read the config file
	jsonBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	// Parse the config
	if err := p.ParseJSON(jsonBytes); err != nil {
		return err
	}

	return nil
}
