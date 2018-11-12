package chroma

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const URI = "http://localhost:54235/razer/chromasdk"

// Client for communicating with the Razer Chroma REST API.
type Client struct {
	uri       string
	sessionId int
	cli       *http.Client
}

func NewClient() *Client {
	return &Client{cli: &http.Client{}}
}

// Register the application with Chroma server.
// The Client is unusable until this method is called.
func (c *Client) Register() {
	body := `{
		"title": "Chroma Reflect",
		"description": "Chroma Reflect",
		"author": {
			"name": "Edward Moore",
			"contact": "github.com/edjmore"
		},
		"device_supported": [
			"keyboard"
		],
		"category": "application"
	}`
	req, err := http.NewRequest(http.MethodPost, URI, bytes.NewBuffer([]byte(body)))
	checkError(err)

	req.Header.Add("content-type", "application/json")
	resp, err := c.cli.Do(req)
	checkError(err)

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	checkError(err)

	session := struct {
		Uri string
		Id  int `json:"sessionid"`
	}{}
	err = json.Unmarshal([]byte(respBody), &session)
	checkError(err)

	c.uri = session.Uri
	c.sessionId = session.Id
	log.Printf("client registered: %d", c.sessionId)
}

// Set all keys to one color (in BGR format).
func (c *Client) SetStaticColor(color int) {
	body := `{
		"effect": "CHROMA_STATIC",
		"param": {
			"color": %d
		}
	}`
	body = fmt.Sprintf(body, color)
	req, err := http.NewRequest(http.MethodPut, c.uri+"/keyboard", bytes.NewBuffer([]byte(body)))
	checkError(err)

	req.Header.Add("content-type", "application/json")
	resp, err := c.cli.Do(req)
	checkError(err)

	defer resp.Body.Close()
	checkResult(resp)
}

// Set custom colors for each key.
// Each element of colors is a key color in BGR format.
func (c *Client) SetCustom(colors [6][22]int) {
	body := `{
		"effect": "CHROMA_CUSTOM",
		"param": %s
	}`
	b, err := json.Marshal(colors)
	checkError(err)
	body = fmt.Sprintf(body, b)
	req, err := http.NewRequest(http.MethodPut, c.uri+"/keyboard", bytes.NewBuffer([]byte(body)))
	checkError(err)

	req.Header.Add("content-type", "application/json")
	resp, err := c.cli.Do(req)
	checkError(err)

	defer resp.Body.Close()
	checkResult(resp)
}

// Unregister the application so the Chroma server can free related resources.
func (c *Client) Unregister() {
	req, err := http.NewRequest(http.MethodDelete, c.uri, nil)
	checkError(err)

	req.Header.Add("content-type", "application/json")
	resp, err := c.cli.Do(req)
	checkError(err)

	defer resp.Body.Close()
	checkResult(resp)
	log.Printf("client unregistered: %d", c.sessionId)
}

// Check the result of a Chroma API request.
// If the result is non-zero, then an error occured.
func checkResult(resp *http.Response) {
	body, err := ioutil.ReadAll(resp.Body)
	checkError(err)

	res := struct {
		Result int
	}{}
	err = json.Unmarshal(body, &res)
	checkError(err)

	if res.Result != 0 {
		panic(fmt.Errorf("chroma result: %d", res.Result))
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
