package trellox

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Info struct {
	Key          string `json:"key"`
	Token        string `json:"token"`
	Organization string `json:"organization"`
}

type Member struct {
	Active        bool     `json:"active"`
	Fullname      string   `json:"fullName"`
	ID            string   `json:"id"`
	Username      string   `json:"username"`
	Organizations []string `json:"idOrganizations"`
}

func (c *Info) Dial() *Info {
	return c
}

func (c *Info) GetOrgID() string {
	var tOrganization struct {
		ID string `json:"id"`
	}

	httpRequest := fmt.Sprintf("https://api.trello.com/1/organizations/%s?key=%s&token=%s", c.Organization, c.Key, c.Token)
	httpRes, err := http.Get(httpRequest)
	if err != nil {
		log.Fatalln(err)
	}

	data, _ := ioutil.ReadAll(httpRes.Body)
	json.Unmarshal(data, &tOrganization)

	return tOrganization.ID
}

func (c *Info) Search(email string) (*Member, int) {
	var tMembers []Member

	httpRequest := fmt.Sprintf("https://api.trello.com/1/search/members?query=%s&key=%s&token=%s&limit=1", email, c.Key, c.Token)
	httpRes, err := http.Get(httpRequest)
	if err != nil {
		log.Fatalln(err)
	}

	if httpRes.StatusCode == 429 {
		return &Member{}, httpRes.StatusCode
	}

	data, _ := ioutil.ReadAll(httpRes.Body)
	json.Unmarshal(data, &tMembers)

	return &tMembers[0], httpRes.StatusCode
}
