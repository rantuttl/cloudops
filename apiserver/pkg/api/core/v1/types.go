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

package v1

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
	v1meta.TypeMeta `json:",inline"`
	v1meta.ObjectMeta `json:"metadata,omitempty"`
	Spec AccountSpec `json:"spec,omitempty"`
	Status AccountStatus `json:"status,omitempty"`
}

type AccountSpec struct {
	// TODO (rantuttl): add additional info as needed about account specifics
	// Can we use ObjectMeta's OwnerRefernce as the owner of the account, which can drive
	// associatoins in the data base
	// We need anther type which is User, and AccountSpec contains:
	//Users []User //Users in this account
}

type AccountList struct {
	v1meta.TypeMeta `json:",inline"`
	v1meta.ListMeta `json:"metadata,omitempty"`
	Items []Account `json:"items"`
}

// TODO (rantuttl): Add other "core" group types below
type User struct {
	v1meta.TypeMeta `json:",inline"`
	v1meta.ObjectMeta `json:"metadata,omitempty"`
	Spec UserSpec `json:"spec,omitempty"`
}

type UserSpec struct {
	// TODO (rantuttl): add additional info as needed about user specifics
	// First/Last name, username, email, password(salt:sha256-hash), account
	FirstName string `json:"firstname,omitempty"`
	LastName string `json:"lastname,omitempty"`
	Username string `json:"username"`
}

type UserList struct {
	v1meta.TypeMeta `json:",inline"`
	v1meta.ListMeta `json:"metadata,omitempty"`
	Items []User `json:"items"`
}
