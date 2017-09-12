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
	"github.com/rantuttl/cloudops/apiserver/pkg/api"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/core/account"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/generic"
	genericapirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
	genericregistry "github.com/rantuttl/cloudops/apiserver/pkg/registry/generic/registry"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
)

type REST struct {
	store *genericregistry.Store
}

func NewREST(optsGetter generic.RESTOptionsGetter) *REST {
	// FIXME (rantuttl): Partially stubbed for now...
	store := &genericregistry.Store{
		NewFunc:		func() runtime.Object { return &api.Account{} },
		NewListFunc:		func() runtime.Object { return &api.AccountList{} },
		QualifiedResource:	api.Resource("accounts"),
		CreateStrategy:		account.Strategy,
		UpdateStrategy:		account.Strategy,
		DeleteStrategy:		account.Strategy,
	}
	// FIXME (rantuttl)
	//options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: account.GetAttrs}
	options := &generic.StoreOptions{RESTOptions: optsGetter}
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

func (r *REST) Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}
