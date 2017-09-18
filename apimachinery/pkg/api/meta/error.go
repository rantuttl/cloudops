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

package meta

import (
	"fmt"

	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
)

// AmbiguousResourceError is returned if the RESTMapper finds multiple matches for a resource
type AmbiguousResourceError struct {
	PartialResource schema.GroupVersionResource
	MatchingResources []schema.GroupVersionResource
	MatchingKinds     []schema.GroupVersionKind
}

func (e *AmbiguousResourceError) Error() string {
	switch {
	case len(e.MatchingKinds) > 0 && len(e.MatchingResources) > 0:
		return fmt.Sprintf("%v matches multiple resources %v and kinds %v", e.PartialResource, e.MatchingResources, e.MatchingKinds)
	case len(e.MatchingKinds) > 0:
		return fmt.Sprintf("%v matches multiple kinds %v", e.PartialResource, e.MatchingKinds)
	case len(e.MatchingResources) > 0:
		return fmt.Sprintf("%v matches multiple resources %v", e.PartialResource, e.MatchingResources)
	}
	return fmt.Sprintf("%v matches multiple resources or kinds", e.PartialResource)
}

// NoKindMatchError is returned if the RESTMapper can't find any match for a kind
type NoKindMatchError struct {
	PartialKind schema.GroupVersionKind
}

func (e *NoKindMatchError) Error() string {
	return fmt.Sprintf("no matches for %v", e.PartialKind)
}

// NoResourceMatchError is returned if the RESTMapper can't find any match for a resource
type NoResourceMatchError struct {
	PartialResource schema.GroupVersionResource
}

func (e *NoResourceMatchError) Error() string {
	return fmt.Sprintf("no matches for %v", e.PartialResource)
}
