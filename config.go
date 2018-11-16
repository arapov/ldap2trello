package main

import (
	"encoding/json"
	"os"
	"regexp"

	"github.com/arapov/trelldap/jsonconfig"
)

// Config contains the configuration information required for application
// to work.
type Config struct {
	TrelloKey    string `json:"trelloKey"`
	TrelloToken  string `json:"trelloToken"`
	LDAPBindDN   string `json:"ldapBindDN"`
	LDAPPassword string `json:"ldapPassword"`
}

// ParseJSON unmarshals bytes to structs.
func (cfg *Config) ParseJSON(b []byte) error {
	// w is always "env:*" here, hence [4:]
	Getenv := func(w string) string {
		return os.Getenv(w[4:])
	}

	// Looking for all the "env:*" strings to replace with ENV vars
	rx := regexp.MustCompile(`env:[A-Z_]+`)
	newB := rx.ReplaceAllStringFunc(string(b), Getenv)

	return json.Unmarshal([]byte(newB), &cfg)
}

func loadConfig(filename string) (*Config, error) {
	config := &Config{}

	err := jsonconfig.Load(filename, config)

	return config, err
}
