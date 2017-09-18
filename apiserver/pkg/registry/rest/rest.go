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
	"github.com/rantuttl/cloudops/apimachinery/pkg/watch"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	metainternalversion "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/internalversion"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	genericapirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
)

// It is expected that objects exported to the RESTful API of apiserver may implement any of the below interfaces.

// Storage is a generic interface for RESTful storage services.
// Resources which are exported to the RESTful API of apiserver need to implement this interface. It is expected
type Storage interface {
	// New returns an empty object that can be used with Create and Update after request data has been put into it.
	// This object must be a pointer type for use with Codec.DecodeInto([]byte, runtime.Object)
	New() runtime.Object
}

// Creater is an object that can create an instance of a RESTful object.
type Creater interface {
	// New returns an empty object that can be used with Create after request data has been put into it.
	// This object must be a pointer type for use with Codec.DecodeInto([]byte, runtime.Object)
	New() runtime.Object

	// Create creates a new version of a resource. If includeUninitialized is set, the object may be returned
	// without completing initialization.
	Create(ctx genericapirequest.Context, obj runtime.Object, includeUninitialized bool) (runtime.Object, error)
}

// NamedCreater is an object that can create an instance of a RESTful object using a name parameter.
type NamedCreater interface {
	// New returns an empty object that can be used with Create after request data has been put into it.
	// This object must be a pointer type for use with Codec.DecodeInto([]byte, runtime.Object)
	New() runtime.Object

	// Create creates a new version of a resource. It expects a name parameter from the path.
	// This is needed for create operations on subresources which include the name of the parent
	// resource in the path. If includeUninitialized is set, the object may be returned without
	// completing initialization.
	Create(ctx genericapirequest.Context, name string, obj runtime.Object, includeUninitialized bool) (runtime.Object, error)
}

// Getter is an object that can retrieve a named RESTful resource.
type Getter interface {
	// Get finds a resource in the storage by name and returns it.
	// Although it can return an arbitrary error value, IsNotFound(err) is true for the
	// returned error value err when the specified resource is not found.
	Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error)
}

// GracefulDeleter knows how to pass deletion options to allow delayed deletion of a
// RESTful object.
type GracefulDeleter interface {
	// Delete finds a resource in the storage and deletes it.
	// If options are provided, the resource will attempt to honor them or return an invalid
	// request error.
	// Although it can return an arbitrary error value, IsNotFound(err) is true for the
	// returned error value err when the specified resource is not found.
	// Delete *may* return the object that was deleted, or a status object indicating additional
	// information about deletion.
	// It also returns a boolean which is set to true if the resource was instantly
	// deleted or false if it will be deleted asynchronously.
	Delete(ctx genericapirequest.Context, name string, options *metav1.DeleteOptions) (runtime.Object, bool, error)
}

// GracefulDeleteAdapter adapts the Deleter interface to GracefulDeleter
type GracefulDeleteAdapter struct {
	Deleter
}

// Delete implements RESTGracefulDeleter in terms of Deleter
func (a GracefulDeleteAdapter) Delete(ctx genericapirequest.Context, name string, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	obj, err := a.Deleter.Delete(ctx, name)
	return obj, true, err
}

// Deleter is an object that can delete a named RESTful resource.
type Deleter interface {
	// Delete finds a resource in the storage and deletes it.
	// Although it can return an arbitrary error value, IsNotFound(err) is true for the
	// returned error value err when the specified resource is not found.
	// Delete *may* return the object that was deleted, or a status object indicating additional
	// information about deletion.
	Delete(ctx genericapirequest.Context, name string) (runtime.Object, error)
}

// Lister is an object that can retrieve resources that match the provided field and label criteria.
type Lister interface {
	// NewList returns an empty object that can be used with the List call.
	// This object must be a pointer type for use with Codec.DecodeInto([]byte, runtime.Object)
	NewList() runtime.Object
	// List selects resources in the storage which match to the selector. 'options' can be nil.
	List(ctx genericapirequest.Context, options *metainternalversion.ListOptions) (runtime.Object, error)
}

// Watcher should be implemented by all Storage objects that
// want to offer the ability to watch for changes through the watch api.
type Watcher interface {
	// 'label' selects on labels; 'field' selects on the object's fields. Not all fields
	// are supported; an error should be returned if 'field' tries to select on a field that
	// isn't supported. 'resourceVersion' allows for continuing/starting a watch at a
	// particular version.
	Watch(ctx genericapirequest.Context, options *metainternalversion.ListOptions) (watch.Interface, error)
}

// Exporter is an object that knows how to strip a RESTful resource for export
type Exporter interface {
	// Export an object.  Fields that are not user specified (e.g. Status, ObjectMeta.ResourceVersion) are stripped out
	// Returns the stripped object.  If 'exact' is true, fields that are specific to the cluster (e.g. namespace) are
	// retained, otherwise they are stripped also.
	Export(ctx genericapirequest.Context, name string, opts metav1.ExportOptions) (runtime.Object, error)
}

// StorageMetadata is an optional interface that callers can implement to provide additional
// information about their Storage objects.
type StorageMetadata interface {
	// ProducesMIMETypes returns a list of the MIME types the specified HTTP verb (GET, POST, DELETE,
	// PATCH) can respond with.
	ProducesMIMETypes(verb string) []string

	// ProducesObject returns an object the specified HTTP verb respond with. It will overwrite storage object if
	// it is not nil. Only the type of the return object matters, the value will be ignored.
	ProducesObject(verb string) interface{}
}
