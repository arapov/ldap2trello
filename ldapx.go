package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
	ldap "gopkg.in/ldap.v2"
)

// Conn represents an LDAP Connection
type Conn struct {
	*ldap.Conn
}

// Query TBD
func (ldapc *Conn) Query(c *Config) map[string]string {

	searchRequest := ldap.NewSearchRequest(
		c.LDAP.BaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		c.LDAP.Filter,
		[]string{"uid", "cn"},
		nil,
	)
	ldapRes, err := ldapc.Search(searchRequest)
	if err != nil {
		log.Fatalln(err)
	}

	result := make(map[string]string)
	for _, entry := range ldapRes.Entries {
		uid := entry.GetAttributeValue("uid")
		fullname := entry.GetAttributeValue("cn")

		result[uid] = fullname
	}

	// TODO: LDAP folks may have mail aliases they use for Trello

	return result
}

// Dial connects to the given address on the given network
// and then returns a new Conn for the connection.
func (cfg *Config) Dial() (*Conn, error) {
	pConn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%s", cfg.LDAP.Host, cfg.LDAP.Port))
	if err != nil {
		return nil, err
	}

	if cfg.LDAP.Secure {
		if cfg.LDAP.Password == "" {
			cfg.LDAP.Password = askPassword()
		}

		if err := pConn.StartTLS(&tls.Config{ServerName: cfg.LDAP.Host}); err != nil {
			log.Fatalln(err)
		}

		if err := pConn.Bind(cfg.LDAP.BindDN, cfg.LDAP.Password); err != nil {
			log.Fatalln(err)
		}
	}

	return &Conn{Conn: pConn}, err
}

func askPassword() string {
	fmt.Print("LDAP Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Println(err)
	}
	password := string(bytePassword)
	fmt.Println()

	return strings.TrimSpace(password)
}
