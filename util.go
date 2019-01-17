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
	"runtime"

	"github.com/Sirupsen/logrus"
)

func setMaxProcs() {
	var numProcs int

	if *maxProcs < 1 {
		numProcs = runtime.NumCPU()
		logrus.Debugf("Setting max procs to %d", numProcs)
	} else {
		numProcs = *maxProcs
	}

	runtime.GOMAXPROCS(numProcs)

	// Check if the setting was successful.
	actualNumProcs := runtime.GOMAXPROCS(0)
	if actualNumProcs != numProcs {
		logrus.Warningf("Specified max procs of %v but using %v", numProcs, actualNumProcs)
	}
}
