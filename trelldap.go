package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/arapov/trelldap/env"
)

const (
	configfile = "config.json"
	datafile   = "members.json"
)

// Meta defines the data required to do the synchronization between
// LDAP and Trello Organization
type Meta struct {
	Fullname       string   `json:"fullname"`
	Mails          []string `json:"mails"`
	TrelloActive   bool     `json:"trello"`
	TrelloID       string   `json:"trelloid"`
	TrelloMail     string   `json:"trellomail"`
	TrelloName     string   `json:"trelloname"`
	TrelloUserName string   `json:"trellouser"`

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
		member, ok := members.Meta[lMember.UID]
		if !ok {
			// TODO: What if we don't want to look for aliases
			// cmd-line parameter
			ldap.GetAliases(lMember)

			member = &Meta{
				Fullname:     lMember.Fullname,
				Mails:        lMember.Mails,
				TrelloActive: true, // default state
			}
		}

		// Mark everyone who is in LDAP, those who end up with
		// false are the material to be removed from Trello.
		member.seenInLDAP = true

		// .TrelloActive is the default state for the new discovered member
		// we try to find member in Trello when the TrelloID is empty.
		if member.TrelloActive && member.TrelloID == "" {
			for _, mail := range lMember.Mails {
			retry:
				tMember, statusCode := trello.Search(mail)
				if statusCode == 429 {
					log.Println("Hit the Trello API limit. Sleeping for 5 minutes.")
					members.Write()
					time.Sleep(5 * time.Minute)
					goto retry
				}

				member.TrelloActive = tMember.Active
				member.TrelloID = tMember.ID
				member.TrelloMail = mail
				member.TrelloName = tMember.Fullname
				member.TrelloUserName = tMember.Username
			}
		}

	}

	if err := members.Write(); err != nil {
		log.Fatalln(err)
	}
}
