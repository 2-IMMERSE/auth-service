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
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/2-IMMERSE/auth-service/middleware"
	"github.com/2-IMMERSE/auth-service/model"
	"github.com/2-IMMERSE/auth-service/response"
	"github.com/2-IMMERSE/auth-service/tools"

	"github.com/labstack/echo"
)

type MeServer struct {
}

func MountMeServer(prefix string, e *echo.Echo, v *tools.Validator) *MeServer {
	s := &MeServer{}

	g := e.Group(prefix, middleware.Auth())

	g.GET("", s.showUser)
	g.PATCH("", s.updateUser)
	g.GET("/profile", s.showProfile)
	g.PATCH("/profile", s.updateProfile)
	g.GET("/roles", s.showRoles)
	g.GET("/groups", s.showGroups)
	g.POST("/link", s.linkDevice)

	return s
}

func (s *MeServer) showUser(c echo.Context) error {
	user := c.Get("user").(model.User)

	return c.JSON(http.StatusOK, user)
}

func (s *MeServer) updateUser(c echo.Context) error {
	user := c.Get("user").(model.User)
	db := c.Get("mgo_db").(*mgo.Database)

	u := model.User{}
	if err := c.Bind(&u); err != nil {
		return err
	}

	clearTokens := false
	if len(u.PlainPassword) > 0 {
		u.HashPassword()
		clearTokens = true
	}

	if err := db.C("users").UpdateId(user.ID, bson.M{"$set": u}); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	if clearTokens {
		// delete all this users tokens
		db.C("tokens").RemoveAll(bson.M{"user": user.ID})
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *MeServer) showProfile(c echo.Context) error {
	user := c.Get("user").(model.User)

	if user.Profile == nil {
		user.Profile = &model.Profile{}
	}

	return c.JSON(http.StatusOK, user.Profile)
}

func (s *MeServer) updateProfile(c echo.Context) error {
	user := c.Get("user").(model.User)
	db := c.Get("mgo_db").(*mgo.Database)

	if err := c.Bind(&user.Profile); err != nil {
		return err
	}

	if err := db.C("users").UpdateId(user.ID, bson.M{"$set": bson.M{"profile": user.Profile}}); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *MeServer) showRoles(c echo.Context) error {
	user := c.Get("user").(model.User)

	if user.Roles == nil {
		user.Roles = []string{}
	}

	return c.JSON(http.StatusOK, user.Roles)
}

func (s *MeServer) showGroups(c echo.Context) error {
	user := c.Get("user").(model.User)

	if user.Groups == nil {
		user.Groups = []string{}
	}

	return c.JSON(http.StatusOK, user.Groups)
}

func (s *MeServer) linkDevice(c echo.Context) error {
	// lookup device
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("devices")

	d := model.Device{}
	if err := c.Bind(&d); err != nil {
		return err
	}

	aux := d.Aux

	// d := model.Device{}
	if err := collection.Find(bson.M{"code": d.Code}).One(&d); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	d.Owner = c.Get("user").(model.User).ID
	d.Aux = aux

	if err := collection.UpdateId(d.ID, d); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return c.NoContent(http.StatusNoContent)
}
