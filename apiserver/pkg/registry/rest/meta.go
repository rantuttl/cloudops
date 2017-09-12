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
        "github.com/rantuttl/cloudops/apimachinery/pkg/util/uuid"
	"github.com/rantuttl/cloudops/apimachinery/pkg/types"
        genericapirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
)

// FillObjectMetaSystemFields populates fields that are managed by the system on ObjectMeta.
func FillObjectMetaSystemFields(ctx genericapirequest.Context, meta metav1.Object) {
        meta.SetCreationTimestamp(types.Now())
        // allows admission controllers to assign a UID earlier in the request processing
        // to support tracking resources pending creation.
        uid, found := genericapirequest.UIDFrom(ctx)
        if !found {
                uid = uuid.NewUUID()
        }
        meta.SetUID(uid)
        meta.SetSelfLink("")
}
