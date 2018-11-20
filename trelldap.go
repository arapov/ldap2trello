package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/arapov/trelldap/trellox"

	"github.com/arapov/trelldap/env"
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
			// TODO: What if we don't want to look for aliases
			// cmd-line parameter
			ldap.GetAliases(lMember) // LDAP connection may die, while waiting Trello API

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
			for _, mail := range lMember.Mails {

			retry: // TODO: move this to trellox.go
				tMember, statusCode := trello.Search(mail)
				if statusCode == 429 {
					log.Println("Trello API limit has been exceeded. Sleeping for 5 minutes.")
					members.Write()
					time.Sleep(5 * time.Minute)
					goto retry
				}

				member.Trello[mail] = tMember
			}
		}

	}

	if err := members.Write(); err != nil {
		log.Fatalln(err)
	}
}
