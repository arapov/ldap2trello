package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

const (
	configfile = "config.json"
	datafile   = "members.json"
)

// Meta defines the data required to do the synchronization between
// LDAP and Trello Organization
type Meta struct {
	FullName       string `json:"fullname"`
	TrelloActive   bool   `json:"trello"`
	TrelloID       string `json:"trelloid"`
	TrelloUserName string `json:"trellouser"`
	TrelloMail     string `json:"trellomail"`
}

// Members is ssia, map of Meta data
type Members struct {
	Map map[string]*Meta `json:"members"`
}

func (m *Members) Read() error {
	jsonBytes, err := ioutil.ReadFile(datafile)
	json.Unmarshal(jsonBytes, &m)

	return err
}

func (m *Members) Write() error {
	jsonBytes, _ := json.MarshalIndent(m, "", "  ")
	err := ioutil.WriteFile(datafile, jsonBytes, 0644)

	return err
}

func main() {
	c, err := loadConfig(configfile)
	if err != nil {
		log.Fatalln(err)
	}

	var members Members
	if err := members.Read(); err != nil {
		log.Println("no", datafile, "file was found.")
	}

	var ldapMembers map[string]string
	ldapc, err := c.Dial()
	if err != nil {
		log.Fatalln(err)
	}
	ldapMembers = ldapc.Query(c)
	ldapc.Close()

	// TODO: do the stuff!

	if err := members.Write(); err != nil {
		log.Fatalln(err)
	}
}
