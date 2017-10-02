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

package endpoints

import (
	"fmt"
	"time"
	"net/http"
	"sort"
	"strings"
	"reflect"
	gpath "path"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"

	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	"github.com/rantuttl/cloudops/apimachinery/pkg/api/meta"
	"github.com/rantuttl/cloudops/apimachinery/pkg/conversion"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/rest"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/handlers"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/handlers/negotiation"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
)

const (
	ROUTE_META_GVK    = "x-cloudops-group-version-kind"
	ROUTE_META_ACTION = "x-cloudops-action"
)


type APIInstaller struct {
	group			*APIGroupVersion
	prefix			string
	minRequestTimeout	time.Duration
}

// Struct capturing information about an action (e.g., "GET", "POST", "DELETE", etc).
type action struct {
	Verb		string			// Verb identifying the action
	Path		string			// The path of the action
	Params		[]*restful.Parameter	// List of parameters associated with the action.
	Namer		handlers.ScopeNamer
	AllNamespaces	bool			// true iff the action is namespaced
}

// Installs handlers for API resources.
func (a *APIInstaller) Install(ws *restful.WebService) (apiResources []metav1.APIResource, errors []error) {
	errors = make([]error, 0)

	paths := make([]string, len(a.group.Storage))
	var i int = 0
	for path := range a.group.Storage {
		paths[i] = path
		i++
	}
	sort.Strings(paths)
	// FIXME (rantuttl): For now, we are not using proxyHandler
	for _, path := range paths {
		apiResource, err := a.registerResourceHandlers(path, a.group.Storage[path], ws)
		if err != nil {
			errors = append(errors, fmt.Errorf("error in registering resource: %s, %v", path, err))
		}
		if apiResource != nil {
			apiResources = append(apiResources, *apiResource)
		}
	}

	return apiResources, errors
}

func (a *APIInstaller) NewWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(a.prefix)
	ws.Doc("API at " + a.prefix)
	ws.Consumes("*/*")

	// This is actually performed in registerResourceHandlers, so not needed here??
	mediaTypes, streamMediaTypes := negotiation.MediaTypesForSerializer(a.group.Serializer)
	allMediaTypes := append(mediaTypes, streamMediaTypes...)
	ws.Produces(allMediaTypes...)

	ws.ApiVersion(a.group.GroupVersion.String())

	return ws
}

func (a *APIInstaller) registerResourceHandlers(path string, storage rest.Storage, ws *restful.WebService) (*metav1.APIResource, error) {
	// TODO (rantuttl): Handle addmission controls here
	context := a.group.Context
	if context == nil {
		return nil, fmt.Errorf("%v missing Context", a.group.GroupVersion)
	}


	optionsExternalVersion := a.group.GroupVersion

	resource, subresource, err := splitSubresource(path)
	if err != nil {
		return nil, err
	}
	hasSubresource := len(subresource) > 0

	mapping, err := a.restMapping(resource)
	if err != nil {
		return nil, err
	}

	fqKindToRegister, err := a.getResourceKind(path, storage)
	if err != nil {
		return nil, err
	}

	versionedPtr, err := a.group.Creater.New(fqKindToRegister)
	if err != nil {
		return nil, err
	}
	defaultVersionedObject := indirectArbitraryPointer(versionedPtr)

	creater, isCreater := storage.(rest.Creater)
	lister, isLister := storage.(rest.Lister)
	getter, isGetter := storage.(rest.Getter)
	deleter, isDeleter := storage.(rest.Deleter)
	gracefulDeleter, isGracefulDeleter := storage.(rest.GracefulDeleter)
	watcher, _ := storage.(rest.Watcher)
	storageMeta, isMetadata := storage.(rest.StorageMetadata)
	if !isMetadata {
		storageMeta = defaultStorageMetadata{}
	}
	exporter, isExporter := storage.(rest.Exporter)
	if !isExporter {
		exporter = nil
	}

	var versionedDeleteOptions runtime.Object
	var versionedDeleterObject interface{}
	switch {
	case isGracefulDeleter:
		versionedDeleteOptions, err = a.group.Creater.New(optionsExternalVersion.WithKind("DeleteOptions"))
		if err != nil {
			return nil, err
		}
		versionedDeleterObject = indirectArbitraryPointer(versionedDeleteOptions)
		isDeleter = true
	case isDeleter:
		gracefulDeleter = rest.GracefulDeleteAdapter{Deleter: deleter}
	}

	versionedStatusPtr, err := a.group.Creater.New(optionsExternalVersion.WithKind("Status"))
	if err != nil {
		return nil, err
	}
	versionedStatus := indirectArbitraryPointer(versionedStatusPtr)

	var versionedList interface{}
	if isLister {
		list := lister.NewList()
		listGVKs, _, err := a.group.Typer.ObjectKinds(list)
		if err != nil {
			return nil, err
		}
		versionedListPtr, err := a.group.Creater.New(a.group.GroupVersion.WithKind(listGVKs[0].Kind))
		if err != nil {
			return nil, err
		}
		versionedList = indirectArbitraryPointer(versionedListPtr)
	}

	var ctxFn handlers.ContextFunc
	ctxFn = func(req *http.Request) request.Context {
		if ctx, ok := context.Get(req); ok {
			return request.WithUserAgent(ctx, req.Header.Get("User-Agent"))
		}
		return request.WithUserAgent(request.NewContext(), req.Header.Get("User-Agent"))
	}

	resourceKind := fqKindToRegister.Kind
	scope := mapping.Scope
	nameParam := ws.PathParameter("name", "name of the "+resourceKind).DataType("string")

	params := []*restful.Parameter{}
	actions := []action{}

	var apiResource metav1.APIResource

	switch scope.Name() {
	case meta.RESTScopeNameRoot:
		resourcePath := resource
		resourceParams := params
		itemPath := resourcePath + "/{name}"
		nameParams := append(params, nameParam)
		suffix := ""
		if hasSubresource {
			suffix = "/" + subresource
			itemPath = itemPath + suffix
			resourcePath = itemPath
			resourceParams = nameParams
		}
		apiResource.Name = path
		apiResource.Namespaced = false
		apiResource.Kind = resourceKind
		namer := handlers.ContextBasedNaming{
			GetContext:		ctxFn,
			SelfLinker:		a.group.Linker,
			ClusterScoped:		true,
			SelfLinkPathPrefix:	gpath.Join(a.prefix, resource) + "/",
			SelfLinkPathSuffix:	suffix,
		}

		// Add actions at the resource path
		actions = appendIf(actions, action{"LIST", resourcePath, resourceParams, namer, false}, isLister)
		actions = appendIf(actions, action{"POST", resourcePath, resourceParams, namer, false}, isCreater)

		// Add actions at the item path
		actions = appendIf(actions, action{"GET", itemPath, nameParams, namer, false}, isGetter)
		actions = appendIf(actions, action{"DELETE", itemPath, nameParams, namer, false}, isDeleter)
		break
	//case meta.RESTScopeNameNamespace:
	default:
		return nil, fmt.Errorf("unsupported REST scope: %s", scope.Name())
	}

	mediaTypes, streamMediaTypes := negotiation.MediaTypesForSerializer(a.group.Serializer)
	allMediaTypes := append(mediaTypes, streamMediaTypes...)
	ws.Produces(allMediaTypes...)

	// Create the routes for the actions discovered

	// A request scope object for handling requests on this resource type within this API group
	reqScope := handlers.RequestScope{
		ContextFunc:		ctxFn,
		Serializer:		a.group.Serializer, // actual CodecFactory for the API group
		Creater:		a.group.Creater,
		Copier:			a.group.Copier, // used in update, i.e., PUT
		Convertor:		a.group.Convertor, // used in list
		Typer:			a.group.Typer,
		Defaulter:		a.group.Defaulter, // set the defaults on the API resource
		Resource:		a.group.GroupVersion.WithResource(resource),
		Subresource:		subresource,
		Kind:			fqKindToRegister,
	}


	for _, action := range actions {
		versionedObject := storageMeta.ProducesObject(action.Verb)
		// Is optional interface implemented
		if versionedObject == nil {
			versionedObject = defaultVersionedObject
		}

		reqScope.Namer = action.Namer
		namespaced := ""
		if apiResource.Namespaced {
			namespaced = "Namespaced"
		}
		operationSuffix := ""
		if strings.HasSuffix(action.Path, "/{path:*}") {
			operationSuffix = operationSuffix + "WithPath"
		}
		if action.AllNamespaces {
			operationSuffix = operationSuffix + "ForAllNamespaces"
			// no specific namespace
			namespaced = ""
		}

		routes := []*restful.RouteBuilder{}

		glog.V(5).Infof("Installing web service route handler for action: %v", action)
		switch action.Verb {
		case "GET": // Get a resource
			var handler restful.RouteFunction

			// Create a handler function for this group
			handler = restfulGetResource(getter, exporter, reqScope)
			doc := "read the specified " + resourceKind

			route := ws.GET(action.Path).To(handler).
				Doc(doc).
				Param(ws.QueryParameter("pretty", "If 'true', then the output is pretty printed.")).
				Operation("read"+namespaced+resourceKind+strings.Title(subresource)+operationSuffix).
				Produces(append(storageMeta.ProducesMIMETypes(action.Verb), mediaTypes...)...).
				Returns(http.StatusOK, "OK", versionedObject).
				Reads(versionedObject).
				Writes(versionedObject)
			addParams(route, action.Params)
			routes = append(routes, route)
		case "DELETE": // Delete a resource
			var handler restful.RouteFunction

			handler = restfulDeleteResource(gracefulDeleter, isGracefulDeleter, reqScope)
			doc := "delete" + resourceKind

			route := ws.DELETE(action.Path).To(handler).
				Doc(doc).
				Param(ws.QueryParameter("pretty", "If 'true', then the output is pretty printed.")).
				Operation("delete"+namespaced+resourceKind+strings.Title(subresource)+operationSuffix).
				Produces(append(storageMeta.ProducesMIMETypes(action.Verb), mediaTypes...)...).
				Writes(versionedStatus).
				Returns(http.StatusOK, "OK", versionedStatus)
			if isGracefulDeleter {
				route.Reads(versionedDeleterObject)
				if err := addObjectParams(ws, route, versionedDeleteOptions); err != nil {
					return nil, err
				}
			}
			addParams(route, action.Params)
			routes = append(routes, route)
		case "POST": // Create a resource
			var handler restful.RouteFunction

			// Create a handler function for this group
			handler = restfulCreateResource(creater, reqScope, a.group.Typer)
			doc := "create " + resourceKind

			route := ws.POST(action.Path).To(handler).
				Doc(doc).
				Param(ws.QueryParameter("pretty", "If 'true', then the output is pretty printed.")).
				Operation("create"+namespaced+resourceKind+strings.Title(subresource)+operationSuffix).
				Produces(append(storageMeta.ProducesMIMETypes(action.Verb), mediaTypes...)...).
				Returns(http.StatusOK, "OK", versionedObject).
				Reads(versionedObject).
				Writes(versionedObject)
			addParams(route, action.Params)
			routes = append(routes, route)
		case "LIST":
			var handler restful.RouteFunction

			// FIXME (rantuttl): Fix up for watcher later, especially docs and subresources.
			handler = restfulListResource(lister, watcher, reqScope, false, a.minRequestTimeout)
			doc := "list objects of kind " + resourceKind

			route := ws.GET(action.Path).To(handler).
				Doc(doc).
				Param(ws.QueryParameter("pretty", "If 'true', then the output is pretty printed.")).
				Operation("list"+namespaced+resourceKind+strings.Title(subresource)+operationSuffix).
				Produces(append(storageMeta.ProducesMIMETypes(action.Verb), mediaTypes...)...).
				Returns(http.StatusOK, "OK", versionedList).
				Reads(versionedList).
				Writes(versionedList)
			addParams(route, action.Params)
			routes = append(routes, route)

		default:
			return nil, fmt.Errorf("unrecognized action verb: %s", action.Verb)
		}

		// Loop over routes, and install them to our web service
		for _, route := range routes {
			route.Metadata(ROUTE_META_GVK, metav1.GroupVersionKind{
				Group:   reqScope.Kind.Group,
				Version: reqScope.Kind.Version,
				Kind:    reqScope.Kind.Kind,
			})
			route.Metadata(ROUTE_META_ACTION, strings.ToLower(action.Verb))
			ws.Route(route)
		}
	}

	return &apiResource, nil
}

func (a *APIInstaller) restMapping(resource string) (*meta.RESTMapping, error) {
	storage, ok := a.group.Storage[resource]
	if !ok {
		return nil, fmt.Errorf("unable to locate the storage object for resource: %s", resource)
	}
	fqKindToRegister, err := a.getResourceKind(resource, storage)
	if err != nil {
		return nil, fmt.Errorf("unable to locate fully qualified kind for mapper resource %s: %v", resource, err)
	}
	return a.group.Mapper.RESTMapping(fqKindToRegister.GroupKind(), fqKindToRegister.Version)
}

// splitSubresource checks if the given storage path is the path of a subresource and returns
// the resource and subresource components.
func splitSubresource(path string) (string, string, error) {
	var resource, subresource string

	switch parts := strings.Split(path, "/"); len(parts) {
	case 2:
		resource, subresource = parts[0], parts[1]
	case 1:
		resource = parts[0]
	default:
		// TODO: support deeper paths. For now, only 2 levels deep
		return "", "", fmt.Errorf("api_installer allows only one or two segment paths (resource or resource/subresource)")
	}
	return resource, subresource, nil
}

// getResourceKind returns the external group version kind registered for the given storage
// object. If the storage object is a subresource and has an override supplied for it, it returns
// the group version kind supplied in the override.
func (a *APIInstaller) getResourceKind(path string, storage rest.Storage) (schema.GroupVersionKind, error) {
	if fqKindToRegister, ok := a.group.SubresourceGroupVersionKind[path]; ok {
		return fqKindToRegister, nil
	}

	// see e.g., apiserver/pkg/registry/core/account/storage/storage.go
	object := storage.New()
	fqKinds, _, err := a.group.Typer.ObjectKinds(object)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	// a given go type can have multiple potential fully qualified kinds.  Find the one that corresponds with the group
	// we're trying to register here
	fqKindToRegister := schema.GroupVersionKind{}
	for _, fqKind := range fqKinds {
		if fqKind.Group == a.group.GroupVersion.Group {
			fqKindToRegister = a.group.GroupVersion.WithKind(fqKind.Kind)
			break
		}
	}
	if fqKindToRegister.Empty() {
		return schema.GroupVersionKind{}, fmt.Errorf("unable to locate fully qualified kind for %v: found %v when registering for %v", reflect.TypeOf(object), fqKinds, a.group.GroupVersion)
	}
	return fqKindToRegister, nil
}

func appendIf(actions []action, a action, shouldAppend bool) []action {
	if shouldAppend {
		actions = append(actions, a)
	}
	return actions
}

// This magic incantation returns *ptrToObject for an arbitrary pointer
func indirectArbitraryPointer(ptrToObject interface{}) interface{} {
	return reflect.Indirect(reflect.ValueOf(ptrToObject)).Interface()
}

// FIXME (rantuttl): Nothing uses this yet. Just a hijacked idea
// An interface to see if an object supports swagger documentation as a method
type documentable interface {
	SwaggerDoc() map[string]string
}

// addObjectParams converts a runtime.Object into a set of go-restful Param() definitions on the route.
// The object must be a pointer to a struct; only fields at the top level of the struct that are not
// themselves interfaces or structs are used; only fields with a json tag that is non empty (the standard
// Go JSON behavior for omitting a field) become query parameters. The name of the query parameter is
// the JSON field name. If a description struct tag is set on the field, that description is used on the
// query parameter. In essence, it converts a standard JSON top level object into a query param schema.
func addObjectParams(ws *restful.WebService, route *restful.RouteBuilder, obj interface{}) error {
	sv, err := conversion.EnforcePtr(obj)
	if err != nil {
		return err
	}
	st := sv.Type()
	switch st.Kind() {
	case reflect.Struct:
		for i := 0; i < st.NumField(); i++ {
			name := st.Field(i).Name
			sf, ok := st.FieldByName(name)
			if !ok {
				continue
			}
			switch sf.Type.Kind() {
			case reflect.Interface, reflect.Struct:
			case reflect.Ptr:
				// TODO: This is a hack to let metav1.Time through. This needs to be fixed in a more generic way eventually. bug #36191
				if (sf.Type.Elem().Kind() == reflect.Interface || sf.Type.Elem().Kind() == reflect.Struct) && strings.TrimPrefix(sf.Type.String(), "*") != "metav1.Time" {
					continue
				}
				fallthrough
			default:
				jsonTag := sf.Tag.Get("json")
				if len(jsonTag) == 0 {
					continue
				}
				jsonName := strings.SplitN(jsonTag, ",", 2)[0]
				if len(jsonName) == 0 {
					continue
				}

				var desc string
				if docable, ok := obj.(documentable); ok {
					desc = docable.SwaggerDoc()[jsonName]
				}
				route.Param(ws.QueryParameter(jsonName, desc).DataType(typeToJSON(sf.Type.String())))
			}
		}
	}
	return nil
}

// Convert the name of a golang type to the name of a JSON type
func typeToJSON(typeName string) string {
	switch typeName {
	case "bool", "*bool":
		return "boolean"
	case "uint8", "*uint8", "int", "*int", "int32", "*int32", "int64", "*int64", "uint32", "*uint32", "uint64", "*uint64":
		return "integer"
	case "float64", "*float64", "float32", "*float32":
		return "number"
	case "metav1.Time", "*metav1.Time":
		return "string"
	case "byte", "*byte":
		return "string"
	case "v1.DeletionPropagation", "*v1.DeletionPropagation":
		return "string"

	// TODO: Fix these when go-restful supports a way to specify an array query param:
	// https://github.com/emicklei/go-restful/issues/225
	case "[]string", "[]*string":
		return "string"
	case "[]int32", "[]*int32":
		return "integer"

	default:
		return typeName
	}
}

// TODO (rantuttl): add admission control capabilities
// FIXME (rantuttl): 'Typer' already sent in scope object. Remove from this and associated method signatures
func restfulCreateResource(r rest.Creater, scope handlers.RequestScope, typer runtime.ObjectTyper) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handlers.CreateResource(r, scope, typer)(res.ResponseWriter, req.Request)
	}
}

func restfulGetResource(r rest.Getter, e rest.Exporter, scope handlers.RequestScope) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handlers.GetResource(r, e, scope)(res.ResponseWriter, req.Request)
	}
}

func restfulListResource(r rest.Lister, rw rest.Watcher, scope handlers.RequestScope, forceWatch bool, minRequestTimeout time.Duration) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handlers.ListResource(r, rw, scope, forceWatch, minRequestTimeout)(res.ResponseWriter, req.Request)
	}
}

func restfulDeleteResource(r rest.GracefulDeleter, allowsOptions bool, scope handlers.RequestScope) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handlers.DeleteResource(r, allowsOptions, scope)(res.ResponseWriter, req.Request)
	}
}

// defaultStorageMetadata provides default answers to rest.StorageMetadata.
type defaultStorageMetadata struct{}

// defaultStorageMetadata implements rest.StorageMetadata
var _ rest.StorageMetadata = defaultStorageMetadata{}

func (defaultStorageMetadata) ProducesObject(verb string) interface{} {
	return nil
}

func (defaultStorageMetadata) ProducesMIMETypes(verb string) []string {
	return nil
}

func addParams(route *restful.RouteBuilder, params []*restful.Parameter) {
	for _, param := range params {
		route.Param(param)
	}
}
