// Package env reads application settings
package env

import (
	"encoding/json"
	"os"
	"regexp"

	"github.com/arapov/trelldap/jsonconfig"
	"github.com/arapov/trelldap/ldapx"
	"github.com/arapov/trelldap/trellox"
)

// Info contains the configuration information required for application
// to work.
type Info struct {
	Trello *trellox.Info `json:"trello"`
	LDAP   *ldapx.Info   `json:"ldap"`
}

// ParseJSON unmarshals bytes to structs.
func (c *Info) ParseJSON(b []byte) error {
	// w is always "env:*" here, hence [4:]
	Getenv := func(w string) string {
		return os.Getenv(w[4:])
	}

	// Looking for all the "env:*" strings to replace with ENV vars
	rx := regexp.MustCompile(`env:[A-Z_]+`)
	newB := rx.ReplaceAllStringFunc(string(b), Getenv)

	return json.Unmarshal([]byte(newB), &c)
}

// LoadConfig reads the configuration file.
func LoadConfig(filename string) (*Info, error) {
	config := &Info{}

	err := jsonconfig.Load(filename, config)

	return config, err
}
