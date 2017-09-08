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
	"strings"

	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
)

// Implements RESTScope interface
type restScope struct {
	name	     RESTScopeName
	paramName	string
	argumentName     string
	paramDescription string
}

func (r *restScope) Name() RESTScopeName {
	return r.name
}
func (r *restScope) ParamName() string {
	return r.paramName
}
func (r *restScope) ArgumentName() string {
	return r.argumentName
}
func (r *restScope) ParamDescription() string {
	return r.paramDescription
}

var RESTScopeNamespace = &restScope{
	name:	     RESTScopeNameNamespace,
	paramName:	"namespaces",
	argumentName:     "namespace",
	paramDescription: "object name and auth scope, such as for teams and projects",
}

var RESTScopeRoot = &restScope{
	name: RESTScopeNameRoot,
}

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

// NewDefaultRESTMapper initializes a mapping between Kind and APIVersion
// to a resource name and back based on the objects in a runtime.Scheme
// and the API conventions. Takes a group name, a priority list of the versions
// to search when an object has no default version (set empty to return an error),
// and a function that retrieves the correct metadata for a given version.
func NewDefaultRESTMapper(defaultGroupVersions []schema.GroupVersion, f VersionInterfacesFunc) *DefaultRESTMapper {
	resourceToKind := make(map[schema.GroupVersionResource]schema.GroupVersionKind)
	kindToPluralResource := make(map[schema.GroupVersionKind]schema.GroupVersionResource)
	kindToScope := make(map[schema.GroupVersionKind]RESTScope)
	singularToPlural := make(map[schema.GroupVersionResource]schema.GroupVersionResource)
	pluralToSingular := make(map[schema.GroupVersionResource]schema.GroupVersionResource)

	return &DefaultRESTMapper{
		resourceToKind:		resourceToKind,
		kindToPluralResource:	kindToPluralResource,
		kindToScope:		kindToScope,
		defaultGroupVersions:	defaultGroupVersions,
		singularToPlural:	singularToPlural,
		pluralToSingular:	pluralToSingular,
		interfacesFunc:		f,
	}
}

func (m *DefaultRESTMapper) Add(kind schema.GroupVersionKind, scope RESTScope) {
	plural, singular := UnsafeGuessKindToResource(kind)

	m.singularToPlural[singular] = plural
	m.pluralToSingular[plural] = singular

	m.resourceToKind[singular] = kind
	m.resourceToKind[plural] = kind

	m.kindToPluralResource[kind] = plural
	m.kindToScope[kind] = scope
}

// unpluralizedSuffixes is a list of resource suffixes that are the same plural and singular
// This is only is only necessary because some bits of code are lazy and don't actually use the RESTMapper like they should.
// TODO eliminate this so that different callers can correctly map to resources.  This probably means updating all
// callers to use the RESTMapper they mean.
var unpluralizedSuffixes = []string{}

// UnsafeGuessKindToResource converts Kind to a resource name.
// Broken. This method only "sort of" works when used outside of this package.  It assumes that Kinds and Resources match
// and they aren't guaranteed to do so.
func UnsafeGuessKindToResource(kind schema.GroupVersionKind) ( /*plural*/ schema.GroupVersionResource /*singular*/, schema.GroupVersionResource) {
	kindName := kind.Kind
	if len(kindName) == 0 {
		return schema.GroupVersionResource{}, schema.GroupVersionResource{}
	}
	singularName := strings.ToLower(kindName)
	singular := kind.GroupVersion().WithResource(singularName)

	for _, skip := range unpluralizedSuffixes {
		if strings.HasSuffix(singularName, skip) {
			return singular, singular
		}
	}

	switch string(singularName[len(singularName)-1]) {
	case "s":
		return kind.GroupVersion().WithResource(singularName + "es"), singular
	case "y":
		return kind.GroupVersion().WithResource(strings.TrimSuffix(singularName, "y") + "ies"), singular
	}

	return kind.GroupVersion().WithResource(singularName + "s"), singular
}
