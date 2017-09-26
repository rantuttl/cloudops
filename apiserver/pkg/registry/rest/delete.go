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
	"fmt"

	"github.com/rantuttl/cloudops/apimachinery/pkg/api/errors"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	genericapirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
)

// RESTDeleteStrategy defines deletion behavior on an object that follows API conventions.
type RESTDeleteStrategy interface {
	runtime.ObjectTyper
}

func BeforeDelete(strategy RESTDeleteStrategy, ctx genericapirequest.Context, obj runtime.Object, options *metav1.DeleteOptions) (graceful, gracefulPending bool, err error) {
	objectMeta, gvk, err := objectMetaAndKind(strategy, obj)
	if err != nil {
		return false, false, err
	}

	// Checking the Preconditions here to fail early. It will be check just before deletion too.
	if options.Preconditions != nil && options.Preconditions.UID != nil && *options.Preconditions.UID != objectMeta.GetUID() {
		return false, false, errors.NewConflict(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, objectMeta.GetName(), fmt.Errorf("the UID in the precondition (%s) does not match the UID in record (%s). The object might have been deleted and then recreated", *options.Preconditions.UID, objectMeta.GetUID()))
	}

	// TODO (rantuttl): Implement graceful deletion

	return false, false, nil
}
