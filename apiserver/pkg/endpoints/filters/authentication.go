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

package filters

import (
	"errors"
	"net/http"

	"github.com/golang/glog"

	apierrors "github.com/rantuttl/cloudops/apimachinery/pkg/api/errors"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/authentication/authenticator"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/handlers/responsewriters"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
)

func WithAuthentication(handler http.Handler, mapper request.RequestContextMapper, auth authenticator.Request, failed http.Handler) http.Handler {
	if auth == nil {
		glog.Warningf("Authentication is disabled")
		return handler
	}
	return request.WithRequestContext(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user, ok, err := auth.AuthenticateRequest(req)
			if err != nil || !ok {
				if err != nil {
					glog.Errorf("Unable to authenticate the request due to an error: %v", err)
				}
				if failed != nil {
					failed.ServeHTTP(w, req)
				}
				return
			}

			req.Header.Del("Authorization")

			if ctx, ok := mapper.Get(req); ok {
				mapper.Update(req, request.WithUser(ctx, user))
			}

			handler.ServeHTTP(w, req)
		}),
		mapper,
	)
}

// Unauthorized returns a handler for users that are not authenticated, typically used for 'sercured' servers.
// This returned handler is invoked from WithAuthentication
func Unauthorized(requestContextMapper request.RequestContextMapper, s runtime.NegotiatedSerializer, supportsBasicAuth bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if supportsBasicAuth {
			w.Header().Set("WWW-Authenticate", `Basic realm="cloudops-master"`)
		}
		ctx, ok := requestContextMapper.Get(req)
		if !ok {
			responsewriters.InternalError(w, req, errors.New("no context found for request"))
			return
		}
		requestInfo, found := request.RequestInfoFrom(ctx)
		if !found {
			responsewriters.InternalError(w, req, errors.New("no RequestInfo found in the context"))
			return
		}

		gv := schema.GroupVersion{Group: requestInfo.APIGroup, Version: requestInfo.APIVersion}
		responsewriters.ErrorNegotiated(ctx, apierrors.NewUnauthorized("Unauthorized"), s, gv, w, req)
	})
}
