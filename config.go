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
	// Trello configuration
	Trello struct {
		Key   string `json:"key"`
		Token string `json:"token"`
	} `json:"trello"`

	// LDAP configuration
	LDAP struct {
		Host     string `json:"hostname"`
		Port     string `json:"port"`
		Secure   bool   `json:"secure"`
		BindDN   string `json:"bindDN"`
		Password string `json:"password"`
		Filter   string `json:"filter"`
		BaseDN   string `json:"baseDN"`
	} `json:"ldap"`
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
