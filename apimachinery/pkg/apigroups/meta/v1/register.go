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
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
)

// WatchEventKind is name reserved for serializing watch events.
const WatchEventKind = "WatchEvent"

func AddToGroupVersion(scheme *runtime.Scheme, groupVersion schema.GroupVersion) {
	// TODO (rantuttl): Handle watch events here in the future
	//scheme.AddKnownTypeWithName(groupVersion.WithKind(WatchEventKind), &WatchEvent{})
	//scheme.AddKnownTypeWithName(
	//	schema.GroupVersion{Group: groupVersion.Group, Version: runtime.APIVersionInternal}.WithKind(WatchEventKind),
	//	// FIXME (rantuttl): Add InternalEvent to apimachinery/pkg/apigroups/meta/v1/types.go
	//	&InternalEvent{},
	//)

	scheme.AddKnownTypes(groupVersion,
		// FIXME (rantuttl): Add ListOptions to apimachinery/pkg/apigroups/meta/v1/types.go
		//&ListOptions{},
		// TODO (rantuttl): Consider moving 'Status" to here from the likes of /apiserver/pkg/api/core/v1/register.go
		//&Status{},
		&ExportOptions{},
		&GetOptions{},
		&DeleteOptions{},
	)
	// See ./staging/src/k8s.io/apimachinery/pkg/apis/meta/v1/register.go for generic conversion functions
	// FUTURE FIXME (rantuttl): Add watcher conversion funcs as needed.

	scheme.AddGeneratedDeepCopyFuncs(GetGeneratedDeepCopyFuncs()...)
	//AddConversionFuncs(scheme)

}
