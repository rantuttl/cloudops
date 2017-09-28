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
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/conversion"
	"github.com/rantuttl/cloudops/apiserver/pkg/apigroups/core"
	"github.com/rantuttl/cloudops/apiserver/pkg/api/core/v1"
)

// function that convert between external representations to internal representation and back again.
// Internal representation are in apiserver/pkg/apigroups/<group>,
// External representations are in apiserver/pkg/api/<group>
func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

func RegisterConversions(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedConversionFuncs(
		Convert_v1_Account_To_core_Account,
		Convert_v1_AccountSpec_To_core_AccountSpec,
		Convert_v1_AccountStatus_To_core_AccountStatus,
		Convert_core_Account_To_v1_Account,
		Convert_core_AccountSpec_To_v1_AccountSpec,
		Convert_core_AccountStatus_To_v1_AccountStatus,
	)
}

// TODO (rantuttl): Need to unstub once AccountSpec defined
func Convert_v1_AccountSpec_To_core_AccountSpec(in *v1.AccountSpec, out *core.AccountSpec, s conversion.Scope) error {
	return nil
}

func Convert_v1_AccountStatus_To_core_AccountStatus(in *v1.AccountStatus, out *core.AccountStatus, s conversion.Scope) error {
	out.Phase = core.AccountPhase(in.Phase)
	return nil
}

func Convert_v1_Account_To_core_Account(in *v1.Account, out *core.Account, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1_AccountSpec_To_core_AccountSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1_AccountStatus_To_core_AccountStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// TODO (rantuttl): Need to unstub once AccountSpec defined
func Convert_core_AccountSpec_To_v1_AccountSpec(in *core.AccountSpec, out *v1.AccountSpec, s conversion.Scope) error {
	return nil
}

func Convert_core_AccountStatus_To_v1_AccountStatus(in *core.AccountStatus, out *v1.AccountStatus, s conversion.Scope) error {
	out.Phase = v1.AccountPhase(in.Phase)
	return nil
}

func Convert_core_Account_To_v1_Account(in *core.Account, out *v1.Account, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_core_AccountSpec_To_v1_AccountSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_core_AccountStatus_To_v1_AccountStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}
