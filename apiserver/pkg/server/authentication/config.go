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

package authentication

import (
	"github.com/go-openapi/spec"

	"github.com/rantuttl/cloudops/apiserver/pkg/server/authentication/authenticator"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/authentication/request/x509"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/authentication/request/anonymous"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/authentication/request/union"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/authentication/group"
	"github.com/rantuttl/cloudops/apiserver/plugin/pkg/authenticator/password/keystone"
	"github.com/rantuttl/cloudops/apiserver/plugin/pkg/authenticator/password/passwordfile"
	"github.com/rantuttl/cloudops/apiserver/plugin/pkg/authenticator/request/basicauth"
	certutil "github.com/rantuttl/cloudops/apiserver/pkg/util/cert"
)

type AuthenticatorConfig struct {
	Anonymous			bool
	BasicAuthFile			string
	ClientCAFile			string
	KeystoneURL			string
	KeystoneCAFile			string
}

// creates a chain of authenticators
func (c AuthenticatorConfig) New() (authenticator.Request, *spec.SecurityDefinitions, error) {
	var authenticators  []authenticator.Request
	securityDefinitions := spec.SecurityDefinitions{}
	hasBasicAuth := false
	hasTokenAuth := false

	if len(c.BasicAuthFile) > 0 {
		basicAuth, err := newAuthenticatorFromBasicAuthFile(c.BasicAuthFile)
		if err != nil {
			return nil, nil, err
		}
		authenticators = append(authenticators, basicAuth)
		hasBasicAuth = true
	}

	if len(c.KeystoneURL) > 0 {
		// basic auth via stored password in keystone
		keystoneAuth, err := newAuthenticatorFromKeystoneURL(c.KeystoneURL, c.KeystoneCAFile)
		if err != nil {
			return nil, nil, err
		}
		authenticators = append(authenticators, keystoneAuth)
		hasBasicAuth = true
	}

	if len(c.ClientCAFile) > 0 {
		// basic auth via x509 certs
		certAuth, err := newAuthenticatorFromClientCAFile(c.ClientCAFile)
		if err != nil {
			return nil, nil, err
		}
		authenticators = append(authenticators, certAuth)
	}

	if hasBasicAuth {
		securityDefinitions["HTTPBasic"] = &spec.SecurityScheme{
			SecuritySchemeProps: spec.SecuritySchemeProps{
				Type:        "basic",
				Name:        "HTTP Basic authentication",
			},
		}
	}

	// TODO (rantuttl): Implement (above) Token support of some type
	if hasTokenAuth {
		securityDefinitions["BearerToken"] = &spec.SecurityScheme{
			SecuritySchemeProps: spec.SecuritySchemeProps{
				Type:        "apiKey",
				Name:        "authorization",
				In:          "header",
				Description: "Bearer Token authentication",
			},
		}
	}

	// if no authenticators found, then fallback to anonymous if so configured
	if len(authenticators) == 0 {
		if c.Anonymous {
			return anonymous.NewAuthenticator(), &securityDefinitions, nil
		}
		return nil, &securityDefinitions, nil
	}

	// this creates a wrapper that implements AuthenticateRequest, but loops through all authenticators
	authenticator := union.New(authenticators...)
	// implements AuthenticateRequest, and invokes the union's AuthenticateRequest method
	authenticator = group.NewAuthenticatedGroupAdder(authenticator)

	if c.Anonymous {
		// If the authenticator chain returns an error, return an error (don't consider a bad bearer token
		// or invalid username/password combination anonymous).
		// implements AuthenticateRequest, and invokes the AuthenticatedGroupAdder's AuthenticateRequest
		authenticator = union.NewFailOnError(authenticator, anonymous.NewAuthenticator())
	}

	return authenticator, &securityDefinitions, nil
}

// newAuthenticatorFromBasicAuthFile returns an authenticator.Request or an error
func newAuthenticatorFromBasicAuthFile(basicAuthFile string) (authenticator.Request, error) {
	basicAuthenticator, err := passwordfile.NewCSV(basicAuthFile)
	if err != nil {
		return nil, err
	}
	return basicauth.New(basicAuthenticator), nil
}

// newAuthenticatorFromKeystoneURL returns an authenticator.Request or an error
func newAuthenticatorFromKeystoneURL(keystoneURL string, keystoneCAFile string) (authenticator.Request, error) {
	keystoneAuthenticator, err := keystone.NewKeystoneAuthenticator(keystoneURL, keystoneCAFile)
	if err != nil {
		return nil, err
	}
	return basicauth.New(keystoneAuthenticator), nil
}

// newAuthenticatorFromClientCAFile returns an authenticator.Request or an error
func newAuthenticatorFromClientCAFile(clientCAFile string) (authenticator.Request, error) {
	roots, err := certutil.NewPool(clientCAFile)
	if err != nil {
		return nil, err
	}
	opts := x509.DefaultVerifyOptions()
	opts.Roots = roots
	return x509.New(opts, x509.CommonNameUserConversion), nil
}
