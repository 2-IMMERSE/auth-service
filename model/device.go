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
	"fmt"
	"math/rand"

	"gopkg.in/mgo.v2/bson"
)

const DeviceTypeCommunal = "communal"
const DeviceTypeCompanion = "companion"

type Device struct {
	ID    string        `bson:"_id" json:"id"`
	Type  string        `bson:"type" json:"type"`
	Code  string        `bson:"code" json:"code,omitempty"`
	Owner bson.ObjectId `bson:"owner,omitempty" json:"owner,omitempty"`
	Aux   string        `bson:"aux,omitempty" json:"aux,omitempty"`
}

func (d *Device) GenerateId() {
	d.ID = fmt.Sprintf("%s-%s", d.Type, generateToken(10, true))
}

// RefreshCode generates a new random 6 digit connection code
func (d *Device) RefreshCode() {
	num := 10 + rand.Intn(89)
	str := generateToken(5, false)

	d.Code = fmt.Sprintf("%s%d%s", str[:2], num, str[2:len(str)])
}
