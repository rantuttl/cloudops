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

package backend

import (
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
)

type Config struct {
	// ServerList is the list of backend servers to connect with.
	ServerList []string
	// TLS credentials
	KeyFile  string
	CertFile string
	CAFile   string
	Codec  runtime.Codec
	Copier runtime.ObjectCopier
}

func NewDefaultConfig(copier runtime.ObjectCopier, codec runtime.Codec) *Config {
	return &Config{
		Codec:	codec,
		Copier:	copier,
	}
}
