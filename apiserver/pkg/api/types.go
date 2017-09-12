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

// keep in-sync with apiserver/pkg/apigroups/core/types.go

package api

import (
	v1meta "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
)

type AccountStatus struct {
	Phase AccountPhase
}
type AccountPhase string

const (
	AccountActive AccountPhase = "Active"
	AccountTerminating AccountPhase = "Terminating"
	AccountInactive AccountPhase = "Inactive"
)

type Account struct {
	v1meta.TypeMeta
	v1meta.ObjectMeta
	Spec AccountSpec
	Status AccountStatus
}

type AccountSpec struct {
	// TODO (rantuttl): add additional info as needed about account specifics
}

type AccountList struct {
	v1meta.TypeMeta
	v1meta.ListMeta
	Items []Account
}
