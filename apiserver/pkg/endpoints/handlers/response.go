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

package handlers

import (
	"net/http"

	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/handlers/responsewriters"
)

// transformResponseObject takes a loaded object and performs any necessary transformations.
func transformResponseObject(ctx request.Context, scope RequestScope, req *http.Request, w http.ResponseWriter, statusCode int, result runtime.Object) {
	// TODO (rantuttl): Put any transformations needed here

	responsewriters.WriteObject(ctx, statusCode, scope.Kind.GroupVersion(), scope.Serializer, result, w, req)
}
