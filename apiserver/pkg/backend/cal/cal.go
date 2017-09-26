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

package cal

import (
	"golang.org/x/net/context"
	"github.com/golang/glog"

	"github.com/rantuttl/cloudops/apiserver/pkg/backend"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
)

func NewCalBackend() backend.Interface {
	return &calHelper{}
}

type calHelper struct {
	// TODO (rantuttl): Put things needed for CAL communication and other helper functions that
	// would be helpful, especially things about the CAL client and things unique to the API
	// group that can be used for CAL communications. Codec libraries for encoding / decoding
	// requests / responses to CAL; things for managing cache (if used)
}

func (h *calHelper) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	if ctx == nil {
		glog.Errorf("Context is nil")
	}
	glog.Infof("Create key: %s", key)

	return nil
}

func (h *calHelper) Get(ctx context.Context, key string, resourceVersion string, objPtr runtime.Object, ignoreNotFound bool) error {
	if ctx == nil {
		glog.Errorf("Context is nil")
	}
	glog.Infof("Get key: %s", key)

	return nil
}

func (h *calHelper) Delete(ctx context.Context, key string, out runtime.Object, preconditions *metav1.Preconditions) error {
	if ctx == nil {
		glog.Errorf("Context is nil")
	}
	glog.Infof("Delete key: %s", key)
	// NOTE: preconditions.UID is the UID of the object

	return nil
}
