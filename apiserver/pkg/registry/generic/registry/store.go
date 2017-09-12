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

package registry

import (
	"fmt"
	"strings"

	"github.com/rantuttl/cloudops/apimachinery/pkg/api/meta"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	"github.com/rantuttl/cloudops/apimachinery/pkg/api/errors"
	"github.com/rantuttl/cloudops/apimachinery/pkg/api/validation/path"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/rest"
	"github.com/rantuttl/cloudops/apiserver/pkg/backend"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/generic"
	genericapirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
)

// ObjectFunc is a function to act on a given object. An error may be returned
// if the hook cannot be completed. An ObjectFunc may transform the provided
// object.
type ObjectFunc func(obj runtime.Object) error

type Store struct {
	// NewFunc returns a new instance of the type this registry returns for a
	// GET of a single object
	NewFunc func() runtime.Object

	// NewListFunc returns a new list of the type this registry; it is the
	// type returned when the resource is listed
	NewListFunc func() runtime.Object

	// QualifiedResource is the pluralized name of the resource.
	QualifiedResource schema.GroupResource

	// CreateStrategy implements resource-specific behavior during creation.
	CreateStrategy rest.RESTCreateStrategy
	// AfterCreate implements a further operation to run after a resource is
	// created and before it is decorated, optional.
	AfterCreate ObjectFunc


	// UpdateStrategy implements resource-specific behavior during updates.
	UpdateStrategy rest.RESTUpdateStrategy
	// AfterUpdate implements a further operation to run after a resource is
	// updated and before it is decorated, optional.
	AfterUpdate ObjectFunc

	// DeleteStrategy implements resource-specific behavior during deletion.
	DeleteStrategy rest.RESTDeleteStrategy
	// AfterDelete implements a further operation to run after a resource is
	// deleted and before it is decorated, optional.
	AfterDelete ObjectFunc

	// KeyRootFunc returns the root key for this resource; should not
	// include trailing "/".  This is used for operations that work on the
	// entire collection (listing and watching).
	//
	// KeyRootFunc and KeyFunc must be supplied together or not at all.
	KeyRootFunc func(ctx genericapirequest.Context) string

	// KeyFunc returns the key for a specific object in the collection.
	// KeyFunc is called for Create/Update/Get/Delete. Note that 'namespace'
	// can be gotten from ctx.
	//
	// KeyFunc and KeyRootFunc must be supplied together or not at all.
	KeyFunc func(ctx genericapirequest.Context, name string) (string, error)

	// ObjectNameFunc returns the name of an object or an error.
	ObjectNameFunc func(obj runtime.Object) (string, error)

	// TTLFunc returns the TTL (time to live) that objects should be persisted
	// with. The existing parameter is the current TTL or the default for this
	// operation. The update parameter indicates whether this is an operation
	// against an existing object.
	//
	// Objects that are persisted with a TTL are evicted once the TTL expires.
	TTLFunc func(obj runtime.Object, existing uint64, update bool) (uint64, error)

	Backend backend.Interface
}

// CompleteWithOptions updates the store with the provided options and
// defaults common fields.
func (e *Store) CompleteWithOptions(options *generic.StoreOptions) error {
	if e.QualifiedResource.Empty() {
		return fmt.Errorf("store %#v must have a non-empty qualified resource", e)
	}
	if e.NewFunc == nil {
		return fmt.Errorf("store for %s must have NewFunc set", e.QualifiedResource.String())
	}
	if e.NewListFunc == nil {
		return fmt.Errorf("store for %s must have NewListFunc set", e.QualifiedResource.String())
	}
	if (e.KeyRootFunc == nil) != (e.KeyFunc == nil) {
		return fmt.Errorf("store for %s must set both KeyRootFunc and KeyFunc or neither", e.QualifiedResource.String())
	}

	var isNamespaced bool
	switch {
	case e.CreateStrategy != nil:
		isNamespaced = e.CreateStrategy.NamespaceScoped()
	case e.UpdateStrategy != nil:
		isNamespaced = e.UpdateStrategy.NamespaceScoped()
	default:
		return fmt.Errorf("store for %s must have CreateStrategy or UpdateStrategy set", e.QualifiedResource.String())
	}
	if e.DeleteStrategy == nil {
		return fmt.Errorf("store for %s must have DeleteStrategy set", e.QualifiedResource.String())
	}
	if options.RESTOptions == nil {
		return fmt.Errorf("options for %s must have RESTOptions set", e.QualifiedResource.String())
	}
	// FIXME (rantuttl): Once we figure if we need this or not
	//if options.AttrFunc == nil {
	//	return fmt.Errorf("options for %s must have AttrFunc set", e.QualifiedResource.String())
	//}

	opts, err := options.RESTOptions.GetRESTOptions(e.QualifiedResource)
	if err != nil {
		return err
	}
	// Resource prefix must come from the underlying factory
	prefix := opts.ResourcePrefix
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if prefix == "/" {
		return fmt.Errorf("store for %s has an invalid prefix %q", e.QualifiedResource.String(), opts.ResourcePrefix)
	}

	// Set the default behavior for storage key generation if none provided
	if e.KeyRootFunc == nil && e.KeyFunc == nil {
		if isNamespaced {
			e.KeyRootFunc = func(ctx genericapirequest.Context) string {
				return NamespaceKeyRootFunc(ctx, prefix)
			}
			e.KeyFunc = func(ctx genericapirequest.Context, name string) (string, error) {
				return NamespaceKeyFunc(ctx, prefix, name)
			}
		} else {
			e.KeyRootFunc = func(ctx genericapirequest.Context) string {
				return prefix
			}
			e.KeyFunc = func(ctx genericapirequest.Context, name string) (string, error) {
				return NoNamespaceKeyFunc(ctx, prefix, name)
			}
		}
	}

	// Create an ObjectNameFunc if none provided
	if e.ObjectNameFunc == nil {
		e.ObjectNameFunc = func(obj runtime.Object) (string, error) {
			accessor, err := meta.Accessor(obj)
			if err != nil {
				return "", err
			}
			return accessor.GetName(), nil
		}
	}

	// TODO (rantuttl): Define how to reach backend store
	if e.Backend == nil {
		e.Backend = opts.Decorator(
			opts.BackendConfig,
		)
	}

	return nil
}

// NamespaceKeyRootFunc is the default function for constructing storage paths
// to resource directories enforcing namespace rules.
func NamespaceKeyRootFunc(ctx genericapirequest.Context, prefix string) string {
	key := prefix
	ns, ok := genericapirequest.NamespaceFrom(ctx)
	if ok && len(ns) > 0 {
		key = key + "/" + ns
	}
	return key
}

// NamespaceKeyFunc is the default function for constructing storage paths to
// a resource relative to the given prefix enforcing namespace rules. If the
// context does not contain a namespace, it errors.
func NamespaceKeyFunc(ctx genericapirequest.Context, prefix string, name string) (string, error) {
	key := NamespaceKeyRootFunc(ctx, prefix)
	ns, ok := genericapirequest.NamespaceFrom(ctx)
	if !ok || len(ns) == 0 {
		return "", errors.NewBadRequest("Namespace parameter required.")
	}
	if len(name) == 0 {
		return "", errors.NewBadRequest("Name parameter required.")
	}
	if msgs := path.IsValidPathSegmentName(name); len(msgs) != 0 {
		return "", errors.NewBadRequest(fmt.Sprintf("Name parameter invalid: %q: %s", name, strings.Join(msgs, ";")))
	}
	key = key + "/" + name
	return key, nil
}

// NoNamespaceKeyFunc is the default function for constructing storage paths
// to a resource relative to the given prefix without a namespace.
func NoNamespaceKeyFunc(ctx genericapirequest.Context, prefix string, name string) (string, error) {
	if len(name) == 0 {
		return "", errors.NewBadRequest("Name parameter required.")
	}
	if msgs := path.IsValidPathSegmentName(name); len(msgs) != 0 {
		return "", errors.NewBadRequest(fmt.Sprintf("Name parameter invalid: %q: %s", name, strings.Join(msgs, ";")))
	}
	key := prefix + "/" + name
	return key, nil
}

// New implements RESTStorage.New.
func (e *Store) New() runtime.Object {
	return e.NewFunc()
}

// NewList implements rest.Lister.
func (e *Store) NewList() runtime.Object {
	return e.NewListFunc()
}

// Create inserts a new item according to the unique key from the object.
func (e *Store) Create(ctx genericapirequest.Context, obj runtime.Object, includeUninitialized bool) (runtime.Object, error) {
	// BeforeCreate will also call CreateStrategy's PrepareForCreate()
	if err := rest.BeforeCreate(e.CreateStrategy, ctx, obj); err != nil {
		return nil, err
	}
	name, err := e.ObjectNameFunc(obj)
	if err != nil {
		return nil, err
	}
	key, err := e.KeyFunc(ctx, name)
	if err != nil {
		return nil, err
	}
	ttl, err := e.calculateTTL(obj, 0, false)
	if err != nil {
		return nil, err
	}

	out := e.NewFunc()
	if err := e.Backend.Create(ctx, key, obj, out, ttl); err != nil {
		// TODO (rantuttl): Maybe some better error type determinations, such as already exists, timeout, etc.
		return nil, err
	}
	if e.AfterCreate != nil {
		if err := e.AfterCreate(out); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (e *Store) Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	// FIXME (rantuttl): Stubbed for now
	obj := e.NewFunc()

	return obj, nil
}

// calculateTTL is a helper for retrieving the updated TTL for an object or
// returning an error if the TTL cannot be calculated. The defaultTTL is
// changed to 1 if less than zero. Zero means no TTL, not expire immediately.
func (e *Store) calculateTTL(obj runtime.Object, defaultTTL int64, update bool) (ttl uint64, err error) {
	// TODO: validate this is assertion is still valid.

	if defaultTTL < 0 {
		defaultTTL = 1
	}
	ttl = uint64(defaultTTL)
	if e.TTLFunc != nil {
		ttl, err = e.TTLFunc(obj, ttl, update)
	}
	return ttl, err
}
