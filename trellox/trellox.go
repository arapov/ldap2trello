package trellox

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	trelloApiURL = "https://api.trello.com/1"
)

type Info struct {
	Key          string `json:"key"`
	Token        string `json:"token"`
	Organization string `json:"organization"`
	BoardID      string `json:"boardid"`
}

type Member struct {
	Fullname      string   `json:"fullName"`
	ID            string   `json:"id"`
	Username      string   `json:"username"`
	Organizations []string `json:"idOrganizations"`
}

func (c *Info) Dial() *Info {
	return c
}

func (c *Info) callAPI(api string, v interface{}) {
	auth := fmt.Sprintf("key=%s&token=%s", c.Key, c.Token)
	if strings.Contains(api, "?") {
		auth = fmt.Sprintf("&%s", auth)
	} else {
		auth = fmt.Sprintf("?%s", auth)
	}

retry:
	httpRequest := fmt.Sprintf("%s%s%s", trelloApiURL, api, auth)
	httpRes, err := http.Get(httpRequest)
	if err != nil {
		log.Fatalln(err)
	}

	if httpRes.StatusCode == 429 {
		// TODO: members.Write()
		log.Println("Trello API limit has been reached. Sleeping for 5 minutes.")
		time.Sleep(5 * time.Minute)
		goto retry
	}

	data, _ := ioutil.ReadAll(httpRes.Body)
	json.Unmarshal(data, &v)
}

func (c *Info) GetOrgID() string {
	var tOrganization struct {
		ID string `json:"id"`
	}

	api := fmt.Sprintf("/organization/%s", c.Organization)
	c.callAPI(api, &tOrganization)

	return tOrganization.ID
}

func (c *Info) GetBoardMembers() map[string]struct{} {
	var res []struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
		UserName string `json:"username"`
	}

	api := fmt.Sprintf("/boards/%s/members", c.BoardID)
	c.callAPI(api, &res)

	return mapify(res, "id")
}

func (c *Info) GetMembers() map[string]struct{} {
	var res []struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
		UserName string `json:"username"`
	}

	api := fmt.Sprintf("/organizations/%s/members", c.Organization)
	c.callAPI(api, &res)

	return mapify(res, "id")
}

func (c *Info) Search(email string) *Member {
	var tMembers []Member

	api := fmt.Sprintf("/search/members?query=%s&limit=1", email)
	c.callAPI(api, &tMembers)

	return &tMembers[0]
}

func mapify(structure interface{}, key string) map[string]struct{} {
	res := make(map[string]struct{})

	var records []map[string]interface{}
	jsonByte, _ := json.Marshal(structure)
	json.Unmarshal(jsonByte, &records)

	for _, record := range records {
		v := record[key].(string)
		res[v] = struct{}{}
	}

	return res
}
