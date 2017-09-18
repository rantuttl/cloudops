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

	"github.com/rantuttl/cloudops/apimachinery/pkg/api/meta"
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

// RegisterVersions adds the given group versions to the list of registered group versions.
func (m *APIRegistrationManager) RegisterVersions(availableVersions []schema.GroupVersion) {
	for _, v := range availableVersions {
		m.registeredVersions[v] = struct{}{}
	}
}

// RegisterGroup adds the given group to the list of registered groups.
func (m *APIRegistrationManager) RegisterGroup(groupMeta apimachinery.GroupMeta) error {
	groupName := groupMeta.GroupVersion.Group
	if _, found := m.groupMetaMap[groupName]; found {
		return fmt.Errorf("group %q is already registered in groupsMap: %v", groupName, m.groupMetaMap)
	}
	m.groupMetaMap[groupName] = &groupMeta
	return nil
}

// EnableVersions adds the versions for the given group to the list of enabled versions.
// Note that the caller should call RegisterGroup before calling this method.
// The caller of this function is responsible to add the versions to scheme and RESTMapper.
func (m *APIRegistrationManager) EnableVersions(versions ...schema.GroupVersion) error {
	var unregisteredVersions []schema.GroupVersion
	for _, v := range versions {
		if _, found := m.registeredVersions[v]; !found {
			unregisteredVersions = append(unregisteredVersions, v)
		}
		m.enabledVersions[v] = struct{}{}
	}
	if len(unregisteredVersions) != 0 {
		return fmt.Errorf("Please register versions before enabling them: %v", unregisteredVersions)
	}
	return nil
}

// IsAllowedVersion returns if the version is allowed by the API_VERSIONS
// environment variable. If the environment variable is empty, then it always
// returns true.
func (m *APIRegistrationManager) IsAllowedVersion(v schema.GroupVersion) bool {
	if len(m.envRequestedVersions) == 0 {
		return true
	}
	for _, envGV := range m.envRequestedVersions {
		if v == envGV {
			return true
		}
	}
	return false
}

// IsEnabledVersion returns if a version is enabled.
func (m *APIRegistrationManager) IsEnabledVersion(v schema.GroupVersion) bool {
	_, found := m.enabledVersions[v]
	return found
}

// EnabledVersions returns all enabled versions.  Groups are randomly ordered, but versions within groups
// are priority order from best to worst
func (m *APIRegistrationManager) EnabledVersions() []schema.GroupVersion {
	ret := []schema.GroupVersion{}
	for _, groupMeta := range m.groupMetaMap {
		for _, version := range groupMeta.GroupVersions {
			if m.IsEnabledVersion(version) {
				ret = append(ret, version)
			}
		}
	}
	return ret
}

// EnabledVersionsForGroup returns all enabled versions for a group in order of best to worst
func (m *APIRegistrationManager) EnabledVersionsForGroup(group string) []schema.GroupVersion {
	groupMeta, ok := m.groupMetaMap[group]
	if !ok {
		return []schema.GroupVersion{}
	}

	ret := []schema.GroupVersion{}
	for _, version := range groupMeta.GroupVersions {
		if m.IsEnabledVersion(version) {
			ret = append(ret, version)
		}
	}
	return ret
}

// Group returns the metadata of a group if the group is registered, otherwise
// an error is returned.
func (m *APIRegistrationManager) Group(group string) (*apimachinery.GroupMeta, error) {
	groupMeta, found := m.groupMetaMap[group]
	if !found {
		return nil, fmt.Errorf("group %v has not been registered", group)
	}
	groupMetaCopy := *groupMeta
	return &groupMetaCopy, nil
}

// IsRegistered takes a string and determines if it's one of the registered groups
func (m *APIRegistrationManager) IsRegistered(group string) bool {
	_, found := m.groupMetaMap[group]
	return found
}

// IsRegisteredVersion returns if a version is registered.
func (m *APIRegistrationManager) IsRegisteredVersion(v schema.GroupVersion) bool {
	_, found := m.registeredVersions[v]
	return found
}

// RegisteredGroupVersions returns all registered group versions.
func (m *APIRegistrationManager) RegisteredGroupVersions() []schema.GroupVersion {
	ret := []schema.GroupVersion{}
	for groupVersion := range m.registeredVersions {
		ret = append(ret, groupVersion)
	}
	return ret
}

// InterfacesFor is a union meta.VersionInterfacesFunc func for all registered types
func (m *APIRegistrationManager) InterfacesFor(version schema.GroupVersion) (*meta.VersionInterfaces, error) {
	groupMeta, err := m.Group(version.Group)
	if err != nil {
		return nil, err
	}
	return groupMeta.InterfacesFor(version)
}

func (m *APIRegistrationManager) GroupOrDie(group string) *apimachinery.GroupMeta {
	groupMeta, found := m.groupMetaMap[group]
	if !found {
		panic(fmt.Sprintf("Group %s is not registered.", group))
	}
	groupMetaCopy := *groupMeta
	return &groupMetaCopy
}
