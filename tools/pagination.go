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

package tools

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const maxLimit = 50

type Pagination struct {
	collection *mgo.Collection
	query      *mgo.Query
	count      int

	filter bson.M
	sort   string
	offset int
	limit  int
}

func NewPagination(collection string, c echo.Context) *Pagination {
	db := c.Get("mgo_db").(*mgo.Database)

	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit < 1 {
		limit = 10
	} else if limit > maxLimit {
		limit = maxLimit
	}

	s := c.QueryParam("sort")
	if len(s) == 0 || s == "id" {
		s = "_id"
	} else if s == "-id" {
		s = "-_id"
	}

	p := &Pagination{
		collection: db.C(collection),
		filter:     bson.M{},
		offset:     offset,
		limit:      limit,
		sort:       s,
	}

	return p
}

func (p *Pagination) createQuery() *mgo.Query {
	return p.collection.Find(p.filter).Sort(p.sort)
}

func (p *Pagination) AddFilter(key string, filter interface{}) {
	p.filter[key] = filter
}

func (p *Pagination) Count() (int, error) {
	return p.createQuery().Count()
}

func (p *Pagination) All(data interface{}) error {
	return p.createQuery().
		Skip(p.offset).
		Limit(p.limit).
		All(data)
}

func (p *Pagination) AddHeaders(headers http.Header) error {
	count, err := p.Count()
	if err != nil {
		return err
	}

	headers.Set("Accept-Ranges", p.collection.Name)
	headers.Set("Content-Range", fmt.Sprintf("%s 0-10/%d", p.collection.Name, count))

	return nil
}
