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

package generic

import (
	"github.com/golang/glog"

	"github.com/rantuttl/cloudops/apiserver/pkg/backend"
	"github.com/rantuttl/cloudops/apiserver/pkg/backend/factory"
)

type BackendDecorator func(config *backend.Config, transformer backend.BackendTransformer) (backend.Interface)

func UndecoratedBackend(config *backend.Config, transformer backend.BackendTransformer) (backend.Interface) {
	return NewBackend(config, transformer)
}

func NewBackend(config *backend.Config, transformer backend.BackendTransformer) (backend.Interface) {
	s, err := factory.Create(*config, transformer)
	if err != nil {
		glog.Fatalf("Unable to create backend: config (%v), err (%v)", config, err)
	}
	return s
}
