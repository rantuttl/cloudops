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

package app

import (
	//"errors"

	"github.com/golang/glog"
	"github.com/go-openapi/spec"

	"github.com/rantuttl/cloudops/cmd/app/options"
	"github.com/rantuttl/cloudops/apiserver/pkg/util/version"
	"github.com/rantuttl/cloudops/apiserver/pkg/master"
	"github.com/rantuttl/cloudops/apiserver/pkg/api"
	"github.com/rantuttl/cloudops/apiserver/pkg/authentication/authenticator"
	genericapiserver "github.com/rantuttl/cloudops/apiserver/pkg/genericserver/server"
	utilerrors "github.com/rantuttl/cloudops/apimachinery/pkg/util/errors"
	serverstorage "github.com/rantuttl/cloudops/apiserver/pkg/server/storage"
)

// Run runs the specified APIServer.  This should never exit.
func Run(runOptions *options.ServerRunOptions, stopCh <-chan struct{}) error {
        // To help debugging, immediately log version
        glog.Infof("Version: %+v", version.Get())

        server, err := CreateServerChain(runOptions, stopCh)
        if err != nil {
                return err
        }

        return server.PrepareRun().Run(stopCh)
}

func CreateServerChain(runOptions *options.ServerRunOptions, stopCh <-chan struct{}) (*genericapiserver.GenericAPIServer, error) {

	config, insecureServingOptions, err := CreateMasterAPIServerConfig(runOptions)
	if err != nil {
		return nil, err
	}
	apiServer, err := CreateAPIServer(config, genericapiserver.EmptyDelegate)
	if err != nil {
		return nil, err
	}
	//apiServer.GenericAPIServer.PrepareRun()

	s := apiServer.GenericAPIServer

	if insecureServingOptions != nil {
		insecureHandlerChain := genericapiserver.BuildInsecureHandlerChain(s.UnprotectedHandler(), config.GenericConfig)
		if err := genericapiserver.NonBlockingRun(insecureServingOptions, insecureHandlerChain, stopCh); err != nil {
			return nil, err
		}
	}
	return s, err
}

func CreateAPIServer(c *master.Config, delegateAPITarget genericapiserver.DelegationTarget) (*master.Master, error) {

	s, err := c.Complete().New(delegateAPITarget)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func CreateMasterAPIServerConfig(s *options.ServerRunOptions) (*master.Config, *genericapiserver.InsecureServingInfo, error) {
	// A good place to register plugins (admissions) when we support them

	if err := defaultOptions(s); err != nil {
		return nil, nil, err
	}

	if errs := s.Validate(); len(errs) != 0 {
		return nil, nil, utilerrors.NewAggregate(errs)
	}

	genericConfig, insecureServingOptions, err := BuildGenericConfig(s)
	if err != nil {
		return nil, nil, err
	}
	// TODO (rantuttl): Add BuildStorageFactory call here, but maybe not needed since we don't store anything
	storageFactory, err := BuildStorageFactory(s)
	if err != nil {
		return nil, nil, err
	}

	config := &master.Config{
		GenericConfig: genericConfig,
		// FIXME (rantuttl): does this config belong in storageFactory
		//APIResourceConfigSource: storageFactory.APIResourceConfigSource,
		APIResourceConfigSource: master.DefaultAPIResourceConfigSource(),
		StorageFactory: storageFactory,
		// TODO (rantuttl): Put future config info here
	}
	return config, insecureServingOptions, nil
}

func BuildGenericConfig(s *options.ServerRunOptions) (*genericapiserver.Config, *genericapiserver.InsecureServingInfo, error) {

	config := genericapiserver.NewConfig(api.Codecs)
	if err := s.GenericServerRunOptions.ApplyTo(config); err != nil {
		return nil, nil, err
	}
	if err := s.Backend.ApplyTo(config); err != nil {
		return nil, nil, err
	}
	insecureServingOptions, err := s.InsecureServing.ApplyTo(config)
	if err != nil {
		return nil, nil, err
	}
	if err := s.SecureServing.ApplyTo(config); err != nil {
		return nil, nil, err
	}
	if err := s.Authentication.ApplyTo(config); err != nil {
		return nil, nil, err
	}

	config.Authenticator, _, err = BuildAuthenticator(s)

	// TODO (rantuttl): Build Authorizor and implement Admission controls here

	v := version.Get()
	config.Version = &v

	return config, insecureServingOptions, nil
}

// defaultOptions sets necessay options to their defaults
func defaultOptions(s *options.ServerRunOptions) error {

	if err := s.GenericServerRunOptions.DefaultAdvertiseAddress(s.SecureServing); err != nil {
		return err
	}

	return nil
}

func BuildAuthenticator(s *options.ServerRunOptions) (authenticator.Request, *spec.SecurityDefinitions, error) {
	// apply any server run options to an AuthenticationConfig, then return the Authenticator that implements Request interface
	authenticatorConfig := s.Authentication.ToAuthenticationConfig()
	return authenticatorConfig.New()
}

func BuildStorageFactory(s *options.ServerRunOptions) (*serverstorage.DefaultStorageFactory, error) {
	return nil, nil
}
