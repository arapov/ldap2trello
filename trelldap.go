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

// Meta contains the data required to do the synchronization between
// LDAP and Trello Organization
type Meta struct {
	Fullname string                     `json:"fullname"`
	Mails    []string                   `json:"mails"`
	Trello   map[string]*trellox.Member `json:"trello"`

	seenInLDAP   bool
	seenInTrello bool
}

// TODO: name it, rework it, map of Meta data
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
		members.Meta = make(map[string]*Meta)
		log.Println("no", datafile, "file was found.")
	}

	// trello && ldap connections to work with
	trello := c.Trello.Dial()
	ldap := c.LDAP.Dial()

	// tMembers - represent Trello UIDs, which are the members of Trello Org.
	tMembers := trello.GetOrgBoardMeMemberIDs()
	// lMembers - represent LDAP members, which should be in Trello Org.
	lMembers := ldap.GetMembers()

	for _, lMember := range lMembers {
		if _, ok := members.Meta[lMember.UID]; !ok {
			// We've found new LDAP user we aren't aware of
		reconnect:
			// TODO: What if we don't want to look for aliases
			if err := ldap.GetAliases(lMember); err != nil {
				// Trello API has limits for calls, so that calls are throttled.
				// LDAP connection could be lost, while we wait for Trello API
				// to be available
				ldap = c.LDAP.Dial()
				goto reconnect
			}

			members.Meta[lMember.UID] = &Meta{
				Fullname: lMember.Fullname,
				Mails:    lMember.Mails,
				Trello:   make(map[string]*trellox.Member),
			}
		}

		// member is the common container of LDAPxTrello Member
		member := members.Meta[lMember.UID]

		// Mark member who is in LDAP. Those who end up with false are
		// material to be removed from Trello and cache file.
		member.seenInLDAP = true

		for _, mail := range lMember.Mails {
			if _, ok := member.Trello[mail]; !ok {
				// Newbie has been found
				member.Trello[mail] = trello.Search(mail)
			}

			if _, ok := tMembers[member.Trello[mail].ID]; !ok {
				member.seenInTrello = false
			}
		}
	}

	// TODO: isUseful?
	//	trello.GetOrgID()
	//	trello.GetOrgMembers()

	// Serialize members
	if err := members.Write(); err != nil {
		log.Fatalln(err)
	}
}
