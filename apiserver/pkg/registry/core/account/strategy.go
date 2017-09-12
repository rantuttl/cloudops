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

package account

import (
	"github.com/rantuttl/cloudops/apiserver/pkg/api"
	"github.com/rantuttl/cloudops/apiserver/pkg/storage/names"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apiserver/pkg/api/validation"
	genericapirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
	"github.com/rantuttl/cloudops/apimachinery/pkg/util/validation/field"
)

// FIXME (rantuttl): Delete, but for reference, see: pkg/registry/core/namespace/strategy.go

// accountStrategy implements behavior for Accounts
type accountStrategy struct {
        runtime.ObjectTyper
        names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Namespace
// objects via the REST API.
var Strategy = accountStrategy{api.Scheme, names.SimpleNameGenerator}

// NamespaceScoped is false for accounts.
func (accountStrategy) NamespaceScoped() bool {
        return false
}

func (accountStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	account := obj.(*api.Account)
	account.Status = api.AccountStatus{
		Phase: api.AccountActive,
	}
}

func (accountStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	account := obj.(*api.Account)
	return validation.ValidateAccount(account)
}

func (accountStrategy) Canonicalize(obj runtime.Object) {
}

// TODO (rantuttl): Place Update strategy functions here...

// FIXME (rantuttl): Unstub these methods
func (accountStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
}
