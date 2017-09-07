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

package options

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	genericserver "github.com/rantuttl/cloudops/apiserver/pkg/genericserver/server"
	genericopts "github.com/rantuttl/cloudops/apiserver/pkg/genericserver/options"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/authenticator"
	authzmodes "github.com/rantuttl/cloudops/apiserver/pkg/server/authorizer/modes"
)

type BuiltInAuthenticationOptions struct {
	// FIXME (rantuttl): Added support for Bearer Token in request header
	Anonymous       *AnonymousAuthenticationOptions
	ClientCert      *genericopts.ClientCertAuthenticationOptions
	Keystone        *KeystoneAuthenticationOptions
}

type AnonymousAuthenticationOptions struct {
	Allow bool
}

type KeystoneAuthenticationOptions struct {
	URL    string
	CAFile string
}

func NewBuiltInAuthenticationOptions() *BuiltInAuthenticationOptions {
	return &BuiltInAuthenticationOptions{}
}

func (s *BuiltInAuthenticationOptions) WithAll() *BuiltInAuthenticationOptions {
	return s.
		WithAnyonymous().
		WithClientCert().
		WithKeystone()
}

func (s *BuiltInAuthenticationOptions) WithAnyonymous() *BuiltInAuthenticationOptions {
	s.Anonymous = &AnonymousAuthenticationOptions{Allow: true}
	return s
}

func (s *BuiltInAuthenticationOptions) WithClientCert() *BuiltInAuthenticationOptions {
	s.ClientCert = &genericopts.ClientCertAuthenticationOptions{}
	return s
}

func (s *BuiltInAuthenticationOptions) WithKeystone() *BuiltInAuthenticationOptions {
	s.Keystone = &KeystoneAuthenticationOptions{}
	return s
}

// Validate checks invalid config combination
func (s *BuiltInAuthenticationOptions) Validate() []error {
	allErrors := []error{}

	return allErrors
}

func (s *BuiltInAuthenticationOptions) AddFlags(fs *pflag.FlagSet) {
	if s.Anonymous != nil {
		fs.BoolVar(&s.Anonymous.Allow, "anonymous-auth", s.Anonymous.Allow, ""+
			"Enables anonymous requests to the secure port of the API server. "+
			"Requests that are not rejected by another authentication method are treated as anonymous requests. "+
			"Anonymous requests have a username of system:anonymous, and a group name of system:unauthenticated.")
	}

	if s.ClientCert != nil {
		s.ClientCert.AddFlags(fs)
	}

	if s.Keystone != nil {
		fs.StringVar(&s.Keystone.URL, "keystone-url", s.Keystone.URL,
			"If passed, activates the keystone authentication plugin.")

		fs.StringVar(&s.Keystone.CAFile, "keystone-ca-file", s.Keystone.CAFile, ""+
			"If set, the Keystone server's certificate will be verified by one of the authorities "+
			"in the keystone-ca-file, otherwise the host's root CA set will be used.")
	}
}

func (s *BuiltInAuthenticationOptions) AddDeprecatedFlags(fs *pflag.FlagSet) {
        // place deprecated flags here. For example:
        //fs.IPVar(&s.BindAddress, "address", s.BindAddress,
        //      "DEPRECATED: see --insecure-bind-address instead.")
        //fs.MarkDeprecated("address", "see --insecure-bind-address instead.")

        //fs.IntVar(&s.BindPort, "port", s.BindPort, "DEPRECATED: see --insecure-port instead.")
        //fs.MarkDeprecated("port", "see --insecure-port instead.")
}

func (s *BuiltInAuthenticationOptions) ToAuthenticationConfig() authenticator.AuthenticatorConfig {
	ret := authenticator.AuthenticatorConfig{}

	if s.Anonymous != nil {
		ret.Anonymous = s.Anonymous.Allow
	}

	if s.ClientCert != nil {
		ret.ClientCAFile = s.ClientCert.ClientCA
	}

	if s.Keystone != nil {
		ret.KeystoneURL = s.Keystone.URL
		ret.KeystoneCAFile = s.Keystone.CAFile
	}

	return ret
}

func (o *BuiltInAuthenticationOptions) ApplyTo(c *genericserver.Config) error {
	if o == nil {
		return nil
	}

	var err error
	if o.ClientCert != nil {
		c, err = c.ApplyClientCert(o.ClientCert.ClientCA)
		if err != nil {
			return fmt.Errorf("unable to load client CA file: %v", err)
		}
	}

	return nil
}

// ApplyAuthorization will conditionally modify the authentication options based on the authorization options
func (o *BuiltInAuthenticationOptions) ApplyAuthorization(authorization *BuiltInAuthorizationOptions) {
	if o == nil || authorization == nil || o.Anonymous == nil {
		return
	}

	// authorization ModeAlwaysAllow cannot be combined with AnonymousAuth.
	// in such a case the AnonymousAuth is stomped to false and you get a message
	if o.Anonymous.Allow {
		found := false
		for _, mode := range strings.Split(authorization.Mode, ",") {
			if mode == authzmodes.ModeAlwaysAllow {
				found = true
				break
			}
		}
		if found {
			glog.Warningf("AnonymousAuth is not allowed with the AllowAll authorizer.  Resetting AnonymousAuth to false. You should use a different authorizer")
			o.Anonymous.Allow = false
		}
	}
}
