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

package apimachinery

import (
	"github.com/rantuttl/cloudops/apimachinery/pkg/api/meta"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
)

// GroupMeta stores the metadata of a group.
type GroupMeta struct {
	// GroupVersion represents the preferred version of the group.
	GroupVersion schema.GroupVersion

	// GroupVersions is Group + all versions in that group.
	GroupVersions []schema.GroupVersion

	// RESTMapper provides the default mapping between REST paths and the objects declared in api.Scheme and all known
	// versions.
	RESTMapper meta.RESTMapper
}
