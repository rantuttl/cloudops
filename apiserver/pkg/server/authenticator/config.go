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

package authenticator

import (
	"github.com/go-openapi/spec"

	"github.com/rantuttl/cloudops/apiserver/pkg/authentication/authenticator"
	"github.com/rantuttl/cloudops/apiserver/pkg/authentication/group"
	"github.com/rantuttl/cloudops/apiserver/pkg/authentication/request/union"
	"github.com/rantuttl/cloudops/apiserver/pkg/authentication/request/anonymous"
	"github.com/rantuttl/cloudops/apiserver/pkg/authentication/request/x509"
	"github.com/rantuttl/cloudops/apiserver/plugin/pkg/authenticator/password/keystone"
	"github.com/rantuttl/cloudops/apiserver/plugin/pkg/authenticator/request/basicauth"
	certutil "github.com/rantuttl/cloudops/apiserver/pkg/util/cert"
)

type AuthenticatorConfig struct {
	// FIXME (rantuttle): Added support for Bearer Token in request header
	Anonymous                   bool
	ClientCAFile                string
	KeystoneURL                 string
	KeystoneCAFile              string
}

// New returns an authenticator.Request or an error that supports the standard
// product authentication mechanisms.
func (config AuthenticatorConfig) New() (authenticator.Request, *spec.SecurityDefinitions, error) {
	var authenticators []authenticator.Request

	securityDefinitions := spec.SecurityDefinitions{}
	hasBasicAuth := false
	hasTokenAuth := false

	if len(config.KeystoneURL) > 0 {
		keystoneAuth, err := newAuthenticatorFromKeystoneURL(config.KeystoneURL, config.KeystoneCAFile)
		if err != nil {
			return nil, nil, err
		}
		authenticators = append(authenticators, keystoneAuth)
		hasBasicAuth = true
	}

	// FIXME (rantuttle): Added support for Bearer Token as X-Auth-Token: request header -- hasTokenAuth = true


	// X509 methods
	if len(config.ClientCAFile) > 0 {
		certAuth, err := newAuthenticatorFromClientCAFile(config.ClientCAFile)
		if err != nil {
			return nil, nil, err
		}
		authenticators = append(authenticators, certAuth)
	}

	if hasBasicAuth {
		securityDefinitions["HTTPBasic"] = &spec.SecurityScheme{
			SecuritySchemeProps: spec.SecuritySchemeProps{
				Type:        "basic",
				Description: "HTTP Basic authentication",
			},
		}
	}

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

	if len(authenticators) == 0 {
		if config.Anonymous {
			return anonymous.NewAuthenticator(), &securityDefinitions, nil
		}
	}

	switch len(authenticators) {
	case 0:
		return nil, &securityDefinitions, nil
	}

	authenticator := union.New(authenticators...)

	authenticator = group.NewAuthenticatedGroupAdder(authenticator)

	if config.Anonymous {
		// If the authenticator chain returns an error, return an error (don't consider a bad bearer token
		// or invalid username/password combination anonymous).
		authenticator = union.NewFailOnError(authenticator, anonymous.NewAuthenticator())
	}

	return authenticator, &securityDefinitions, nil
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

// newAuthenticatorFromKeystoneURL returns an authenticator.Request or an error
func newAuthenticatorFromKeystoneURL(keystoneURL string, keystoneCAFile string) (authenticator.Request, error) {
	keystoneAuthenticator, err := keystone.NewKeystoneAuthenticator(keystoneURL, keystoneCAFile)
	if err != nil {
		return nil, err
	}

	return basicauth.New(keystoneAuthenticator), nil
}
