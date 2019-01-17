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

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type TykUser struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
	Meta    string `json:"Meta"`
}

type Tyk struct {
	api    string
	secret string
	client *http.Client
}

func NewTyk() (*Tyk, error) {
	api := os.Getenv("TYK_API")
	if len(api) == 0 {
		return nil, fmt.Errorf("TYK_API not set")
	}

	return &Tyk{
		api:    api,
		secret: os.Getenv("TYK_SHARED_SECRET"),
		client: &http.Client{},
	}, nil
}

// RegisterRoutes allows the driver to add additional routes at boot time
func (t *Tyk) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/", t.notify).Methods("POST")
}

// CreateUser creates a new consumer
func (t *Tyk) CreateUser(user *User) (*User, error) {
	data := url.Values{}
	data.Add("email_address", user.Username)
	data.Add("active", "1")

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/users", t.api), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result *TykUser
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	logrus.Infof("%+v", resp.Body)
	logrus.Infof("%+v", result)
	// set id of user fomr here

	return user, nil
}

// GetUser looks up a consumer
func (t *Tyk) GetUser(userID string) (*User, error) {
	return nil, nil
}

// DeleteUser removes a consumer
func (t *Tyk) DeleteUser(userID string) error {
	return nil
}

// CreateClient creates a new application
func (t *Tyk) CreateClient(client *Client) (*Client, error) {
	return nil, nil
}

// GetClient looks up an application
func (t *Tyk) GetClient(clientID string) (*Client, error) {
	return nil, nil
}

// DeleteClient deletes a client
func (t *Tyk) DeleteClient(client *Client) error {
	return nil
}

// Authorize attempts to authenticate a user
func (t *Tyk) Authorize(client *Client, user *User, scope string) (*Redirect, error) {
	return nil, nil
}

func (t *Tyk) notify(w http.ResponseWriter, r *http.Request) {
	// check shared secret matches otherwise discard.
	// store updated user data in manager
}
