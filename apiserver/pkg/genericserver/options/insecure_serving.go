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

// Package options contains flags and options for initializing an apiserver
package options

import (
	"fmt"
	"net"
	"strconv"

	//"github.com/pborman/uuid"
	"github.com/spf13/pflag"

	utilnet "github.com/rantuttl/cloudops/apimachinery/pkg/util/net"
	genericserver "github.com/rantuttl/cloudops/apiserver/pkg/genericserver/server"
	//serveropts "github.com/rantuttl/cloudops/apiserver/pkg/genericserver/options"
	//apiserver "github.com/rantuttl/cloudops/apiserver/pkg/server"
)

//TODO (rantuttl): Delete as it's not referenced from here, and is also defined in secure_serving.go
/*
type SecureServingOptions struct {
        BindAddress net.IP
        BindPort    int

        // ServerCert is the TLS cert info for serving secure traffic
        ServerCert GeneratableKeyCert
}

type CertKey struct {
        // CertFile is a file containing a PEM-encoded certificate, and possibly the complete certificate chain
        CertFile string
        // KeyFile is a file containing a PEM-encoded private key for the certificate specified by CertFile
        KeyFile string
}

type GeneratableKeyCert struct {
        CertKey CertKey

        // CACertFile is an optional file containing the certificate chain for CertKey.CertFile
        CACertFile string
        // CertDirectory is a directory that will contain the certificates.  If the cert and key aren't specifically set
        // this will be used to derive a match with the "pair-name"
        CertDirectory string
        // PairName is the name which will be used with CertDirectory to make a cert and key names
        // It becomes CertDirector/PairName.crt and CertDirector/PairName.key
        PairName string
}


// NewSecureServingOptions gives default values for the apiserver which are not the options wanted by
// "normal" API servers running on the platform
func NewSecureServingOptions() *serveropts.SecureServingOptions {
	return &serveropts.SecureServingOptions{
		BindAddress: net.ParseIP("0.0.0.0"),
		BindPort:    6443,
		ServerCert: serveropts.GeneratableKeyCert{
			PairName:      "apiserver",
			CertDirectory: "/var/run/cloudtops",
		},
	}
}
*/

// DefaultAdvertiseAddress sets the field AdvertiseAddress if
// unset. The field will be set based on the SecureServingOptions. If
// the SecureServingOptions is not present, DefaultExternalAddress
// will fall back to the insecure ServingOptions.
func DefaultAdvertiseAddress(s *ServerRunOptions, insecure *InsecureServingOptions) error {
	if insecure == nil {
		return nil
	}

	if s.AdvertiseAddress == nil || s.AdvertiseAddress.IsUnspecified() {
		hostIP, err := insecure.DefaultExternalAddress()
		if err != nil {
			return fmt.Errorf("Unable to find suitable network address.error='%v'. "+
				"Try to set the AdvertiseAddress directly or provide a valid BindAddress to fix this.", err)
		}
		s.AdvertiseAddress = hostIP
	}

	return nil
}

// InsecureServingOptions are for creating an unauthenticated, unauthorized, insecure port.
// No one should be using these anymore.
type InsecureServingOptions struct {
	BindAddress net.IP
	BindPort    int
}

// NewInsecureServingOptions is for creating an unauthenticated, unauthorized, insecure port.
// No one should be using these anymore.
func NewInsecureServingOptions() *InsecureServingOptions {
	return &InsecureServingOptions{
		BindAddress: net.ParseIP("127.0.0.1"),
		BindPort:    8080,
	}
}

func (s InsecureServingOptions) Validate(portArg string) []error {
	errors := []error{}

	if s.BindPort < 0 || s.BindPort > 65535 {
		errors = append(errors, fmt.Errorf("--insecure-port %v must be between 0 and 65535, inclusive. 0 for turning off secure port.", s.BindPort))
	}

	return errors
}

func (s *InsecureServingOptions) DefaultExternalAddress() (net.IP, error) {
	return utilnet.ChooseBindAddress(s.BindAddress)
}

func (s *InsecureServingOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IPVar(&s.BindAddress, "insecure-bind-address", s.BindAddress, ""+
		"The IP address on which to serve the --insecure-port (set to 0.0.0.0 for all interfaces).")

	fs.IntVar(&s.BindPort, "insecure-port", s.BindPort, ""+
		"The port on which to serve unsecured, unauthenticated access. It is assumed "+
		"that firewall rules are set up such that this port is not reachable from outside of "+
		"the cluster and that port 443 on the cluster's public address is proxied to this "+
		"port. This is performed by nginx in the default setup.")
}

func (s *InsecureServingOptions) AddDeprecatedFlags(fs *pflag.FlagSet) {
	// place deprecated flags here. For example:
	//fs.IPVar(&s.BindAddress, "address", s.BindAddress,
	//	"DEPRECATED: see --insecure-bind-address instead.")
	//fs.MarkDeprecated("address", "see --insecure-bind-address instead.")

	//fs.IntVar(&s.BindPort, "port", s.BindPort, "DEPRECATED: see --insecure-port instead.")
	//fs.MarkDeprecated("port", "see --insecure-port instead.")
}

func (s *InsecureServingOptions) ApplyTo(c *genericserver.Config) (*genericserver.InsecureServingInfo, error) {
	if s.BindPort <= 0 {
		return nil, nil
	}

	ret := &genericserver.InsecureServingInfo{
		BindAddress: net.JoinHostPort(s.BindAddress.String(), strconv.Itoa(s.BindPort)),
	}

	// FIXME (rantuttl): Investigate if we need this. Comment for now
	//var err error
	//privilegedLoopbackToken := uuid.NewRandom().String()
	//if c.LoopbackClientConfig, err = ret.NewLoopbackClientConfig(privilegedLoopbackToken); err != nil {
	//	return nil, err
	//}

	return ret, nil
}
