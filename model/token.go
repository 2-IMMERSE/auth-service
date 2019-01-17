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
	"math/rand"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const letterBytesLower = "abcdefghijklmnopqrstuvwxyz"
const letterBytesUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

type Token struct {
	ID        bson.ObjectId `bson:"_id,omitempty" json:"-"`
	ExpiresAt time.Time     `bson:"expires" json:"expires_at"`
	Token     string        `bson:"token" json:"access_token"`
	User      bson.ObjectId `bson:"user" json:"-"`
	Client    bson.ObjectId `bson:"client,omitempty" json:"-"`
	Aux       string        `bson:"aux,omitempty" json:"aux,omitempty"`
}

func NewToken(user User) Token {
	return Token{
		ID:        bson.NewObjectId(),
		ExpiresAt: time.Now().Add(time.Duration(604800 * time.Second)),
		Token:     generateToken(64, true),
		User:      user.ID,
	}
}

func generateToken(n int, includeUpper bool) string {
	letterBytes := letterBytesLower

	if includeUpper {
		letterBytes = letterBytes + letterBytesUpper
	}

	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
