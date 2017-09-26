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

package validation

import (
	"github.com/rantuttl/cloudops/apiserver/pkg/apigroups/core"
	"github.com/rantuttl/cloudops/apimachinery/pkg/api/validation"
	"github.com/rantuttl/cloudops/apimachinery/pkg/util/validation/field"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	apimachineryvalidation "github.com/rantuttl/cloudops/apimachinery/pkg/api/validation"
)

type ValidateNameFunc apimachineryvalidation.ValidateNameFunc

var ValidateAccountName = apimachineryvalidation.NameIsDNSLabel

func ValidateAccount(account *core.Account) field.ErrorList {
	allErrs := ValidateObjectMeta(&account.ObjectMeta, false, ValidateAccountName, field.NewPath("metadata"))
	// TODO (rantuttl): Add any finalizer checking here

	return allErrs
}

// ValidateObjectMeta validates an object's metadata on creation. It expects that name generation has already
// been performed.
// It doesn't return an error for rootscoped resources with namespace, because namespace should already be cleared before.
func ValidateObjectMeta(meta *metav1.ObjectMeta, requiresNamespace bool, nameFn ValidateNameFunc, fldPath *field.Path) field.ErrorList {
        allErrs := validation.ValidateObjectMeta(meta, requiresNamespace, apimachineryvalidation.ValidateNameFunc(nameFn), fldPath)
        // run additional checks for the finalizer name
	// FIXME (rantuttl): Fix once we know if we need this or not

        return allErrs
}
