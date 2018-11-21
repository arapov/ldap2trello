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

type Members struct {
	Filter  string `json:"filter"`
	BaseDN  string `json:"baseDN"`
	Attribs struct {
		UID      string `json:"uid"`
		Fullname string `json:"fullname"`
		Mail     string `json:"mail"`
	} `json:"attributes"`
}

type Aliases struct {
	Filter  string `json:"filter"`
	BaseDN  string `json:"baseDN"`
	Attribs struct {
		Once string `json:"once"`
		Mail string `json:"mail"`
	} `json:"attributes"`
}

// Info keeps LDAP setting
type Info struct {
	Host     string `json:"hostname"`
	Port     string `json:"port"`
	BindDN   string `json:"bindDN"`
	Password string `json:"password"`

	Members `json:"members"`
	Aliases `json:"aliases"`
}

// Conn represents an LDAP Connection
type Conn struct {
	*ldap.Conn

	*Members
	*Aliases
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

	if c.BindDN != "" {
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
		Conn:    pConn,
		Members: &c.Members,
		Aliases: &c.Aliases,
	}
}

// TODO: Generalize all hardcoded strings

func (c *Conn) GetAliases(ldapMember *Member) error {
	filter := strings.Replace(c.Aliases.Filter, "<uid>", ldapMember.UID, 1)

	mailAttr := c.Aliases.Attribs.Mail
	onceAttr := c.Aliases.Attribs.Once // this is pure nonsense, though keep it

	searchRequest := ldap.NewSearchRequest(
		c.Aliases.BaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{onceAttr, mailAttr},
		nil,
	)

	ldapRes, err := c.Search(searchRequest)
	if err != nil {
		return err
	}

	for _, entry := range ldapRes.Entries {
		if len(entry.GetAttributeValues(onceAttr)) > 1 {
			continue
		}
		ldapMember.Mails = append(ldapMember.Mails, entry.GetAttributeValue(mailAttr))
	}

	return nil
}

// Query TBD
func (c *Conn) GetMembers() []*Member {
	var ldapMembers []*Member

	uidAttr := c.Members.Attribs.UID
	fullnameAttr := c.Members.Attribs.Fullname
	mailAttr := c.Members.Attribs.Mail

	searchRequest := ldap.NewSearchRequest(
		c.Members.BaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		c.Members.Filter,
		[]string{uidAttr, fullnameAttr, mailAttr},
		nil,
	)
	ldapRes, err := c.Search(searchRequest)
	if err != nil {
		log.Fatalln(err)
	}

	for _, entry := range ldapRes.Entries {
		uid := entry.GetAttributeValue(uidAttr)
		fullname := entry.GetAttributeValue(fullnameAttr)
		mail := entry.GetAttributeValue(mailAttr)

		ldapMembers = append(ldapMembers, &Member{uid, fullname, []string{mail}})
	}

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
