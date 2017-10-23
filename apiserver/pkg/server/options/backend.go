/* Copyright (c) 2016-2017 - CloudPerceptions, LLC. All rights reserved.
  
   Licensed under the Apache License, Version 2.0 (the "License"); you may
   not use this file except in compliance with the License. You may obtain
   a copy of the License at
  
	http://www.apache.org/licenses/LICENSE-2.0
  
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
   WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
   License for the specific language governing permissions and limitations
   under the License.
*/

package options

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/rantuttl/cloudops/apiserver/pkg/genericserver/server"
	"github.com/rantuttl/cloudops/apiserver/pkg/backend"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/generic"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
)

type BackendOptions struct {
	BackendConfig	backend.Config
}

func NewBackendOptions(backendConfig *backend.Config) *BackendOptions {
	return &BackendOptions{
		BackendConfig:	*backendConfig,
	}
}

func (s *BackendOptions) Validate() []error {
	allErrors := []error{}
	if len(s.BackendConfig.ServerList) == 0 {
		allErrors = append(allErrors, fmt.Errorf("--backend-servers must be specified"))
	}

	return allErrors
}

func (s *BackendOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&s.BackendConfig.ServerList, "backend-servers", s.BackendConfig.ServerList,
		"List of backend servers to connect with (scheme://ip:port), comma separated.")
	fs.StringVar(&s.BackendConfig.KeyFile, "backend-keyfile", s.BackendConfig.KeyFile,
		"SSL key file used to secure backend communication.")

	fs.StringVar(&s.BackendConfig.CertFile, "backend-certfile", s.BackendConfig.CertFile,
		"SSL certification file used to secure backend communication.")

	fs.StringVar(&s.BackendConfig.CAFile, "backend-cafile", s.BackendConfig.CAFile,
		"SSL Certificate Authority file used to secure backend communication.")
}

func (s *BackendOptions) ApplyTo(c *server.Config) error {
	// TODO (rantuttl): Set c.RESTOptionsGetter
	c.RESTOptionsGetter = &SimpleRestOptionsFactory{Options: *s}
	return nil
}

type SimpleRestOptionsFactory struct {
	Options BackendOptions
}

// Usually called when a new REST API resources is created
func (f *SimpleRestOptionsFactory) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
	ret := generic.RESTOptions{
		BackendConfig:	&f.Options.BackendConfig,
		Decorator:	generic.UndecoratedBackend,
		ResourcePrefix:	resource.Group + "/" + resource.Resource,
	}
	return ret, nil
}
