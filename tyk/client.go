package tyk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
)

type Config struct {
	BaseURL      string
	Organisation string
	Key          string
}

type Client struct {
	client *http.Client
	config Config
}

func NewClient(config Config) *Client {
	return &Client{
		config: config,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *Client) CreateKey(token string) error {
	k := NewKey(c.config.Organisation)

	rights, err := c.GetAccessRights()
	if err != nil {
		return err
	}
	k.AccessRights = rights

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(k)

	url := fmt.Sprintf("%s/tyk/keys/%s", c.config.BaseURL, token)
	logrus.Debugf("post keys url: %s", url)
	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		return err
	}

	req.Header.Add("x-tyk-authorization", c.config.Key)
	req.Header.Add("Content-Type", "application/json")

	dump, err := httputil.DumpRequestOut(req, true)

	logrus.Debugf("post request: %q", dump)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	logrus.Debugf("post keys response status: %s", strconv.Itoa(res.StatusCode))

	return nil
}

func (c *Client) GetAccessRights() (map[string]AccessRight, error) {
	url := fmt.Sprintf("%s/tyk/apis/", c.config.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-tyk-authorization", c.config.Key)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	logrus.Debugf("get apis response status: %s", strconv.Itoa(res.StatusCode))

	var apis []API
	err = json.NewDecoder(res.Body).Decode(&apis)
	if err != nil {
		logrus.Debugf(err.Error())
	}

	rights := make(map[string]AccessRight)
	for _, api := range apis {
		rights[api.ID] = AccessRight{
			ID:       api.ID,
			Name:     api.Name,
			Versions: []string{"Default"},
		}
		logrus.Debugf("rights: %s, %s", rights[api.ID].ID, rights[api.ID].Name)
	}

	return rights, nil
}

func (c *Client) DeleteKey(token string) error {
	url := fmt.Sprintf("%s/tyk/keys/%s", c.config.BaseURL, token)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("x-tyk-authorization", c.config.Key)

	if _, err := c.client.Do(req); err != nil {
		return err
	}

	return nil
}
