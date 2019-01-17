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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	mgo "gopkg.in/mgo.v2"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	em "github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	"gitlab-ext.irt.de/2-immerse/auth-service/consul"
	"gitlab-ext.irt.de/2-immerse/auth-service/middleware"
	"gitlab-ext.irt.de/2-immerse/auth-service/model"
	"gitlab-ext.irt.de/2-immerse/auth-service/server"
	"gitlab-ext.irt.de/2-immerse/auth-service/tools"
)

const (
	version = "1.0.0"
	dbName  = "auth"
)

var (
	debugFlag        = flag.Bool("debug", false, "enable debug")
	versionFlag      = flag.Bool("version", false, "show version and exit")
	maxProcs         = flag.Int("procs", 0, "max number of CPUs that can be used simultaneously. Less than 1 for default (number of cores).")
	expireTokensFlag = flag.Bool("expire-tokens", false, "expire access tokens")

	listenAddr = flag.String("listen", ":8080", "[hostname:port] to listen on")
)

func init() {
	flag.Parse()

	if *debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if *versionFlag {
		println("Auth Service V%s", version)
		os.Exit(1)
	}

	setMaxProcs()
}

func getMongoAddress() string {
	mongoService := os.Getenv("MONGODB_SERVICE")
	if len(mongoService) == 0 {
		mongoService = "localhost:27017"
	}

	if strings.Contains(mongoService, ":") {
		return mongoService
	}

	client, err := consul.NewClient()
	if err != nil {
		logrus.Fatalf("Consul: %s", err)
	}

	port, err := client.LookupServicePort(mongoService)
	if err != nil {
		logrus.Fatal(err)
	}

	return fmt.Sprintf("%s.service.consul:%d", mongoService, port)
}

func main() {
	allowOrigins := os.Getenv("CORS_ALLOW_ORIGINS")
	if len(allowOrigins) == 0 {
		allowOrigins = "*"
	}

	logrus.Debugf("Allowing origin: %s", allowOrigins)

	logrus.Debugf("Dialing mongo...")
	s, err := mgo.DialWithTimeout(getMongoAddress(), 1*time.Second)
	if err != nil {
		logrus.Fatal(err)
	}
	db := &mgo.Database{
		Session: s,
		Name:    dbName,
	}
	defer db.Session.Close()

	db.Session.SetMode(mgo.Monotonic, true)

	e := echo.New()
	e.HideBanner = true
	if *debugFlag {
		e.Logger.SetLevel(log.DEBUG)
	}

	e.Pre(em.RemoveTrailingSlash())
	e.Use(em.Recover())
	e.Use(em.LoggerWithConfig(em.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			if c.Path() == "/healthcheck" {
				return true
			}

			return false
		},
	}))
	e.Use(em.CORSWithConfig(em.CORSConfig{
		AllowOrigins:     []string{allowOrigins, "*"},
		ExposeHeaders:    []string{"Content-Range"},
		AllowCredentials: true,
		AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE, echo.OPTIONS},
		MaxAge:           3600,
	}))
	e.Use(middleware.MGO(db))
	e.Use(middleware.Token(db))
	e.Use(middleware.Error())

	if *debugFlag {
		e.Use(em.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
			if c.Path() == "/healthcheck" {
				return
			}

			if len(reqBody) > 0 {
				logrus.Debugf("Request: %s Body: %s", c.Path(), reqBody)
			}

			if len(resBody) > 0 {
				logrus.Debugf("Response: %s Body: %s", c.Path(), resBody)
			}
		}))
	}

	logrus.Debug("Mounting services...")
	server.MountHealthcheckServer("/healthcheck", e, *debugFlag)

	v := tools.NewValidator("./schema")

	server.MountAuthServer("/auth", e, v)
	server.MountUserServer("/users", e, v)
	server.MountMeServer("/me", e, v)
	server.MountKeyServer("/keys", e, v)
	server.MountDeviceServer("/devices", e, v)

	if *expireTokensFlag {
		index := mgo.Index{
			Key:         []string{"token"},
			Unique:      true,
			DropDups:    true,
			Background:  true,
			Sparse:      true,
			ExpireAfter: time.Hour * time.Duration(24),
		}
		db.C("token").EnsureIndex(index)
	}

	unique := mgo.Index{
		Key:      []string{"$text:id"},
		Unique:   true,
		DropDups: true,
	}
	db.C("devices").EnsureIndex(unique)

	logrus.Info("Loading fixtures...")
	loadUserFixtures(db)
	loadKeyFixtures(db)

	// start server listening
	go func() {
		logrus.Infof("listening on %s!", *listenAddr)
		if err := e.Start(*listenAddr); err != nil {
			logrus.Error(err)
			logrus.Info("Shutting down server")
		}
	}()

	// handle graceful shutdowns
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logrus.Fatal(err)
	}
}

func loadUserFixtures(db *mgo.Database) error {
	var users []*model.User
	data, err := ioutil.ReadFile("./fixtures/users.json")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	collection := db.C("users")
	for _, u := range users {
		u.HashPassword()
		collection.Upsert(bson.M{"email": u.Email}, u)
	}

	return nil
}

func loadKeyFixtures(db *mgo.Database) error {
	collection := db.C("keys")
	files, _ := ioutil.ReadDir("./fixtures/keys")

	for _, f := range files {
		if !f.IsDir() {
			data, _ := ioutil.ReadFile(path.Join("./fixtures/keys", f.Name()))
			id := path.Base(f.Name())
			collection.Insert(&model.Key{
				ID:   bson.ObjectIdHex(id),
				Data: data,
			})
		}
	}

	return nil
}
