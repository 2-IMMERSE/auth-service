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

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
)

type HealthcheckServer struct {
	echo *echo.Echo
}

func MountHealthcheckServer(prefix string, e *echo.Echo, debug bool) *HealthcheckServer {
	s := &HealthcheckServer{
		echo: e,
	}

	g := e.Group(prefix)

	g.GET("", s.status)

	if debug {
		g.GET("/routes", s.routes)
	}

	return s
}

func (s *HealthcheckServer) status(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)

	if err := db.Session.Ping(); err != nil {
		logrus.Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *HealthcheckServer) routes(c echo.Context) error {
	return c.JSON(http.StatusOK, s.echo.Routes())
}
