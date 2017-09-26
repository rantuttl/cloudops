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

package core

import (
	"reflect"

	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/conversion"
)

func init() {
	SchemeBuilder.Register(RegisterDeepCopies)
}

func RegisterDeepCopies(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedDeepCopyFuncs(
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_core_Account, InType: reflect.TypeOf(&Account{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_core_AccountList, InType: reflect.TypeOf(&AccountList{})},
	)
}

func DeepCopy_core_Account(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*Account)
		out := out.(*Account)
		*out = *in
		if newVal, err := c.DeepCopy(&in.ObjectMeta); err != nil {
			return err
		} else {
			out.ObjectMeta = *newVal.(*metav1.ObjectMeta)
		}
		// TODO (rantuttl): Add Account Spec DeepCopy
		return nil
	}
}

func DeepCopy_core_AccountList(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*AccountList)
		out := out.(*AccountList)
		*out = *in
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			*out = make([]Account, len(*in))
			for i := range *in {
				if err := DeepCopy_core_Account(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		}
		return nil
	}
}
