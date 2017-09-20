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

package master

import (
	"github.com/golang/glog"

	corev1 "github.com/rantuttl/cloudops/apiserver/pkg/api/core/v1"
	corerest "github.com/rantuttl/cloudops/apiserver/pkg/registry/core/rest"
	genericregistry "github.com/rantuttl/cloudops/apiserver/pkg/registry/generic"
	genericapiserver "github.com/rantuttl/cloudops/apiserver/pkg/genericserver/server"
	serverstorage "github.com/rantuttl/cloudops/apiserver/pkg/server/storage"
)

type Config struct {
	GenericConfig *genericapiserver.Config
	APIResourceConfigSource  serverstorage.APIResourceConfigSource
	StorageFactory           serverstorage.StorageFactory
}

type Master struct {
	GenericAPIServer         *genericapiserver.GenericAPIServer
}

type completedConfig struct {
        *Config
}

// Complete fills in any fields not set that are required to have valid data.
func (c *Config) Complete() completedConfig {
	c.GenericConfig.Complete()

	return completedConfig{c}
}

// SkipComplete provides a way to construct a server instance without config completion.
func (c *Config) SkipComplete() completedConfig {
	return completedConfig{c}
}

func (c completedConfig) New(delegate genericapiserver.DelegationTarget) (*Master, error) {

	glog.Info("Starting API Server")
	s, err := c.Config.GenericConfig.SkipComplete().New("apiserver", delegate) // completion is done in Complete, no need for a second time
	if err != nil {
		return nil, err
	}

	m := &Master{
		GenericAPIServer: s,
	}

	restStorageProviders := []RESTStorageProvider{
		corerest.RESTStorageProvider{},
	}
	m.InstallAPIs(c.Config.APIResourceConfigSource, c.Config.GenericConfig.RESTOptionsGetter, restStorageProviders...)

	return m, nil
}
// RESTStorageProvider is a factory type for REST storage.
type RESTStorageProvider interface {
        GroupName() string
        NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter genericregistry.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool)
}

// InstallAPIs will install the APIs for the restStorageProviders if they are enabled.
func (m *Master) InstallAPIs(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter genericregistry.RESTOptionsGetter, restStorageProviders ...RESTStorageProvider) {
	apiGroupsInfo := []genericapiserver.APIGroupInfo{}

	for _, restStorageBuilder := range restStorageProviders {
		groupName := restStorageBuilder.GroupName()

		if !apiResourceConfigSource.AnyResourcesForGroupEnabled(groupName) {
			glog.V(1).Infof("Skipping disabled API group %q.", groupName)
			continue
		}

		apiGroupInfo, enabled := restStorageBuilder.NewRESTStorage(apiResourceConfigSource, restOptionsGetter)
		if !enabled {
			glog.Warningf("Problem initializing API group \"%q\", skipping.", groupName)
			continue
		}

		// TODO (rantuttl): Figure out how to implement post-start hooks if necessary
		//

		apiGroupsInfo = append(apiGroupsInfo, apiGroupInfo)
	}

	for i := range apiGroupsInfo {
		if err := m.GenericAPIServer.InstallAPIGroup(&apiGroupsInfo[i]); err != nil {
			glog.Fatalf("Error in registering group versions: %v", err)
		}
	}
}

// Sets the default API Config.
// TODO (rantuttl): Consider a command line runtime option that can be used to merged command line options
// with any default settings. May aid testing new or modified APIs.
func DefaultAPIResourceConfigSource() *serverstorage.ResourceConfig {
	ret := serverstorage.NewResourceConfig()

	ret.EnableVersions(
		corev1.SchemeGroupVersion,
	)

	ret.EnableResources(
		corev1.SchemeGroupVersion.WithResource("accounts"),
	)
	return ret
}
