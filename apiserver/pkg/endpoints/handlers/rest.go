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
	"fmt"
	"time"
	"net/http"
	//"net/url"
	"io/ioutil"
	"encoding/hex"
	//"github.com/golang/glog"

	"github.com/rantuttl/cloudops/apimachinery/pkg/api/errors"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	utilruntime "github.com/rantuttl/cloudops/apimachinery/pkg/util/runtime"
	metainternalversion "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/internalversion"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/rest"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/handlers/negotiation"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/handlers/responsewriters"
)

// RequestScope encapsulates common fields across all RESTful handler methods.
type RequestScope struct {
	Namer ScopeNamer
	ContextFunc

	Serializer runtime.NegotiatedSerializer
	runtime.ParameterCodec

	Creater         runtime.ObjectCreater
	Convertor       runtime.ObjectConvertor
	Defaulter       runtime.ObjectDefaulter
	Copier          runtime.ObjectCopier
	Typer           runtime.ObjectTyper
	// FIXME (rantuttl)
	//UnsafeConvertor runtime.ObjectConvertor

	// FIXME (rantuttl)
	//TableConvertor rest.TableConvertor

	Resource    schema.GroupVersionResource
	Kind        schema.GroupVersionKind
	Subresource string

	MetaGroupVersion schema.GroupVersion
}

func (scope *RequestScope) err(err error, w http.ResponseWriter, req *http.Request) {
	ctx := scope.ContextFunc(req)
	responsewriters.ErrorNegotiated(ctx, err, scope.Serializer, scope.Kind.GroupVersion(), w, req)
}

// CreateResource returns a function that will handle a resource creation.
// FIXME (rantuttl): 'Typer' already sent in scope object. Remove from this and associated method signatures
func CreateResource(r rest.Creater, scope RequestScope, typer runtime.ObjectTyper) http.HandlerFunc {
	return createHandler(&namedCreaterAdapter{r}, scope, typer, false)
}

type namedCreaterAdapter struct {
	rest.Creater
}

func (c *namedCreaterAdapter) Create(ctx request.Context, name string, obj runtime.Object, includeUninitialized bool) (runtime.Object, error) {
	return c.Creater.Create(ctx, obj, includeUninitialized)
}

// FIXME (rantuttl): 'Typer' already sent in scope object. Remove from this and associated method signatures
func createHandler(r rest.NamedCreater, scope RequestScope, typer runtime.ObjectTyper, includeName bool) http.HandlerFunc {
	return func (w http.ResponseWriter, req *http.Request) {

		var (
			namespace, name string
			err             error
		)
		if includeName {
			namespace, name, err = scope.Namer.Name(req)
		} else {
			namespace, err = scope.Namer.Namespace(req)
		}

		ctx := scope.ContextFunc(req)
		ctx = request.WithNamespace(ctx, namespace)

		s, err := negotiation.NegotiateInputSerializer(req, scope.Serializer)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		gv := scope.Kind.GroupVersion()
		decoder := scope.Serializer.DecoderToVersion(s.Serializer, schema.GroupVersion{Group: gv.Group, Version: runtime.APIVersionInternal})

		body, err := readBody(req)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		expectedGVK := scope.Kind
		original := r.New()
		obj, gvk, err := decoder.Decode(body, &expectedGVK, original)
		if err != nil {
			// FIXME (rantuttl): 'Typer' already sent in scope object. Reference and pass through here.
			err = transformDecodeError(typer, err, original, gvk, body)
			scope.err(err, w, req)
			return
		}
		if gvk.GroupVersion() != gv {
			err = errors.NewBadRequest(fmt.Sprintf("the API version in the data (%s) does not match the expected API version (%v)", gvk.GroupVersion().String(), gv.String()))
			scope.err(err, w, req)
			return
		}
		// TODO (rantuttl): Install admission control mechanisms here to permit this operation.

		includeUninitialized := req.URL.Query().Get("includeUninitialized") == "1"

		// TODO (rantuttl): Decide how we want to handle establishing timeout values. For now, hardcode,
		// but could provide via the API installation, either through the group registration and/or via a default setting.
		timeout := 30 * time.Second
		result, err := finishRequest(timeout, func() (runtime.Object, error) {
			return r.Create(ctx, name, obj, includeUninitialized)
		})
		if err != nil {
			scope.err(err, w, req)
			return
		}

		requestInfo, ok := request.RequestInfoFrom(ctx)
		if !ok {
			err := fmt.Errorf("missing requestInfo")
			scope.err(err, w, req)
			return
		}

		if err := setSelfLink(result, requestInfo, scope.Namer); err != nil {
			scope.err(err, w, req)
			return
		}

		code := http.StatusCreated

		transformResponseObject(ctx, scope, req, w, code, result)
	}
}

// GetResource returns a function that handles retrieving a single resource from a rest.Storage object.
func GetResource(r rest.Getter, e rest.Exporter, scope RequestScope) http.HandlerFunc {
	return getResourceHandler(scope,
		func(ctx request.Context, name string, req *http.Request) (runtime.Object, error) {
			// check for export
			options := metav1.GetOptions{}
			if values := req.URL.Query(); len(values) > 0 {
				exports := metav1.ExportOptions{}
				if err := metainternalversion.ParameterCodec.DecodeParameters(values, scope.MetaGroupVersion, &exports); err != nil {
					err = errors.NewBadRequest(err.Error())
					return nil, err
				}
				if exports.Export {
					if e == nil {
						return nil, errors.NewBadRequest(fmt.Sprintf("export of %q is not supported", scope.Resource.Resource))
					}
					return e.Export(ctx, name, exports)
				}
				// check for other options
				if err := metainternalversion.ParameterCodec.DecodeParameters(values, scope.MetaGroupVersion, &options); err != nil {
					err = errors.NewBadRequest(err.Error())
					return nil, err
				}
			}
			return r.Get(ctx, name, &options)
		})
}

// getResourceHandler is an HTTP handler function for get requests. It delegates to the
// passed-in getterFunc to perform the actual get.
func getResourceHandler(scope RequestScope, getter getterFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		namespace, name, err := scope.Namer.Name(req)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		ctx := scope.ContextFunc(req)
		ctx = request.WithNamespace(ctx, namespace)

		result, err := getter(ctx, name, req)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		requestInfo, ok := request.RequestInfoFrom(ctx)
		if !ok {
			scope.err(fmt.Errorf("missing requestInfo"), w, req)
			return
		}
		if err := setSelfLink(result, requestInfo, scope.Namer); err != nil {
			scope.err(err, w, req)
			return
		}

		transformResponseObject(ctx, scope, req, w, http.StatusOK, result)
	}
}

func ListResource(r rest.Lister, rw rest.Watcher, scope RequestScope, forceWatch bool, minRequestTimeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
	}
}

// DeleteResource returns a function that will handle a resource deletion
func DeleteResource(r rest.GracefulDeleter, allowsOptions bool, scope RequestScope) http.HandlerFunc {
        return func(w http.ResponseWriter, req *http.Request) {
	}
}

// TODO (rantuttl): Stubbed for now.
// setSelfLink sets the self link of an object (or the child items in a list) to the base URL of the request
// plus the path and query generated by the provided linkFunc
func setSelfLink(obj runtime.Object, requestInfo *request.RequestInfo, namer ScopeNamer) error {

	return nil
}

func summarizeData(data []byte, maxLength int) string {
	switch {
	case len(data) == 0:
		return "<empty>"
	case data[0] == '{':
		if len(data) > maxLength {
			return string(data[:maxLength]) + " ..."
		}
		return string(data)
	default:
		if len(data) > maxLength {
			return hex.EncodeToString(data[:maxLength]) + " ..."
		}
		return hex.EncodeToString(data)
	}
}

func readBody(req *http.Request) ([]byte, error) {
	defer req.Body.Close()
	return ioutil.ReadAll(req.Body)
}

// getterFunc performs a get request with the given context and object name. The request
// may be used to deserialize an options object to pass to the getter.
type getterFunc func(ctx request.Context, name string, req *http.Request) (runtime.Object, error)

// resultFunc is a function that returns a rest result and can be run in a goroutine
type resultFunc func() (runtime.Object, error)

// finishRequest makes a given resultFunc asynchronous and handles errors returned by the response.
// An api.Status object with status != success is considered an "error", which interrupts the normal response flow.
func finishRequest(timeout time.Duration, fn resultFunc) (result runtime.Object, err error) {
	// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
	// when the select statement reads something other than the one the goroutine sends on.
	ch := make(chan runtime.Object, 1)
	errCh := make(chan error, 1)
	panicCh := make(chan interface{}, 1)
	go func() {
		// panics don't cross goroutine boundaries, so we have to handle ourselves
		defer utilruntime.HandleCrash(func(panicReason interface{}) {
			// Propagate to parent goroutine
			panicCh <- panicReason
		})

		if result, err := fn(); err != nil {
			errCh <- err
		} else {
			ch <- result
		}
	}()

	select {
	case result = <-ch:
		if status, ok := result.(*metav1.Status); ok {
			if status.Status != metav1.StatusSuccess {
				return nil, errors.FromObject(status)
			}
		}
		return result, nil
	case err = <-errCh:
		return nil, err
	case p := <-panicCh:
		panic(p)
	case <-time.After(timeout):
		return nil, errors.NewTimeoutError("request did not complete within allowed duration", 0)
	}
}

// transformDecodeError adds additional information when a decode fails.
func transformDecodeError(typer runtime.ObjectTyper, baseErr error, into runtime.Object, gvk *schema.GroupVersionKind, body []byte) error {
	objGVKs, _, err := typer.ObjectKinds(into)
	if err != nil {
		return err
	}
	objGVK := objGVKs[0]
	if gvk != nil && len(gvk.Kind) > 0 {
		return errors.NewBadRequest(fmt.Sprintf("%s in version %q cannot be handled as a %s: %v", gvk.Kind, gvk.Version, objGVK.Kind, baseErr))
	}
	summary := summarizeData(body, 30)
	return errors.NewBadRequest(fmt.Sprintf("the object provided is unrecognized (must be of type %s): %v (%s)", objGVK.Kind, baseErr, summary))
}
