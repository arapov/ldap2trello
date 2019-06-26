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

// Meta contains the data required for LDAP and Trello synchronization
type Meta struct {
	Fullname  string   `json:"fullname"`
	Mails     []string `json:"mails"`
	TrelloID  []string `json:"trelloid"`
	Timestamp int64    `json:"timestamp"`

	seenInLDAP   bool
	seenInTrello bool
}

func main() {
	// members is the serialized cache that contains the state of the last run, where
	// key is member ldap uid
	var members map[string]*Meta

	c, err := env.LoadConfig(configfile)
	if err != nil {
		log.Fatalln(err)
	}

	if err := read(&members); err != nil {
		members = make(map[string]*Meta)
		log.Println("no", datafile, "file was found.")
	}

	ldap := c.LDAP.Dial()
	ldapMembers := ldap.GetMembers()
	for _, lMember := range ldapMembers {
		if _, ok := members[lMember.UID]; !ok {
			log.Println("New person has been found since last run:", lMember.UID)
			// We've found new LDAP user we aren't aware of
			if err := ldap.GetAliases(lMember); err != nil {
				log.Println(err)
			}

			members[lMember.UID] = &Meta{
				Fullname:  lMember.Fullname,
				Mails:     lMember.Mails,
				TrelloID:  nil,
				Timestamp: 0,
			}
		}
		// Mark member who is in LDAP. Those who end up with false are
		// material to be removed from Trello and cache file.
		members[lMember.UID].seenInLDAP = true
	}

	trello := c.Trello.Dial()
	for _, lMember := range ldapMembers {
		// member is the common container of LDAPxTrello Member
		member := members[lMember.UID]

		for _, mail := range lMember.Mails {
			// TODO: figure out the delta from now() for retry
			if member.Timestamp == 0 {
				// Newbie has been found
				tMember := trello.Search(mail)
				member.TrelloID = append(member.TrelloID, tMember.ID)
				member.Timestamp = time.Now().Unix()
			}
		}
	}

	// TODO: isUseful?
	//	trello.GetOrgID()

	// TODO:
	// 1. unsubscribe .seenInLDAP = false from Org
	// -- DELETE /boards/{id}/members/{idMember}
	// -- DELETE /organizations/{id}/members/{idMember}
	// 1.a. remove from cache
	// -- remove entity from members/Meta
	// 2. subscribe .seenInTrello = false to Org
	// -- PUT /organizations/{id}/members/{idMember}
	// 2.a. concern trello username vs multiple mails
	for _, member := range members {
		log.Println(member.Fullname, member.seenInLDAP, member.seenInTrello)
	}

	// Serialize members
	if err := write(&members); err != nil {
		log.Fatalln(err)
	}
}

func read(m *map[string]*Meta) error {
	jsonBytes, err := ioutil.ReadFile(datafile)
	json.Unmarshal(jsonBytes, m)

	return err
}

func write(m *map[string]*Meta) error {
	jsonBytes, _ := json.MarshalIndent(m, "", "  ")
	err := ioutil.WriteFile(datafile, jsonBytes, 0640)

	return err
}
