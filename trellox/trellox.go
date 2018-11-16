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
	FullName string `json:"fullName"`
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (c *Info) Dial() *Info {
	return c
}

func (c *Info) Test(uid string) []Member {
	var trelloMember []Member

	httpRequest := fmt.Sprintf("https://api.trello.com/1/search/members?query=%s@redhat.com&key=%s&token=%s", uid, c.Key, c.Token)
	httpRes, err := http.Get(httpRequest)
	if err != nil {
		// TODO: save the state
		log.Fatalln(err)
	}

	// 429

	data, _ := ioutil.ReadAll(httpRes.Body)
	json.Unmarshal(data, &trelloMember)

	return trelloMember
}
