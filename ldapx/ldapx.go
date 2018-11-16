package ldapx

import (
	"crypto/tls"
	"fmt"
	"log"

	ldap "gopkg.in/ldap.v2"
)

// Info keeps LDAP setting
type Info struct {
	Host     string `json:"hostname"`
	Port     string `json:"port"`
	Secure   bool   `json:"secure"`
	BindDN   string `json:"bindDN"`
	Password string `json:"password"`
	Filter   string `json:"filter"`
	BaseDN   string `json:"baseDN"`
}

// Conn represents an LDAP Connection
type Conn struct {
	*ldap.Conn

	baseDN string
	filter string
}

// Dial connects to the given address on the given network
// and then returns a new Conn for the connection.
func (c *Info) Dial() (*Conn, error) {
	pConn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%s", c.Host, c.Port))
	if err != nil {
		return nil, err
	}

	if c.Secure {
		if err := pConn.StartTLS(&tls.Config{ServerName: c.Host}); err != nil {
			log.Fatalln(err)
		}

		if err := pConn.Bind(c.BindDN, c.Password); err != nil {
			log.Fatalln(err)
		}
	}

	return &Conn{
		Conn:   pConn,
		baseDN: c.BaseDN,
		filter: c.Filter,
	}, err
}

// Query TBD
func (c *Conn) Query() map[string]string {

	searchRequest := ldap.NewSearchRequest(
		c.baseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		c.filter,
		[]string{"uid", "cn"},
		nil,
	)
	ldapRes, err := c.Search(searchRequest)
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
