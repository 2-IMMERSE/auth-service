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

	"gitlab-ext.irt.de/2-immerse/auth-service/middleware"
	"gitlab-ext.irt.de/2-immerse/auth-service/model"
	"gitlab-ext.irt.de/2-immerse/auth-service/response"
	"gitlab-ext.irt.de/2-immerse/auth-service/tools"

	"github.com/labstack/echo"
)

type UserServer struct {
}

func MountUserServer(prefix string, e *echo.Echo, v *tools.Validator) *UserServer {
	s := &UserServer{}
	g := e.Group(prefix)

	// account
	g.GET("", s.index, middleware.Auth(), middleware.Admin())
	g.POST("", s.create, middleware.Validator(v.GetValidator("user")))
	g.GET("/:id", s.show, middleware.Auth(), middleware.Admin())
	g.PATCH("/:id", s.update, middleware.Auth(), middleware.Admin())
	g.DELETE("/:id", s.delete, middleware.Auth(), middleware.Admin())

	// profile
	g.GET("/:id/profile", s.showProfile, middleware.Auth(), middleware.Admin())
	g.PATCH("/:id/profile", s.updateProfile, middleware.Auth(), middleware.Admin())

	// roles
	g.GET("/:id/roles", s.showRoles, middleware.Auth(), middleware.Admin())

	// groups
	g.GET("/:id/groups", s.showGroups, middleware.Auth(), middleware.Admin())

	return s
}

func (s *UserServer) index(c echo.Context) error {
	pagination := tools.NewPagination("users", c)

	users := []model.User{}
	if err := pagination.All(&users); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	if err := pagination.AddHeaders(c.Response().Header()); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return c.JSON(http.StatusOK, users)
}

func (s *UserServer) create(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("users")

	u := model.User{}
	if err := c.Bind(&u); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	if count, _ := collection.Find(bson.M{"email": u.Email}).Count(); count > 0 {
		return response.Error{
			Message:    "Email address already in use",
			StatusCode: http.StatusBadRequest,
		}
	}

	u.ID = bson.NewObjectId()
	u.Profile = &model.Profile{
		Communal:  make(map[string]interface{}),
		Companion: make(map[string]interface{}),
	}

	// update the users password
	if err := u.HashPassword(); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	if err := collection.Insert(u); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return c.JSON(http.StatusCreated, response.Resource{
		ID: u.ID.Hex(),
	})
}

func (s *UserServer) show(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("users")

	u := model.User{}
	if err := collection.FindId(bson.ObjectIdHex(c.Param("id"))).One(&u); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	return c.JSON(http.StatusOK, u)
}

func (s *UserServer) update(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("users")

	u := model.User{}
	if err := collection.FindId(bson.ObjectIdHex(c.Param("id"))).One(&u); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	update := model.User{}
	if err := c.Bind(&update); err != nil {
		return err
	}

	clearTokens := false
	if len(update.PlainPassword) > 0 {
		update.HashPassword()
		clearTokens = true
	}

	if err := collection.UpdateId(u.ID, bson.M{"$set": update}); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	if clearTokens {
		// delete all this users tokens
		db.C("tokens").RemoveAll(bson.M{"user": u.ID})
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *UserServer) delete(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("users")

	u := model.User{}
	if err := collection.FindId(bson.ObjectIdHex(c.Param("id"))).One(&u); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	if err := collection.RemoveId(u.ID); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// profile
func (s *UserServer) showProfile(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("users")

	u := model.User{}
	if err := collection.FindId(bson.ObjectIdHex(c.Param("id"))).One(&u); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	return c.JSON(http.StatusOK, u.Profile)
}

func (s *UserServer) updateProfile(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("users")

	u := model.User{}
	if err := collection.FindId(bson.ObjectIdHex(c.Param("id"))).One(&u); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	if err := c.Bind(&u.Profile); err != nil {
		return err
	}

	if err := collection.UpdateId(u.ID, bson.M{"$set": bson.M{"profile": u.Profile}}); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// roles
func (s *UserServer) showRoles(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("users")

	u := model.User{}
	if err := collection.FindId(bson.ObjectIdHex(c.Param("id"))).One(&u); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	return c.JSON(http.StatusOK, u.Roles)
}

// groups
func (s *UserServer) showGroups(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("users")

	u := model.User{}
	if err := collection.FindId(bson.ObjectIdHex(c.Param("id"))).One(&u); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	return c.JSON(http.StatusOK, u.Groups)
}
