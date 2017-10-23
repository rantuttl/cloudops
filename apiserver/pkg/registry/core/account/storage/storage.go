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

package storage

import (
	"fmt"

	"github.com/rantuttl/cloudops/apiserver/pkg/apigroups/core"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/core/account"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/generic"
	genericapirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
	genericregistry "github.com/rantuttl/cloudops/apiserver/pkg/registry/generic/registry"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	apierrors "github.com/rantuttl/cloudops/apimachinery/pkg/api/errors"
)

type REST struct {
	store *genericregistry.Store
}

func NewREST(optsGetter generic.RESTOptionsGetter) *REST {
	store := &genericregistry.Store{
		NewFunc:		func() runtime.Object { return &core.Account{} },
		NewListFunc:		func() runtime.Object { return &core.AccountList{} },
		QualifiedResource:	core.Resource("accounts"),
		CreateStrategy:		account.Strategy,
		UpdateStrategy:		account.Strategy,
		DeleteStrategy:		account.Strategy,
		ReturnDeletedObject:	true,
	}
	// FIXME (rantuttl)
	//options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: account.GetAttrs}
	options := &generic.StoreOptions{
		RESTOptions: optsGetter,
		Transformer: &accountTransformer{resource: "accounts"},
	}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err)
	}
	return &REST{store}
}

func (r *REST) New() runtime.Object {
	return r.store.New() // Calls the above NewFunc
}

func (r *REST) NewList() runtime.Object {
	return r.store.NewList() // Calls the above NewListFunc
}

// TODO (rantuttl): Add other methods List, Update, Delete, etc...

func (r *REST) Create(ctx genericapirequest.Context, obj runtime.Object, includeUninitialized bool) (runtime.Object, error) {
	return r.store.Create(ctx, obj, includeUninitialized)
}

// TODO (rantuttl): Switch to GetterWithOptions interface support. This will allow us to add query parms to
// the resource. The 'Get' signature has a different type for options. Will also need to implement 'NewGetOptions'
// method to support the installer. It will return the type (e.g., AccountOptions) to support the query parms.
func (r *REST) Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

func (r *REST) Delete(ctx genericapirequest.Context, name string, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	obj, err := r.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}

	account := obj.(*core.Account)

	// make sure we have the object's UID
	if options == nil {
		options = metav1.NewDeleteOptions(0)
	}

	if options.Preconditions == nil {
		options.Preconditions = &metav1.Preconditions{}
	}

	if options.Preconditions.UID == nil {
		options.Preconditions.UID = &account.UID
	} else if *options.Preconditions.UID != account.UID {
		apierrors.NewConflict(
			core.Resource("accounts"),
			name,
			fmt.Errorf("Precondition failed: UID in precondition: %v, UID in object meta: %v", *options.Preconditions.UID, account.UID),
		)
	}

	return r.store.Delete(ctx, name, options)
}
