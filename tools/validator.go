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
// 	"io/ioutil"
// 	"path"
// 	"path/filepath"
// 	"strings"

	"github.com/Sirupsen/logrus"
	jsschema "github.com/lestrrat-go/jsschema"
	jsval "github.com/lestrrat-go/jsval"
	"github.com/lestrrat-go/jsval/builder"
)

type Validator struct {
	schemas map[string]*jsschema.Schema
}

func NewValidator(schemasDir string) *Validator {
	v := &Validator{
		schemas: make(map[string]*jsschema.Schema),
	}

	//logrus.Debugf("Loading validation schema...")
	//files, _ := ioutil.ReadDir(schemasDir)
	//for _, f := range files {
	//	if !f.IsDir() {
	//		logrus.Debugf("Loading %s...", f.Name())
	//		if s, err := jsschema.ReadFile(path.Join(schemasDir, f.Name())); err != nil {
	//			logrus.Error(err)
	//		} else {
	//			key := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
	//			v.schemas[key] = s
	//		}
	//	}
	//}
	//logrus.Debugf("%+v", v.schemas)


	return v
}

func (v *Validator) GetValidator(name string) *jsval.JSVal {
	s, ok := v.schemas[name]
	if !ok {
		logrus.Errorf("schema for %s not found", name)
		return nil
	}

	b := builder.New()
	val, err := b.Build(s)
	if err != nil {
		logrus.Error(err)
		return nil
	}

	return val
}
