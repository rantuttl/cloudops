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
	"strings"

	"github.com/rantuttl/cloudops/apimachinery/pkg/api/meta"
	"github.com/rantuttl/cloudops/apimachinery/pkg/util/validation"
	"github.com/rantuttl/cloudops/apimachinery/pkg/util/validation/field"
	v1validation "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1/validation"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
)

const totalAnnotationSizeLimitB int = 256 * (1 << 10) // 256 kB

// ValidateClusterName can be used to check whether the given cluster name is valid.
var ValidateClusterName = NameIsDNS1035Label

// ValidateAnnotations validates that a set of annotations are correctly defined.
func ValidateAnnotations(annotations map[string]string, fldPath *field.Path) field.ErrorList {
        allErrs := field.ErrorList{}
        var totalSize int64
        for k, v := range annotations {
                for _, msg := range validation.IsQualifiedName(strings.ToLower(k)) {
                        allErrs = append(allErrs, field.Invalid(fldPath, k, msg))
                }
                totalSize += (int64)(len(k)) + (int64)(len(v))
        }
        if totalSize > (int64)(totalAnnotationSizeLimitB) {
                allErrs = append(allErrs, field.TooLong(fldPath, "", totalAnnotationSizeLimitB))
        }
        return allErrs
}

// ValidateObjectMeta validates an object's metadata on creation. It expects that name generation has already
// been performed.
// It doesn't return an error for rootscoped resources with namespace, because namespace should already be cleared before.
func ValidateObjectMetaAccessor(meta metav1.Object, requiresNamespace bool, nameFn ValidateNameFunc, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(meta.GetGenerateName()) != 0 {
		for _, msg := range nameFn(meta.GetGenerateName(), true) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("generateName"), meta.GetGenerateName(), msg))
		}
	}
	// If the generated name validates, but the calculated value does not, it's a problem with generation, and we
	// report it here. This may confuse users, but indicates a programming bug and still must be validated.
	// If there are multiple fields out of which one is required then add an or as a separator
	if len(meta.GetName()) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), "name or generateName is required"))
	} else {
		for _, msg := range nameFn(meta.GetName(), false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), meta.GetName(), msg))
		}
	}
	if requiresNamespace {
		if len(meta.GetNamespace()) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("namespace"), ""))
		} else {
			for _, msg := range ValidateNamespaceName(meta.GetNamespace(), false) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("namespace"), meta.GetNamespace(), msg))
			}
		}
	} else {
		if len(meta.GetNamespace()) != 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("namespace"), "not allowed on this type"))
		}
	}
	if len(meta.GetClusterName()) != 0 {
		for _, msg := range ValidateClusterName(meta.GetClusterName(), false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterName"), meta.GetClusterName(), msg))
		}
	}
	allErrs = append(allErrs, ValidateNonnegativeField(meta.GetGeneration(), fldPath.Child("generation"))...)
	allErrs = append(allErrs, v1validation.ValidateLabels(meta.GetLabels(), fldPath.Child("labels"))...)
	allErrs = append(allErrs, ValidateAnnotations(meta.GetAnnotations(), fldPath.Child("annotations"))...)
	// FIXME (rantuttl): Fix once we know if we need this or not
	//allErrs = append(allErrs, ValidateOwnerReferences(meta.GetOwnerReferences(), fldPath.Child("ownerReferences"))...)
	//allErrs = append(allErrs, ValidateInitializers(meta.GetInitializers(), fldPath.Child("initializers"))...)
	//allErrs = append(allErrs, ValidateFinalizers(meta.GetFinalizers(), fldPath.Child("finalizers"))...)

	return allErrs
}

// ValidateObjectMeta validates an object's metadata on creation. It expects that name generation has already
// been performed.
// It doesn't return an error for rootscoped resources with namespace, because namespace should already be cleared before.
func ValidateObjectMeta(objMeta *metav1.ObjectMeta, requiresNamespace bool, nameFn ValidateNameFunc, fldPath *field.Path) field.ErrorList {
        metadata, err := meta.Accessor(objMeta)
        if err != nil {
                allErrs := field.ErrorList{}
                allErrs = append(allErrs, field.Invalid(fldPath, objMeta, err.Error()))
                return allErrs
        }
        return ValidateObjectMetaAccessor(metadata, requiresNamespace, nameFn, fldPath)
}
