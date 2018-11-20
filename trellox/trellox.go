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
	Fullname string `json:"fullname"`
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (c *Info) Dial() *Info {
	return c
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
