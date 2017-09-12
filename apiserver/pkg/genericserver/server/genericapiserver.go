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
package server

import (
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/serializer"
	"github.com/rantuttl/cloudops/apimachinery/pkg/apimachinery"
	"github.com/rantuttl/cloudops/apimachinery/pkg/apimachinery/registered"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/rest"
)

// FIXME (rantuttl): Stub for now
// Info about an API group.
type APIGroupInfo struct {
	GroupMeta apimachinery.GroupMeta
	// Info about the resources in this group. Its a map from version to resource to the storage.
	VersionedResourcesStorageMap map[string]map[string]rest.Storage
}

type GenericAPIServer struct {
	SecureServingInfo *SecureServingInfo
	// numerical ports, set after listening
	effectiveSecurePort int
	Serializer runtime.NegotiatedSerializer
	Handler *APIServerHandler
}

// EffectiveSecurePort returns the secure port we bound to.
func (s *GenericAPIServer) EffectiveSecurePort() int {
	return s.effectiveSecurePort
}

type preparedGenericAPIServer struct {
	*GenericAPIServer
}

// PrepareRun does post API installation setup steps.
func (s *GenericAPIServer) PrepareRun() preparedGenericAPIServer {
	// initialize some things on the server

	return preparedGenericAPIServer{s}
}

// Run spawns the secure http server. It only returns if stopCh is closed
// or the secure port cannot be listened on initially.
func (s preparedGenericAPIServer) Run(stopCh <-chan struct{}) error {
	err := s.NonBlockingRun(stopCh)
	if err != nil {
		return err
	}

	<-stopCh
	return nil
}

// NonBlockingRun spawns the secure http server. An error is
// returned if the secure port cannot be listened on.
func (s preparedGenericAPIServer) NonBlockingRun(stopCh <-chan struct{}) error {
	return nil
}

// Exposes the given api group in the API.
func (s *GenericAPIServer) InstallAPIGroup(apiGroupInfo *APIGroupInfo) error {

	return nil
}

// NewDefaultAPIGroupInfo returns an APIGroupInfo stubbed with "normal" values
// exposed for easier composition from other packages
func NewDefaultAPIGroupInfo(group string, registry *registered.APIRegistrationManager, scheme *runtime.Scheme, parameterCodec runtime.ParameterCodec, codecs serializer.CodecFactory) APIGroupInfo {
	groupMeta := registry.GroupOrDie(group)

	return APIGroupInfo{
		GroupMeta:			*groupMeta,
		VersionedResourcesStorageMap:	map[string]map[string]rest.Storage{},
	}
}
