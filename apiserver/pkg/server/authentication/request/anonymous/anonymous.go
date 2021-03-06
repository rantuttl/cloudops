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

package anonymous

import (
	"net/http"

	"github.com/rantuttl/cloudops/apiserver/pkg/server/authentication/authenticator"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/authentication/user"
)

const (
	anonymousUser = user.Anonymous

	unauthenticatedGroup = user.AllUnauthenticated
)

func NewAuthenticator() authenticator.Request {
	return authenticator.RequestFunc(func(req *http.Request) (user.Info, bool, error) {
		return &user.DefaultInfo{Name: anonymousUser, Groups: []string{unauthenticatedGroup}}, true, nil
	})
}
