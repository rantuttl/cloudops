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
	"github.com/rantuttl/cloudops/apimachinery/pkg/api/errors"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	"github.com/rantuttl/cloudops/apimachinery/pkg/api/meta"
        metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
        "github.com/rantuttl/cloudops/apimachinery/pkg/util/uuid"
        genericapirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
)

// FillObjectMetaSystemFields populates fields that are managed by the system on ObjectMeta.
func FillObjectMetaSystemFields(ctx genericapirequest.Context, meta metav1.Object) {
        meta.SetCreationTimestamp(metav1.Now())
	// TODO (rantuttl): Admission controllers aren't in yet, so this is in expectation of that.
        // allows admission controllers to assign a UID earlier in the request processing
        // to support tracking resources pending creation.
        uid, found := genericapirequest.UIDFrom(ctx)
        if !found {
                uid = uuid.NewUUID()
        }
        meta.SetUID(uid)
        meta.SetSelfLink("")
}

// objectMetaAndKind retrieves kind and ObjectMeta from a runtime object, or returns an error.
func objectMetaAndKind(typer runtime.ObjectTyper, obj runtime.Object) (metav1.Object, schema.GroupVersionKind, error) {
	objectMeta, err := meta.Accessor(obj)
	if err != nil {
		return nil, schema.GroupVersionKind{}, errors.NewInternalError(err)
	}
	kinds, _, err := typer.ObjectKinds(obj)
	if err != nil {
		return nil, schema.GroupVersionKind{}, errors.NewInternalError(err)
	}
	return objectMeta, kinds[0], nil
}
