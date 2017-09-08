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

package registered

import (
	"fmt"
	"strings"

	"github.com/golang/glog"

	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	"github.com/rantuttl/cloudops/apimachinery/pkg/apimachinery"
)

// APIRegistrationManager provides the concept of what API groups are enabled.
type APIRegistrationManager struct {
	// registeredGroupVersions stores all API group versions for which RegisterGroup is called.
	registeredVersions map[schema.GroupVersion]struct{}

	// enabledVersions represents all enabled API versions. It should be a
	// subset of registeredVersions. Please call EnableVersions() to add
	// enabled versions.
	enabledVersions map[schema.GroupVersion]struct{}

	// map of group meta for all groups.
	groupMetaMap map[string]*apimachinery.GroupMeta

	// envRequestedVersions represents the versions requested via the
	// API_VERSIONS environment variable. The install of each API group
	// should check this enviorment variable before adding their versions
	// to the latest package and Scheme.
	envRequestedVersions []schema.GroupVersion

}

func NewAPIRegistrationManager(APIVersions string) (*APIRegistrationManager, error) {
	m := &APIRegistrationManager{
		registeredVersions:	map[schema.GroupVersion]struct{}{},
		enabledVersions:	map[schema.GroupVersion]struct{}{},
		groupMetaMap:		map[string]*apimachinery.GroupMeta{},
		envRequestedVersions:	[]schema.GroupVersion{},
	}

	if len(APIVersions) != 0 {
		for _, version := range strings.Split(APIVersions, ",") {
			gv, err := schema.ParseGroupVersion(version)
			if err != nil {
				return nil, fmt.Errorf("invalid api version: %s in API_VERSIONS: %s.", version, APIVersions)
			}
			m.envRequestedVersions = append(m.envRequestedVersions, gv)
		}
	}
	return m, nil
}

func NewOrDie(APIVersions string) *APIRegistrationManager {
	m, err := NewAPIRegistrationManager(APIVersions)
	if err != nil {
		glog.Fatalf("Could not construct version manager: %v (API_VERSIONS=%q)", err, APIVersions)
	}
	return m
}

func (m *APIRegistrationManager) GroupOrDie(group string) *apimachinery.GroupMeta {
	groupMeta, found := m.groupMetaMap[group]
	if !found {
		if group == "" {
			panic("The legacy v1 API is not registered.")
		} else {
			panic(fmt.Sprintf("Group %s is not registered.", group))
		}
	}
	groupMetaCopy := *groupMeta
	return &groupMetaCopy
}
