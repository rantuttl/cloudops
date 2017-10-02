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

package authorization

import (
	"fmt"
	"time"
	"errors"

	"github.com/rantuttl/cloudops/apiserver/pkg/server/authorization/authorizer"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/authorization/authorizer/authorizerfactory"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/authorization/request/union"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/authorization/modes"
)

type AuthorizationConfig struct {
	AuthorizationModes []string

	// Path to an ABAC policy file.
	PolicyFile string

	WebhookConfigFile string
	// TTL for caching of authorized responses from the webhook server.
	WebhookCacheAuthorizedTTL time.Duration
	// TTL for caching of unauthorized responses from the webhook server.
	WebhookCacheUnauthorizedTTL time.Duration
}

func (c AuthorizationConfig) New() (authorizer.Authorizer, error) {
	if len(c.AuthorizationModes) == 0 {
		return nil, errors.New("At least one authorization mode should be passed")
	}

	var authorizers []authorizer.Authorizer
	authorizerMap := make(map[string]bool)

	for _, authorizationMode := range c.AuthorizationModes {
		if authorizerMap[authorizationMode] {
			return nil, fmt.Errorf("Authorization mode %s specified more than once", authorizationMode)
		}

		switch authorizationMode {
		case modes.ModeAlwaysAllow:
			authorizers = append(authorizers, authorizerfactory.NewAlwaysAllowAuthorizer())
		case modes.ModeAlwaysDeny:
			authorizers = append(authorizers, authorizerfactory.NewAlwaysDenyAuthorizer())
		// TODO (rantuttl): Add RBAC and other modes
		default:
			return nil, fmt.Errorf("Unknown authorization mode %s specified", authorizationMode)
		}
		authorizerMap[authorizationMode] = true
	}

	return union.New(authorizers...), nil
}
