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
	"os"
	"time"

	"github.com/2-IMMERSE/auth-service/middleware"
	"github.com/2-IMMERSE/auth-service/model"
	"github.com/2-IMMERSE/auth-service/response"
	"github.com/2-IMMERSE/auth-service/tools"
	"github.com/2-IMMERSE/auth-service/tyk"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
)

type AuthServer struct {
	client *tyk.Client
}

func MountAuthServer(prefix string, e *echo.Echo, v *tools.Validator) *AuthServer {
	var s *AuthServer
	tykOrg := os.Getenv("TYK_ORG")
	tykKey := os.Getenv("TYK_KEY")
	tykURL := os.Getenv("TYK_URL")

	if len(tykURL) == 0 || len(tykOrg) == 0 || len(tykKey) == 0 {
		logrus.Info("Missing Tyk configuration, disabling integration!")

		s = &AuthServer{}
	} else {
		c := tyk.NewClient(tyk.Config{
			Organisation: tykOrg,
			BaseURL:      tykURL,
			Key:          tykKey,
		})

		s = &AuthServer{
			client: c,
		}
	}

	g := e.Group(prefix)

	g.POST("/tokens", s.createToken)
	g.POST("/revoke", s.revokeToken, middleware.Auth())

	return s
}

func (s *AuthServer) createToken(c echo.Context) error {
	a := model.Auth{}
	if err := c.Bind(&a); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("users")

	u := model.User{}
	if err := collection.Find(bson.M{"email": a.Username}).One(&u); err != nil {
		return response.Error{
			Message:    "Username or password incorrect",
			StatusCode: http.StatusBadRequest,
		}
	}

	if !u.ValidatePassword(a.Password) {
		return response.Error{
			Message:    "Username or password incorrect",
			StatusCode: http.StatusBadRequest,
		}
	}

	// create token
	t := model.NewToken(u)

	// send to tyk before storing in our db
	if s.client != nil {
		if err := s.client.CreateKey(t.Token); err != nil {
			return response.Error{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			}
		}
	}

	tokenCollection := db.C("tokens")
	if err := tokenCollection.Insert(t); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	cookie := new(http.Cookie)
	cookie.Name = "2immerse_token"
	cookie.Value = t.Token
	cookie.Path = "/"
	cookie.Expires = time.Now().Add(36 * time.Hour)

	c.SetCookie(cookie)

	return c.JSON(http.StatusCreated, t)
}

func (s *AuthServer) revokeToken(c echo.Context) error {
	a := model.Auth{}
	if err := c.Bind(&a); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("tokens")

	if err := collection.Remove(bson.M{"token": a.Token}); err != nil {
		// we have to check the string value as mgo returns many possible errors
		if err.Error() != "not found" {
			return response.Error{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			}
		}
	}

	if s.client != nil {
		s.client.DeleteKey(a.Token)
	}

	return c.NoContent(http.StatusNoContent)
}
