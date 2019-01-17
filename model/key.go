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
	"github.com/gosimple/slug"
	"gopkg.in/mgo.v2/bson"
)

type Key struct {
	ID    bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Title string        `bson:"title" json:"title,omitempty"`
	Slug  string        `bson:"slug" json:"slug,omitempty"`
	Data  []byte        `bson:"data,omitempty" json:"data"`
}

func (k *Key) GenerateSlug() {
	k.Slug = slug.Make(k.Title)
}
