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

package middleware

import (
	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	"github.com/2-IMMERSE/auth-service/model"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// if there is an access token provided we need to extract it and lookup the user based on the token
// should we lookup the user now or make it explicit?
func Token(db *mgo.Database) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			
			// Check submitted tokens for all routes except when creating new users
			if strings.Contains(c.Path(), "/tokens") && c.Request().Method == "POST" {
				return next(c);
			}

			token := c.QueryParam("access_token")
			if len(token) == 0 {
				token = c.Request().Header.Get("Authorization")
			}
			if len(token) == 0 {
				// fall back to cookie
				if cookie, err := c.Request().Cookie("2immerse_token"); err == nil {
					token = cookie.Value
				}
			}

			if len(token) > 0 {
				// we have a token so now we lookup the user
				s := db.Session.Clone()
				defer s.Close()

				t := model.Token{}
				if err := s.DB(db.Name).C("tokens").Find(bson.M{"token": token}).One(&t); err != nil {
					logrus.Errorf("Error fetching token: %v", err)
					return echo.ErrForbidden
				}

				// we have the token so now get the user
				u := model.User{}
				if err := s.DB(db.Name).C("users").Find(bson.M{"_id": t.User}).One(&u); err != nil {
					logrus.Errorf("Error fetching user: %v", err)
					return echo.ErrForbidden
				}

				c.Set("user", u)
			}

			return next(c)
		}
	}
}
