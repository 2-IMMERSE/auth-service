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

	"gitlab-ext.irt.de/2-immerse/auth-service/middleware"
	"gitlab-ext.irt.de/2-immerse/auth-service/model"
	"gitlab-ext.irt.de/2-immerse/auth-service/response"
	"gitlab-ext.irt.de/2-immerse/auth-service/tools"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/labstack/echo"

	"github.com/Sirupsen/logrus"
)

type DeviceServer struct {
	echo *echo.Echo
}

func MountDeviceServer(prefix string, e *echo.Echo, v *tools.Validator) *DeviceServer {
	s := &DeviceServer{
		echo: e,
	}

	g := e.Group(prefix)

	g.GET("", s.index, middleware.Auth())
	g.POST("", s.register)
	g.GET("/:id", s.check)
	g.DELETE("/:id", s.delete, middleware.Auth())

	return s
}

func (s *DeviceServer) index(c echo.Context) error {
	user := c.Get("user").(model.User)
	pagination := tools.NewPagination("devices", c)

	if !user.HasRole("ROLE_ADMIN") {
		pagination.AddFilter("owner", user.ID)
	}

	devices := []model.Device{}
	if err := pagination.All(&devices); err != nil {
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

	return c.JSON(http.StatusOK, devices)
}

func (s *DeviceServer) register(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("devices")

	d := model.Device{}
	if err := c.Bind(&d); err != nil {
		return err
	}

	if len(d.Type) == 0 {
		d.Type = model.DeviceTypeCommunal
	}

	if d.Type != model.DeviceTypeCompanion && d.Type != model.DeviceTypeCommunal {
		return response.Error{
			Message:    "Invalid device type",
			StatusCode: http.StatusBadRequest,
		}
	}

	// generate a code for this device
	if d.Type == model.DeviceTypeCommunal {
		if len(d.Code) == 0 {
			d.RefreshCode()
		} else {
			// If there is already a record for d.Code, don't allocate a new deviceId,
			// return existing deviceId instead
			if err := collection.Find(bson.M{"code": d.Code}).One(&d); err != nil {
				logrus.Debugf("No record found for d.Code=%s, creating new record", d.Code)
			} else {
				return c.JSON(http.StatusCreated, d)
			}
		}
	}

	// insert into the database
	for {
		d.GenerateId()

		if err := collection.Insert(d); err != nil {
			if !mgo.IsDup(err) {
				return response.Error{
					Message:    err.Error(),
					StatusCode: http.StatusInternalServerError,
				}
			}
		} else {
			return c.JSON(http.StatusCreated, d)
		}
	}
}

func (s *DeviceServer) check(c echo.Context) error {
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("devices")

	d := model.Device{}
	if err := collection.FindId(c.Param("id")).One(&d); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	// is the owner set yet
	if !d.Owner.Valid() {
		// what is a sensible status code to return here?
		// it needs to be simple and clear so that requesting clients
		//   know quickly that thye should try again later
		c.Response().Header().Set("Retry-After", "2")
		c.Response().WriteHeader(http.StatusNoContent)
		return nil
	}

	// lookup the token attached to the owner
	tokenCollection := db.C("tokens")
	t := model.Token{}
	if err := tokenCollection.Find(bson.M{"user": d.Owner}).Sort("-expires").One(&t); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}
	t.Aux = d.Aux

	// send the token to the user. We don't check if it's expired here as their next requests will fail
	return c.JSON(http.StatusOK, t)
}

func (s *DeviceServer) delete(c echo.Context) error {
	user := c.Get("user").(model.User)
	db := c.Get("mgo_db").(*mgo.Database)
	collection := db.C("devices")

	d := model.Device{}
//	if err := collection.FindId(bson.ObjectIdHex(c.Param("id"))).One(&d); err != nil {
	if err := collection.FindId(c.Param("id")).One(&d); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	if !user.HasRole("ROLE_ADMIN") && d.Owner != user.ID {
		return echo.ErrForbidden
	}

	if err := collection.RemoveId(d.ID); err != nil {
		return response.Error{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return c.NoContent(http.StatusNoContent)
}
