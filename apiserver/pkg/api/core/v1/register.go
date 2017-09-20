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
	"github.com/rantuttl/cloudops/apiserver/pkg/apigroups/core"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	v1meta "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: core.GroupName, Version: "v1"}

var (
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	localSchemeBuilder = &SchemeBuilder
	AddToScheme   = localSchemeBuilder.AddToScheme
)

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
        scheme.AddKnownTypes(SchemeGroupVersion,
                &core.Account{},
                &core.AccountList{},
		// Add more types to group 'core' as needed. See also apiserver/pkg/apigroups/core/v1/types.go
        )

	// Add common types
	scheme.AddKnownTypes(SchemeGroupVersion, &v1meta.Status{})

	// Do some extra stuff for our scheme
	v1meta.AddToGroupVersion(scheme, SchemeGroupVersion)
        return nil
}
