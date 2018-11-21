package ldapx

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
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

type Member struct {
	UID      string
	Fullname string
	Mails    []string
}

// Dial connects to the given address on the given network
// and then returns a new Conn for the connection.
func (c *Info) Dial() *Conn {
	pConn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%s", c.Host, c.Port))
	if err != nil {
		log.Fatalln(err)
	}

	if c.Secure {
		if c.Password == "" {
			c.Password = askPassword()
		}

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
	}
}

// TODO: Generalize all hardcoded strings

func (c *Conn) GetAliases(ldapMember *Member) error {
	filter := fmt.Sprintf("(sendmailMTAAliasValue=%s)", ldapMember.UID)

	searchRequest := ldap.NewSearchRequest(
		"ou=mx,dc=redhat,dc=com",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"sendmailMTAAliasValue", "rhatEmailAddress"},
		nil,
	)

	ldapRes, err := c.Search(searchRequest)
	if err != nil {
		return err
	}

	for _, entry := range ldapRes.Entries {
		if len(entry.GetAttributeValues("sendmailMTAAliasValue")) > 1 {
			continue
		}
		ldapMember.Mails = append(ldapMember.Mails, entry.GetAttributeValue("rhatEmailAddress"))
	}

	return nil
}

// Query TBD
func (c *Conn) GetMembers() []*Member {
	var ldapMembers []*Member

	searchRequest := ldap.NewSearchRequest(
		c.baseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		c.filter,
		[]string{"uid", "cn", "mail"},
		nil,
	)
	ldapRes, err := c.Search(searchRequest)
	if err != nil {
		log.Fatalln(err)
	}

	for _, entry := range ldapRes.Entries {
		uid := entry.GetAttributeValue("uid")
		fullname := entry.GetAttributeValue("cn")
		mail := entry.GetAttributeValue("mail")

		ldapMembers = append(ldapMembers, &Member{uid, fullname, []string{mail}})
	}

	// TODO: LDAP folks may have mail aliases they use for Trello

	return ldapMembers
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
