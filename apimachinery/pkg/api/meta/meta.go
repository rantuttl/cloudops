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
	"fmt"

	v1meta "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	"github.com/rantuttl/cloudops/apimachinery/pkg/types"
)

// errNotList is returned when an object implements the Object style interfaces but not the List style
// interfaces.
var errNotList = fmt.Errorf("object does not implement the List interfaces")

// ListAccessor returns a List interface for the provided object or an error if the object does
// not provide List.
// IMPORTANT: Objects are a superset of lists, so all Objects return List metadata. Do not use this
// check to determine whether an object *is* a List.
// TODO: return bool instead of error
func ListAccessor(obj interface{}) (List, error) {
	switch t := obj.(type) {
	case List:
		return t, nil
	case v1meta.List:
		return t, nil
	case ListMetaAccessor:
		if m := t.GetListMeta(); m != nil {
			return m, nil
		}
		return nil, errNotList
	case v1meta.ListMetaAccessor:
		if m := t.GetListMeta(); m != nil {
			return m, nil
		}
		return nil, errNotList
	case v1meta.Object:
		return t, nil
	case v1meta.ObjectMetaAccessor:
		if m := t.GetObjectMeta(); m != nil {
			return m, nil
		}
		return nil, errNotList
	default:
		return nil, errNotList
	}
}

// errNotObject is returned when an object implements the List style interfaces but not the Object style
// interfaces.
var errNotObject = fmt.Errorf("object does not implement the Object interfaces")

// Accessor takes an arbitrary object pointer and returns meta.Interface.
// obj must be a pointer to an API type. An error is returned if the minimum
// required fields are missing. Fields that are not required return the default
// value and are a no-op if set.
// TODO: return bool instead of error
func Accessor(obj interface{}) (v1meta.Object, error) {
	switch t := obj.(type) {
	case v1meta.Object:
		return t, nil
	case v1meta.ObjectMetaAccessor:
		if m := t.GetObjectMeta(); m != nil {
			return m, nil
		}
		return nil, errNotObject
	default:
		return nil, errNotObject
	}
}

type objectAccessor struct {
	runtime.Object
}

func (obj objectAccessor) GetKind() string {
	return obj.GetObjectKind().GroupVersionKind().Kind
}

func (obj objectAccessor) SetKind(kind string) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	gvk.Kind = kind
	obj.GetObjectKind().SetGroupVersionKind(gvk)
}

func (obj objectAccessor) GetAPIVersion() string {
	return obj.GetObjectKind().GroupVersionKind().GroupVersion().String()
}

func (obj objectAccessor) SetAPIVersion(version string) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	gv, err := schema.ParseGroupVersion(version)
	if err != nil {
		gv = schema.GroupVersion{Version: version}
	}
	gvk.Group, gvk.Version = gv.Group, gv.Version
	obj.GetObjectKind().SetGroupVersionKind(gvk)
}

// NewAccessor returns a MetadataAccessor that can retrieve
// or manipulate resource version on objects derived from core API
// metadata concepts.
func NewAccessor() MetadataAccessor {
	return resourceAccessor{}
}

// resourceAccessor implements ResourceVersioner and SelfLinker.
type resourceAccessor struct{}

func (resourceAccessor) Kind(obj runtime.Object) (string, error) {
	return objectAccessor{obj}.GetKind(), nil
}

func (resourceAccessor) SetKind(obj runtime.Object, kind string) error {
	objectAccessor{obj}.SetKind(kind)
	return nil
}

func (resourceAccessor) APIVersion(obj runtime.Object) (string, error) {
	return objectAccessor{obj}.GetAPIVersion(), nil
}

func (resourceAccessor) SetAPIVersion(obj runtime.Object, version string) error {
	objectAccessor{obj}.SetAPIVersion(version)
	return nil
}

func (resourceAccessor) Namespace(obj runtime.Object) (string, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return "", err
	}
	return accessor.GetNamespace(), nil
}

func (resourceAccessor) SetNamespace(obj runtime.Object, namespace string) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetNamespace(namespace)
	return nil
}

func (resourceAccessor) Name(obj runtime.Object) (string, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return "", err
	}
	return accessor.GetName(), nil
}

func (resourceAccessor) SetName(obj runtime.Object, name string) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetName(name)
	return nil
}

func (resourceAccessor) GenerateName(obj runtime.Object) (string, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return "", err
	}
	return accessor.GetGenerateName(), nil
}

func (resourceAccessor) SetGenerateName(obj runtime.Object, name string) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetGenerateName(name)
	return nil
}

func (resourceAccessor) UID(obj runtime.Object) (types.UID, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return "", err
	}
	return accessor.GetUID(), nil
}

func (resourceAccessor) SetUID(obj runtime.Object, uid types.UID) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetUID(uid)
	return nil
}

func (resourceAccessor) SelfLink(obj runtime.Object) (string, error) {
	accessor, err := ListAccessor(obj)
	if err != nil {
		return "", err
	}
	return accessor.GetSelfLink(), nil
}

func (resourceAccessor) SetSelfLink(obj runtime.Object, selfLink string) error {
	accessor, err := ListAccessor(obj)
	if err != nil {
		return err
	}
	accessor.SetSelfLink(selfLink)
	return nil
}

func (resourceAccessor) Labels(obj runtime.Object) (map[string]string, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return nil, err
	}
	return accessor.GetLabels(), nil
}

func (resourceAccessor) SetLabels(obj runtime.Object, labels map[string]string) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetLabels(labels)
	return nil
}

func (resourceAccessor) Annotations(obj runtime.Object) (map[string]string, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return nil, err
	}
	return accessor.GetAnnotations(), nil
}

func (resourceAccessor) SetAnnotations(obj runtime.Object, annotations map[string]string) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetAnnotations(annotations)
	return nil
}

func (resourceAccessor) ResourceVersion(obj runtime.Object) (string, error) {
	accessor, err := ListAccessor(obj)
	if err != nil {
		return "", err
	}
	return accessor.GetResourceVersion(), nil
}

func (resourceAccessor) SetResourceVersion(obj runtime.Object, version string) error {
	accessor, err := ListAccessor(obj)
	if err != nil {
		return err
	}
	accessor.SetResourceVersion(version)
	return nil
}
