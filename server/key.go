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

package server

import (
	"fmt"
	"io/ioutil"
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"gitlab-ext.irt.de/2-immerse/auth-service/middleware"
	"gitlab-ext.irt.de/2-immerse/auth-service/model"
	"gitlab-ext.irt.de/2-immerse/auth-service/response"
	"gitlab-ext.irt.de/2-immerse/auth-service/tools"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
)

type KeyServer struct {
	echo *echo.Echo
}

func MountKeyServer(prefix string, e *echo.Echo, validator *tools.Validator) *KeyServer {
	s := &KeyServer{
		echo: e,
	}
	g := e.Group(prefix)

	g.GET("", s.index, middleware.Auth()).Name = "keys_list"
	g.POST("", s.create, middleware.Validator(validator.GetValidator("key"))).Name = "keys_create"
	g.GET("/:id", s.show, middleware.Auth()).Name = "keys_show"

	return s
}

func (s *KeyServer) index(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("keys")

	query := collection.Find(bson.M{})

	count, err := query.Count()
	if err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	keys := []model.Key{}
	if err := query.All(&keys); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	c.Response().Header().Set("Accept-Ranges", "keys")
	c.Response().Header().Set("Content-Range", fmt.Sprintf("keys 0-10/%d", count))

	return c.JSON(http.StatusOK, keys)
}

func (s *KeyServer) create(c echo.Context) error {
	d, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		logrus.Error(err)
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	k := model.Key{
		Data: d,
	}

	k.ID = bson.NewObjectId()
	k.GenerateSlug()

	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("keys")

	if err := collection.Insert(k); err != nil {
		logrus.Error(err)
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return c.JSON(http.StatusCreated, response.Resource{
		ID: k.ID.Hex(),
	})
}

func (s *KeyServer) show(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("keys")

	k := model.Key{}
	if err := collection.FindId(bson.ObjectIdHex(c.Param("id"))).One(&k); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	return c.Blob(http.StatusOK, "application/octet-stream", k.Data)
}
