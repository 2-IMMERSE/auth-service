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

package model

import (
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

// The User type encapsulates details about a user and their profile
type User struct {
	ID            bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	DisplayName   string                 `bson:"display_name,omitempty" json:"display_name,omitempty"`
	FirstName     string                 `bson:"first_name,omitempty" json:"first_name,omitempty"`
	LastName      string                 `bson:"last_name,omitempty" json:"last_name,omitempty"`
	Email         string                 `bson:"email,omitempty" json:"email"`
	PlainPassword string                 `bson:"-" json:"password,omitempty"`
	Password      []byte                 `bson:"password,omitempty" json:"-"`
	Roles         []string               `bson:"roles,omitempty" json:"roles,omitempty"`
	Groups        []string               `bson:"groups,omitempty" json:"groups,omitempty"`
	Settings      map[string]interface{} `bson:"settings" json:"settings"`
	Profile       *Profile               `bson:"profile,omitempty" json:"profile,omitempty"`
}

// The Profile type provides a map for companion and communal devices allowing
// arbitry storage of key value pairs.
type Profile struct {
	Companion map[string]interface{} `bson:"companion" json:"companion"`
	Communal  map[string]interface{} `bson:"communal"  json:"communal"`
}

func (u *User) HashPassword() error {
	if len(u.PlainPassword) == 0 {
		return nil
	}

	p, err := bcrypt.GenerateFromPassword([]byte(u.PlainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = p
	// perhaps we should write over the memory here?
	u.PlainPassword = ""

	return nil
}

func (u *User) ValidatePassword(pass string) bool {
	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(pass)); err != nil {
		return false
	}

	return true
}

func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}

	return false
}
