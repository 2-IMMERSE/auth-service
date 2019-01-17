// Copyright 2018 Cisco and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"
)

type KongUser struct {
	ID        string `json:"id"`
	CustomID  string `json:"custom_id"`
	CreatedAt int    `json:"created_at"`
}

type KongClientList struct {
	Clients []*Client `json:"applications"`
}

// Kong is an authentication driver
type Kong struct {
	api    string
	admin  string
	key    string
	client *http.Client
}

// NewKong creates a new instance of the Kong auth driver
func NewKong() (*Kong, error) {
	api := os.Getenv("KONG_API")
	if len(api) == 0 {
		return nil, fmt.Errorf("KONG_API not set")
	}

	admin := os.Getenv("KONG_ADMIN")
	if len(admin) == 0 {
		return nil, fmt.Errorf("KONG_ADMIN not set")
	}

	return &Kong{
		api:    api,
		admin:  admin,
		key:    os.Getenv("KONG_ACCESS_KEY"),
		client: &http.Client{},
	}, nil
}

// RegisterRoutes allows the driver to add additional routes at boot time
func (k *Kong) RegisterRoutes() {
}

// CreateUser creates a new consumer
func (k *Kong) CreateUser(user *User) (*User, error) {
	user.ID = uuid.NewV4().String()

	data := url.Values{}
	data.Add("custom_id", user.ID)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/consumers", k.admin), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	if _, err := k.client.Do(req); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser looks up a consumer
func (k *Kong) GetUser(userID string) (*User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/consumers/%s", k.admin, userID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result *KongUser
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &User{
		ID: result.ID,
	}, nil
}

// DeleteUser removes a consumer
func (k *Kong) DeleteUser(userID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/consumers/%s", k.admin, userID), nil)
	if err != nil {
		return err
	}

	if _, err := k.client.Do(req); err != nil {
		return err
	}

	return nil
}

// CreateClient creates a new application
func (k *Kong) CreateClient(client *Client) (*Client, error) {
	if len(client.ConsumerID) == 0 {
		return nil, fmt.Errorf("Kong requires consumer_id when creating a client")
	}

	data := url.Values{}
	data.Add("name", client.Name)
	data.Add("redirect_uri", client.RedirectURI)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/consumers/%s/oauth2", k.admin, client.ConsumerID), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result *Client
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	client.ClientID = result.ClientID
	client.ClientSecret = result.ClientSecret

	return client, nil
}

// GetClient looks up an application
func (k *Kong) GetClient(clientID string) (*Client, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/oauth2", k.admin), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("client_id", clientID)
	req.URL.RawQuery = q.Encode()

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result *KongClientList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Clients) != 1 {
		return nil, fmt.Errorf("Client not found for %s", clientID)
	}

	return result.Clients[0], nil
}

// DeleteClient deletes a client
func (k *Kong) DeleteClient(client *Client) error {
	if len(client.ConsumerID) == 0 {
		return fmt.Errorf("Kong requires consumer_id when creating a client")
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/consumers/%s/oauth2", k.admin, client.ConsumerID), nil)
	if err != nil {
		return err
	}

	_, err = k.client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

// Authorize attempts to authenticate a user
func (k *Kong) Authorize(client *Client, user *User, scope string) (*Redirect, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/oauth2/authorize", k.api), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("host", client.Host)

	q := req.URL.Query()
	q.Add("user_id", user.ID)
	req.URL.RawQuery = q.Encode()

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var redirect *Redirect
	if err := json.Unmarshal(body, &redirect); err != nil {
		return nil, err
	}

	return redirect, nil
}
