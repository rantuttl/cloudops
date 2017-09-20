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
package server

import (
	"fmt"
	"time"
	"net"
	"net/http"
	"crypto/tls"
	"crypto/x509"
	"strings"

	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/serializer"
	"github.com/rantuttl/cloudops/apimachinery/pkg/version"
	"github.com/rantuttl/cloudops/apimachinery/pkg/util/sets"
	"github.com/rantuttl/cloudops/apiserver/pkg/server/routes"
	//genericapiserver "github.com/rantuttl/cloudops/apiserver/pkg/genericserver/server"
	genericapifilters "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/filters"
	apirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
	certutil "github.com/rantuttl/cloudops/apiserver/pkg/util/cert"
	genericregistry "github.com/rantuttl/cloudops/apiserver/pkg/registry/generic"
)

const (
	APIGroupPrefix = "/api"
)

// Config is a structure used to configure a GenericAPIServer.
type Config struct {
	// Serializer is required and provides the interface for serializing and converting objects to and from the wire
	// The default (api.Codecs) usually works fine.
	Serializer runtime.NegotiatedSerializer
	SecureServingInfo *SecureServingInfo
	CorsAllowedOriginList []string
	BuildHandlerChainFunc func(apiHandler http.Handler, c *Config) (secure http.Handler)
	EnableSwaggerUI bool
        // RequestContextMapper maps requests to contexts. Exported so downstream consumers can provider their own mappers
        // TODO confirm that anyone downstream actually uses this and doesn't just need an accessor
        RequestContextMapper apirequest.RequestContextMapper
	// If specified, requests will be allocated a random timeout between this value, and twice this value.
	// Note that it is up to the request handlers to ignore or honor this timeout. In seconds.
	MinRequestTimeout int
	// MaxRequestsInFlight is the maximum number of parallel non-long-running requests. Every further
	// request has to wait. Applies only to non-mutating requests.
	MaxRequestsInFlight int
	// MaxMutatingRequestsInFlight is the maximum number of parallel mutating requests. Every further
	// request has to wait.
	MaxMutatingRequestsInFlight int
        // Predicate which is true for paths of long-running http requests
        LongRunningFunc apirequest.LongRunningRequestCheck
	Version *version.Info
	PublicAddress net.IP

	// RESTOptionsGetter is used to construct RESTStorage types via the generic registry.
	RESTOptionsGetter genericregistry.RESTOptionsGetter
}

type SecureServingInfo struct {
        // BindAddress is the ip:port to serve on
        BindAddress string
        // BindNetwork is the type of network to bind to - defaults to "tcp", accepts "tcp",
        // "tcp4", and "tcp6".
        BindNetwork string

        // Cert is the main server cert which is used if SNI does not match. Cert must be non-nil and is
        // allowed to be in SNICerts.
        Cert *tls.Certificate

        // CACert is an optional certificate authority used for the loopback connection of the Admission controllers.
        // If this is nil, the certificate authority is extracted from Cert or a matching SNI certificate.
        CACert *tls.Certificate

        // ClientCA is the certificate bundle for all the signers that you'll recognize for incoming client certificates
        ClientCA *x509.CertPool

        // MinTLSVersion optionally overrides the minimum TLS version supported.
        // Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants).
        MinTLSVersion uint16

        // CipherSuites optionally overrides the list of allowed cipher suites for the server.
        // Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants).
        CipherSuites []uint16
}
// NewConfig returns a Config struct with the default values
func NewConfig(codecs serializer.CodecFactory) *Config {
	return &Config{
		Serializer:			codecs,
		BuildHandlerChainFunc:		DefaultHandlerChainBuilder,
		EnableSwaggerUI:		false,
		RequestContextMapper:		apirequest.NewRequestContextMapper(),
		MinRequestTimeout:		1800,
		MaxRequestsInFlight:		400,
		MaxMutatingRequestsInFlight:	200,
	}
}

type completedConfig struct {
        *Config
}

// Complete fills in any fields not set that are required to have valid data and can be derived
// from other fields.
func (c *Config) Complete() completedConfig {
	return completedConfig{c}
}

// SkipComplete provides a way to construct a server instance without config completion.
func (c *Config) SkipComplete() completedConfig {
        return completedConfig{c}
}

func (c completedConfig) New(name string, delegate DelegationTarget) (*GenericAPIServer, error) {

	if c.Serializer == nil {
		return nil, fmt.Errorf("Genericapiserver.New() called with config.Serializer == nil")
	}
	handlerChainBuilder := func(handler http.Handler) http.Handler {
		return c.BuildHandlerChainFunc(handler, c.Config)
	}

	apiServerHandler := NewAPIServerHandler(name, handlerChainBuilder, delegate.UnprotectedHandler())
	s := &GenericAPIServer{
		SecureServingInfo: c.SecureServingInfo,
		Serializer: c.Serializer,
		Handler: apiServerHandler,
		requestContextMapper: c.RequestContextMapper,
		minRequestTimeout: time.Duration(c.MinRequestTimeout) * time.Second,
	}

	installAPIs(s, c.Config)

	return s, nil
}

// install APIs unique to this generic server
func installAPIs(s *GenericAPIServer, c *Config) {
	routes.Version{Version: c.Version}.Install(s.Handler.GoRestfulContainer)
}


func (c *Config) ApplyClientCert(clientCAFile string) (*Config, error) {
        if c.SecureServingInfo != nil {
                if len(clientCAFile) > 0 {
                        clientCAs, err := certutil.CertsFromFile(clientCAFile)
                        if err != nil {
                                return nil, fmt.Errorf("unable to load client CA file: %v", err)
                        }
                        if c.SecureServingInfo.ClientCA == nil {
                                c.SecureServingInfo.ClientCA = x509.NewCertPool()
                        }
                        for _, cert := range clientCAs {
                                c.SecureServingInfo.ClientCA.AddCert(cert)
                        }
                }
        }

        return c, nil
}

func DefaultHandlerChainBuilder(apiHandler http.Handler, c *Config) http.Handler {
	handler := apiHandler
	// FIXME (rantuttl): See ./staging/src/k8s.io/apiserver/pkg/server/config.go
	//handler := genericapifilters.WithAuthorization(apiHandler, c.RequestContextMapper, c.Authorizer, c.Serializer)
	//handler = genericapifilters.WithAuthentication(handler, c.RequestContextMapper, c.Authenticator, genericapifilters.Unauthorized(c.RequestContextMapper, c.Serializer, c.SupportsBasicAuth))
	// etc...
	// build up the chained handlers here (see filters)
	// NOTE that this looks very similar to BuildInsecureHandlerChain in apiserver/pkg/genericserver/server/insecure_handler.go
	handler = genericapifilters.WithRequestInfo(handler, NewRequestInfoResolver(c), c.RequestContextMapper)
	handler = apirequest.WithRequestContext(handler, c.RequestContextMapper)
	return handler
}

func NewRequestInfoResolver(c *Config) *apirequest.RequestInfoFactory {
	apiPrefixes := sets.NewString(strings.Trim(APIGroupPrefix, "/")) // all possible API prefixes

	return &apirequest.RequestInfoFactory{
		APIPrefixes:	  apiPrefixes,
	}
}
