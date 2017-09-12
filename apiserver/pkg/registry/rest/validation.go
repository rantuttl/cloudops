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
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	genericapirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
)



// ValidNamespace returns false if the namespace on the context differs from
// the resource.  If the resource has no namespace, it is set to the value in
// the context.
//
// TODO(sttts): move into pkg/genericapiserver/endpoints
func ValidNamespace(ctx genericapirequest.Context, resource metav1.Object) bool {
	ns, ok := genericapirequest.NamespaceFrom(ctx)
	if len(resource.GetNamespace()) == 0 {
		resource.SetNamespace(ns)
	}
	return ns == resource.GetNamespace() && ok
}
