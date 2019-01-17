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

package consul

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
)

// Client is a convenience wrapper for the consul api
type Client struct {
	consul *api.Client
}

// NewClient creates a new consul api wrapper
func NewClient() (*Client, error) {
	consul, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}

	try := 0
	for {
		try++
		status, err := consul.Status().Leader()
		if err != nil {
			logrus.Infof("Consul: waiting... %s", err)
			if try > 10 {
				return nil, err
			}
		} else {
			logrus.Debugf("Consul: %s", status)
			break
		}
	}

	logrus.Debug("Consul: connected")

	return &Client{
		consul: consul,
	}, nil
}

// LookupServicePort looks up the first port on the first node that a service is running on
func (c *Client) LookupServicePort(name string) (int, error) {
	logrus.Debugf("Consul: Looking up service %s...", name)
	catalog := c.consul.Catalog()

	try := 0
	for {
		try++
		services, _, err := catalog.Service(name, "", nil)
		if err != nil {
			if try > 10 {
				return 0, err
			}
		} else {
			if len(services) < 1 {
				if try > 10 {
					return 0, fmt.Errorf("Consul: No services found called %s", name)
				}
			} else {
				return services[0].ServicePort, nil
			}
		}
	}
}
