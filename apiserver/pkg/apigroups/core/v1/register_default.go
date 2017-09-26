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
	v1core "github.com/rantuttl/cloudops/apiserver/pkg/api/core/v1"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
)

// Register each type's default settings
func RegisterDefaults(scheme *runtime.Scheme) error {
	scheme.AddTypeDefaultingFunc(&v1core.Account{}, func(obj interface{}) {SetObjectDefaults_Account(obj.(*v1core.Account)) })
	scheme.AddTypeDefaultingFunc(&v1core.AccountList{}, func(obj interface{}) {SetObjectDefaults_AccountList(obj.(*v1core.AccountList)) })
	return nil
}

func SetObjectDefaults_Account(in *v1core.Account) {
	SetDefaults_AccountStatus(&in.Status)
}

func SetObjectDefaults_AccountList(in *v1core.AccountList) {
	for i := range in.Items {
		a := &in.Items[i]
		SetObjectDefaults_Account(a)
	}
}
