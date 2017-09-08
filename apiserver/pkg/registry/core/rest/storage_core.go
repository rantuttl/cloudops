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

package rest

import (
	"github.com/rantuttl/cloudops/apiserver/pkg/api"
	"github.com/rantuttl/cloudops/apiserver/pkg/apigroups/core"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/rest"
	v1core "github.com/rantuttl/cloudops/apiserver/pkg/apigroups/core/v1"
	serverstorage "github.com/rantuttl/cloudops/apiserver/pkg/server/storage"
	genericregistry "github.com/rantuttl/cloudops/apiserver/pkg/registry/generic"
	genericapiserver "github.com/rantuttl/cloudops/apiserver/pkg/genericserver/server"
)

type RESTStorageProvider struct {
}

func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter genericregistry.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(p.GroupName(), api.Registry, api.Scheme, api.ParameterCodec, api.Codecs)
	if apiResourceConfigSource.AnyResourcesForVersionEnabled(v1core.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[v1core.SchemeGroupVersion.Version] = p.v1Storage(apiResourceConfigSource, restOptionsGetter)
		apiGroupInfo.GroupMeta.GroupVersion = v1core.SchemeGroupVersion
	}

	return apiGroupInfo, true
}

func (p RESTStorageProvider) v1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter genericregistry.RESTOptionsGetter) map[string]rest.Storage {
	//version := v1core.SchemeGroupVersion

	storage := map[string]rest.Storage{}

	return storage
}

func (p RESTStorageProvider) GroupName() string {
	return core.GroupName
}
