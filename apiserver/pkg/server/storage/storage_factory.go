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

package storage

import (
	"crypto/tls"
)

// Backend describes the storage servers, the information here should be enough
// for health validations.
type Backend struct {
	Server string
	TLSConfig *tls.Config
}

// StorageFactory is the interface to locate the storage for a given GroupResource
type StorageFactory interface {
	Backends() []Backend
}

// DefaultStorageFactory takes a GroupResource and returns back its storage interface.  This result includes:
// 1. Merged etcd config, including: auth, server locations, prefixes
// 2. Resource encodings for storage: group,version,kind to store as
// 3. Cohabitating default: some resources like hpa are exposed through multiple APIs.  They must agree on 1 and 2
type DefaultStorageFactory struct {
	// APIResourceConfigSource indicates whether the *storage* is enabled, NOT the API
	APIResourceConfigSource APIResourceConfigSource
}

func (s *DefaultStorageFactory) Backends() []Backend {

	backends := []Backend{}
	return backends
}
