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

package install

import (
	"github.com/rantuttl/cloudops/apiserver/pkg/api"
	"github.com/rantuttl/cloudops/apiserver/pkg/apigroups/core"
	"github.com/rantuttl/cloudops/apiserver/pkg/apigroups/core/v1"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/util/sets"
	"github.com/rantuttl/cloudops/apimachinery/pkg/apimachinery/registered"
	"github.com/rantuttl/cloudops/apimachinery/pkg/apimachinery/announced"
)

func init() {
	Install(api.GroupFactoryRegistry, api.Registry, api.Scheme)
}

// Install registers the API group and adds types to a scheme
func Install(groupFactoryRegistry announced.APIGroupFactoryRegistry, registry *registered.APIRegistrationManager, scheme *runtime.Scheme) {
	if err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName:			core.GroupName,
			VersionPreferenceOrder:		[]string{v1.SchemeGroupVersion.Version},
			// package path to API resource Go types, e.g., "Account"
			//ImportPrefix:			"apiserver/pkg/apigroups/core",
			ImportPrefix:			"apiserver/pkg/api/core/v1",
			// the list of kinds that are scoped at the root of the api hierarchy; otherwise it's namespace scoped
			RootScopedKinds:		sets.NewString("Account", "AccountList"),
			AddInternalObjectsToScheme:	core.AddToScheme,
		},
		announced.VersionToSchemeFunc{
			v1.SchemeGroupVersion.Version:	v1.AddToScheme,
		},
	).Announce(groupFactoryRegistry).RegisterAndEnable(registry, scheme); err != nil {
		panic(err)
	}
}
