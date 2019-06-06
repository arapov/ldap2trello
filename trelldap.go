package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/arapov/trelldap/env"
	"github.com/arapov/trelldap/trellox"
)

const (
	configfile = "config.json"
	datafile   = "members.json"
)

// Meta defines the data required to do the synchronization between
// LDAP and Trello Organization
type Meta struct {
	Fullname string                     `json:"fullname"`
	Mails    []string                   `json:"mails"`
	Trello   map[string]*trellox.Member `json:"trello"`

	seenInLDAP bool
}

// Members is ssia, map of Meta data
type Members struct {
	Meta map[string]*Meta `json:"members"`
}

func (m *Members) Read() error {
	jsonBytes, err := ioutil.ReadFile(datafile)
	json.Unmarshal(jsonBytes, &m)

	return err
}

func (m *Members) Write() error {
	jsonBytes, _ := json.MarshalIndent(m, "", "  ")
	err := ioutil.WriteFile(datafile, jsonBytes, 0640)

	return err
}

func main() {
	c, err := env.LoadConfig(configfile)
	if err != nil {
		log.Fatalln(err)
	}

	// members - TODO: document the importance
	var members Members
	if err := members.Read(); err != nil {
		log.Println("no", datafile, "file was found.")

		members.Meta = make(map[string]*Meta)
	}

	// trello && ldap connections to work with
	trello := c.Trello.Dial()
	ldap := c.LDAP.Dial()

	// Add newly discovered in LDAP People to 'members'
	lMembers := ldap.GetMembers()
	for _, lMember := range lMembers {
		if _, ok := members.Meta[lMember.UID]; !ok {
		reconnect:
			// TODO: What if we don't want to look for aliases
			if err := ldap.GetAliases(lMember); err != nil {
				ldap = c.LDAP.Dial()
				goto reconnect
			}

			members.Meta[lMember.UID] = &Meta{
				Fullname: lMember.Fullname,
				Mails:    lMember.Mails,
				Trello:   make(map[string]*trellox.Member),
			}
		}
		member := members.Meta[lMember.UID]

		// Mark everyone who is in LDAP, those who end up with
		// false are the material to be removed from Trello.
		member.seenInLDAP = true

		if _, ok := member.Trello[lMember.Mails[0]]; !ok {
			// Newbie has been found
			for _, mail := range lMember.Mails {
				member.Trello[mail] = trello.Search(mail)
			}
		}

	}

	trello.GetOrgID()
	trello.GetOrgMembers()

	// list missing in ldap people
	for k, v := range members.Meta {
		if v.seenInLDAP == false {
			log.Println(k, v.seenInLDAP)
		}
	}

	if err := members.Write(); err != nil {
		log.Fatalln(err)
	}
}
