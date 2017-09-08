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

package meta

import (
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
)

// DefaultRESTMapper exposes mappings between the types defined in a
// runtime.Scheme. It assumes that all types defined the provided scheme
// can be mapped with the provided MetadataAccessor and Codec interfaces.
//
// The resource name of a Kind is defined as the lowercase,
// English-plural version of the Kind string.
// When converting from resource to Kind, the singular version of the
// resource name is also accepted for convenience.

type DefaultRESTMapper struct {
	defaultGroupVersions	[]schema.GroupVersion
	resourceToKind		map[schema.GroupVersionResource]schema.GroupVersionKind
	kindToPluralResource	map[schema.GroupVersionKind]schema.GroupVersionResource
	kindToScope		map[schema.GroupVersionKind]RESTScope
	singularToPlural	map[schema.GroupVersionResource]schema.GroupVersionResource
	pluralToSingular	map[schema.GroupVersionResource]schema.GroupVersionResource
	interfacesFunc		VersionInterfacesFunc
}

// VersionInterfacesFunc returns the appropriate typer, and metadata accessor for a
// given api version, or an error if no such api version exists.
type VersionInterfacesFunc func(version schema.GroupVersion) (*VersionInterfaces, error)

var _ RESTMapper = &DefaultRESTMapper{}

func (m *DefaultRESTMapper) KindsFor(input schema.GroupVersionResource) ([]schema.GroupVersionKind, error) {
	// FIXME (rantuttl): Stub for now.

	ret := []schema.GroupVersionKind{}
	return ret, nil
}

func (m *DefaultRESTMapper) KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	kinds, err := m.KindsFor(resource)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	if len(kinds) == 1 {
		return kinds[0], nil
	}

	return schema.GroupVersionKind{}, &AmbiguousResourceError{PartialResource: resource, MatchingKinds: kinds}
}
