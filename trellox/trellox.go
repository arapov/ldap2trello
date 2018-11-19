package trellox

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Info struct {
	Key   string `json:"key"`
	Token string `json:"token"`
}

type Member struct {
	Active   bool   `json:"active"`
	Fullname string `json:"fullName"`
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (c *Info) Dial() *Info {
	return c
}

func (c *Info) Search(email string) ([]Member, int) {
	var trelloMembers []Member

	httpRequest := fmt.Sprintf("https://api.trello.com/1/search/members?query=%s&key=%s&token=%s", email, c.Key, c.Token)
	httpRes, err := http.Get(httpRequest)
	if err != nil {
		log.Println(err)
		return nil, httpRes.StatusCode
	}

	if httpRes.StatusCode == 429 {
		// TODO: Handle 429 is for Trello API limit exceed
	}

	data, _ := ioutil.ReadAll(httpRes.Body)
	json.Unmarshal(data, &trelloMembers)

	return trelloMembers, httpRes.StatusCode
}
