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

func NewCalBackend(codec runtime.Codec, copier runtime.ObjectCopier, transformer backend.BackendTransformer) backend.Interface {
	h := &calHelper{
		codec:		codec,
		copier:		copier,
		transformer:	transformer,
	}
	if transformer == nil {
		h.transformer = DefaultTransformer
	}
	return h
}

type defaultTransformer struct{}

func (defaultTransformer) BackendTransformerInitializer(c backend.Config) error {
	return nil
}

func (defaultTransformer) TransformToBackend(ctx context.Context, data string) (string, error) {
	return data, nil
}

func (defaultTransformer) TransformFromBackend(ctx context.Context, data string) (string, error) {
	return data, nil
}

var DefaultTransformer backend.BackendTransformer = defaultTransformer{}

type calHelper struct {
	// TODO (rantuttl): Put things needed for CAL communication and other helper functions that
	// would be helpful, especially things about the CAL client and things unique to the API
	// group that can be used for CAL communications. Codec libraries for encoding / decoding
	// requests / responses to CAL; things for managing cache (if used)
	//client whatevertype
	codec		runtime.Codec
	copier		runtime.ObjectCopier
	transformer	backend.BackendTransformer
}

func (h *calHelper) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	if ctx == nil {
		glog.Errorf("Context is nil")
	}
	// 1. Convert and Encode object with calHelper known codecs
	data, err := runtime.Encode(h.codec, obj)
	if err != nil {
		return err
	}
	// 2. Transform object (if needed)
	newBody, err := h.transformer.TransformToBackend(ctx, string(data))
	glog.Infof("Transformed & string-a-fied obj:\n%s", newBody)
	// 3. Set any TTL options for CAL request
	// 4. TODO metrics for latency
	// 5. Send request to client
	// 6. If out != nil, copy CAL response body back to out
	//	6a. Transform object (if needed)
	//	6b. Decode object with calHelper known codecs

	return err
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
