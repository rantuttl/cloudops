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

package v1

import (
	"strings"
)

// GroupVersionResource unambiguously identifies a resource.  It doesn't anonymously include GroupVersion
// to avoid automatic coersion.  It doesn't use a GroupVersion to avoid custom marshalling
//
type GroupVersionResource struct {
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`
}

func (gvr *GroupVersionResource) String() string {
	return strings.Join([]string{gvr.Group, "/", gvr.Version, ", Resource=", gvr.Resource}, "")
}

// GroupVersionKind unambiguously identifies a kind.  It doesn't anonymously include GroupVersion
// to avoid automatic coersion.  It doesn't use a GroupVersion to avoid custom marshalling
//
// +protobuf.options.(gogoproto.goproto_stringer)=false
type GroupVersionKind struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

func (gvk GroupVersionKind) String() string {
	return gvk.Group + "/" + gvk.Version + ", Kind=" + gvk.Kind
}
