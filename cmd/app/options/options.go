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

// TODO (rantuttl): Modify to support command line options we need
// Package options contains flags and options for initializing an apiserver
package options

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/rantuttl/cloudops/apiserver/pkg/api"
	genericopts "github.com/rantuttl/cloudops/apiserver/pkg/genericserver/options"
	serveropts "github.com/rantuttl/cloudops/apiserver/pkg/server/options"
	"github.com/rantuttl/cloudops/apiserver/pkg/backend"
	corev1 "github.com/rantuttl/cloudops/apiserver/pkg/apigroups/core/v1"
)

// ServerRunOptions runs an apiserver.
type ServerRunOptions struct {
	GenericServerRunOptions *genericopts.ServerRunOptions
	Backend			*serveropts.BackendOptions
	SecureServing           *genericopts.SecureServingOptions
	InsecureServing         *genericopts.InsecureServingOptions
	Authentication          *serveropts.BuiltInAuthenticationOptions
	Authorization		*serveropts.BuiltInAuthorizationOptions

	EnableLogsHandler         bool
	EventTTL                  time.Duration
	MaxConnectionBytesPerSec  int64
}

// NewServerRunOptions creates a new ServerRunOptions object with default parameters
func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{
		GenericServerRunOptions: genericopts.NewServerRunOptions(),
		Backend: serveropts.NewBackendOptions(backend.NewDefaultConfig(api.Scheme, api.Codecs.LegacyCodec(corev1.SchemeGroupVersion))),
		SecureServing: genericopts.NewSecureServingOptions(),
		InsecureServing: genericopts.NewInsecureServingOptions(),
		Authentication:  serveropts.NewBuiltInAuthenticationOptions().WithAll(),
		Authorization:	serveropts.NewBuiltInAuthorizationOptions(),
		EnableLogsHandler: true,
		EventTTL:          1 * time.Hour,
	}
	return &s
}

// AddFlags adds flags for a specific APIServer to the specified FlagSet
func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {
	// Add the generic flags.
	s.GenericServerRunOptions.AddUniversalFlags(fs)
	s.Backend.AddFlags(fs)
	s.SecureServing.AddFlags(fs)
	s.SecureServing.AddDeprecatedFlags(fs)
	s.InsecureServing.AddFlags(fs)
	s.InsecureServing.AddDeprecatedFlags(fs)
	s.Authentication.AddFlags(fs)
	s.Authentication.AddDeprecatedFlags(fs)
	s.Authorization.AddFlags(fs)
	s.Authorization.AddDeprecatedFlags(fs)

	// Note: the weird ""+ in below lines seems to be the only way to get gofmt to
	// arrange these text blocks sensibly. Grrr.

	fs.DurationVar(&s.EventTTL, "event-ttl", s.EventTTL,
		"Amount of time to retain events.")

	fs.BoolVar(&s.EnableLogsHandler, "enable-logs-handler", s.EnableLogsHandler,
		"If true, install a /logs handler for the apiserver logs.")

	/* FIXME (rantuttl): Consider using similar approach to CAL comms, although may be hardcoded.
	fs.StringVar(&s.SSHUser, "ssh-user", s.SSHUser,
		"If non-empty, use secure SSH proxy to the nodes, using this user name")

	fs.StringVar(&s.SSHKeyfile, "ssh-keyfile", s.SSHKeyfile,
		"If non-empty, use secure SSH proxy to the nodes, using this user keyfile")

	fs.Int64Var(&s.MaxConnectionBytesPerSec, "max-connection-bytes-per-sec", s.MaxConnectionBytesPerSec, ""+
		"If non-zero, throttle each user connection to this number of bytes/sec. "+
		"Currently only applies to long-running requests.")

	// Kubelet related flags:
	fs.BoolVar(&s.KubeletConfig.EnableHttps, "kubelet-https", s.KubeletConfig.EnableHttps,
		"Use https for kubelet connections.")

	fs.StringSliceVar(&s.KubeletConfig.PreferredAddressTypes, "kubelet-preferred-address-types", s.KubeletConfig.PreferredAddressTypes,
		"List of the preferred NodeAddressTypes to use for kubelet connections.")

	fs.UintVar(&s.KubeletConfig.Port, "kubelet-port", s.KubeletConfig.Port,
		"DEPRECATED: kubelet port.")
	fs.MarkDeprecated("kubelet-port", "kubelet-port is deprecated and will be removed.")

	fs.UintVar(&s.KubeletConfig.ReadOnlyPort, "kubelet-read-only-port", s.KubeletConfig.ReadOnlyPort,
		"DEPRECATED: kubelet port.")

	fs.DurationVar(&s.KubeletConfig.HTTPTimeout, "kubelet-timeout", s.KubeletConfig.HTTPTimeout,
		"Timeout for kubelet operations.")

	fs.StringVar(&s.KubeletConfig.CertFile, "kubelet-client-certificate", s.KubeletConfig.CertFile,
		"Path to a client cert file for TLS.")

	fs.StringVar(&s.KubeletConfig.KeyFile, "kubelet-client-key", s.KubeletConfig.KeyFile,
		"Path to a client key file for TLS.")

	fs.StringVar(&s.KubeletConfig.CAFile, "kubelet-certificate-authority", s.KubeletConfig.CAFile,
		"Path to a cert file for the certificate authority.")

	fs.StringVar(&s.ProxyClientCertFile, "proxy-client-cert-file", s.ProxyClientCertFile, ""+
		"Client certificate used to prove the identity of the aggregator or kube-apiserver "+
		"when it must call out during a request. This includes proxying requests to a user "+
		"api-server and calling out to webhook admission plugins. It is expected that this "+
		"cert includes a signature from the CA in the --requestheader-client-ca-file flag. "+
		"That CA is published in the 'extension-apiserver-authentication' configmap in "+
		"the kube-system namespace. Components recieving calls from kube-aggregator should "+
		"use that CA to perform their half of the mutual TLS verification.")
	fs.StringVar(&s.ProxyClientKeyFile, "proxy-client-key-file", s.ProxyClientKeyFile, ""+
		"Private key for the client certificate used to prove the identity of the aggregator or kube-apiserver "+
		"when it must call out during a request. This includes proxying requests to a user "+
		"api-server and calling out to webhook admission plugins.")

	fs.BoolVar(&s.EnableAggregatorRouting, "enable-aggregator-routing", s.EnableAggregatorRouting,
		"Turns on aggregator routing requests to endoints IP rather than cluster IP.")
	*/

}
