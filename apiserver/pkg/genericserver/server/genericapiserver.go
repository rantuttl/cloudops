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
	"fmt"
	"time"
	"strings"
	"net/http"

	"github.com/golang/glog"

	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/serializer"
	"github.com/rantuttl/cloudops/apimachinery/pkg/apimachinery"
	"github.com/rantuttl/cloudops/apimachinery/pkg/apimachinery/registered"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/rest"
	apirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
	genericapi "github.com/rantuttl/cloudops/apiserver/pkg/endpoints"
)

// FIXME (rantuttl): Stub for now
// Info about an API group.
type APIGroupInfo struct {
	GroupMeta apimachinery.GroupMeta
	// Info about the resources in this group. Its a map from version to resource to the storage.
	VersionedResourcesStorageMap map[string]map[string]rest.Storage

	// Scheme includes all of the types used by this group and how to convert between them (or
	// to convert objects from outside of this group that are accepted in this API).
	Scheme *runtime.Scheme

	NegotiatedSerializer runtime.NegotiatedSerializer

	// SubresourceGroupVersionKind contains the GroupVersionKind overrides for each subresource that is
	// accessible from this API group version.
	SubresourceGroupVersionKind map[string]schema.GroupVersionKind
}

type GenericAPIServer struct {
	SecureServingInfo *SecureServingInfo
	// numerical ports, set after listening
	effectiveSecurePort int
	Serializer runtime.NegotiatedSerializer
	Handler *APIServerHandler
	requestContextMapper apirequest.RequestContextMapper
	minRequestTimeout time.Duration
}

type DelegationTarget interface {
	// UnprotectedHandler returns a handler that is NOT protected by a normal chain
	UnprotectedHandler() http.Handler

	// RequestContextMapper returns the existing RequestContextMapper.  Because we cannot rewire all existing
	// uses of this function, this will be used in any delegating API server
	RequestContextMapper() apirequest.RequestContextMapper
}

func (s *GenericAPIServer) UnprotectedHandler() http.Handler {
	// when we delegate, we need the server we're delegating to choose whether or not to use gorestful
	return s.Handler.Director
}

func (s *GenericAPIServer) RequestContextMapper() apirequest.RequestContextMapper {
	return s.requestContextMapper
}

var EmptyDelegate = emptyDelegate{
	requestContextMapper: apirequest.NewRequestContextMapper(),
}

type emptyDelegate struct {
	requestContextMapper apirequest.RequestContextMapper
}

func (s emptyDelegate) UnprotectedHandler() http.Handler {
	return nil
}

func (s emptyDelegate) RequestContextMapper() apirequest.RequestContextMapper {
	return s.requestContextMapper
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
	// TODO (rantuttl): Setup secured API server utilizing handler chains defined in
	// package server: apiserver/pkg/genericserver/server/config.go:DefaultHandlerChainBuilder
	internalStopCh := make(chan struct{})

	if s.SecureServingInfo != nil && s.Handler != nil {
		if err := s.serveSecurely(internalStopCh); err != nil {
			close(internalStopCh)
			return err
		}
	}
	go func() {
		<-stopCh
		close(internalStopCh)
	}()

	// TODO (rantuttl): If we ever need post-start hooks, this should be where we instantiate them
	return nil
}

// installAPIResources is a private method for installing the REST storage backing each api groupversionresource
func (s *GenericAPIServer) installAPIResources(apiPrefix string, apiGroupInfo *APIGroupInfo) error {
        for _, groupVersion := range apiGroupInfo.GroupMeta.GroupVersions {
                if len(apiGroupInfo.VersionedResourcesStorageMap[groupVersion.Version]) == 0 {
                        glog.Warningf("Skipping API %v because it has no resources.", groupVersion)
                        continue
                }

                apiGroupVersion := s.getAPIGroupVersion(apiGroupInfo, groupVersion, apiPrefix)
		// FIXME (rantuttl): What is OptionsExternalVersion used for??
                //if apiGroupInfo.OptionsExternalVersion != nil {
                //        apiGroupVersion.OptionsExternalVersion = apiGroupInfo.OptionsExternalVersion
                //}

                if err := apiGroupVersion.InstallREST(s.Handler.GoRestfulContainer); err != nil {
                        return fmt.Errorf("Unable to setup API %v: %v", apiGroupInfo, err)
                }
        }

        return nil
}

// Exposes the given api group in the API.
func (s *GenericAPIServer) InstallAPIGroup(apiGroupInfo *APIGroupInfo) error {
	if len(apiGroupInfo.GroupMeta.GroupVersion.Group) == 0 {
		return fmt.Errorf("cannot register handler with an empty group for %#v", *apiGroupInfo)
	}
	if len(apiGroupInfo.GroupMeta.GroupVersion.Version) == 0 {
		return fmt.Errorf("cannot register handler with an empty version for %#v", *apiGroupInfo)
	}

        if err := s.installAPIResources(APIGroupPrefix, apiGroupInfo); err != nil {
                return err
        }

	return nil
}

func (s *GenericAPIServer) getAPIGroupVersion(apiGroupInfo *APIGroupInfo, groupVersion schema.GroupVersion, apiPrefix string) *genericapi.APIGroupVersion {
	storage := make(map[string]rest.Storage)
	for k, v := range apiGroupInfo.VersionedResourcesStorageMap[groupVersion.Version] {
		storage[strings.ToLower(k)] = v
	}
	return &genericapi.APIGroupVersion{
		Root:			apiPrefix,
		Storage:		storage,
		GroupVersion:		groupVersion,
		Mapper:			apiGroupInfo.GroupMeta.RESTMapper,
		Serializer:		apiGroupInfo.NegotiatedSerializer,
		Typer:			apiGroupInfo.Scheme,
		Creater:		apiGroupInfo.Scheme,
		Copier:			apiGroupInfo.Scheme,
		Convertor:		apiGroupInfo.Scheme,
		Defaulter:		apiGroupInfo.Scheme,
		Linker:			apiGroupInfo.GroupMeta.SelfLinker,
		Context:		s.RequestContextMapper(),
		SubresourceGroupVersionKind: apiGroupInfo.SubresourceGroupVersionKind,
		MinRequestTimeout:	s.minRequestTimeout,
	}
}

// NewDefaultAPIGroupInfo returns an APIGroupInfo stubbed with "normal" values
// exposed for easier composition from other packages
func NewDefaultAPIGroupInfo(group string, registry *registered.APIRegistrationManager, scheme *runtime.Scheme, parameterCodec runtime.ParameterCodec, codecs serializer.CodecFactory) APIGroupInfo {
	groupMeta := registry.GroupOrDie(group)

	return APIGroupInfo{
		GroupMeta:			*groupMeta,
		VersionedResourcesStorageMap:	map[string]map[string]rest.Storage{},
		Scheme:				scheme,
		NegotiatedSerializer:		codecs,
		SubresourceGroupVersionKind:	map[string]schema.GroupVersionKind{},
	}
}
