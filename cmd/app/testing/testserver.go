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
package testing

import (
	"fmt"
	"net"
	"testing"

	"github.com/rantuttl/cloudops/cmd/app"
	"github.com/rantuttl/cloudops/cmd/app/options"
)

type TearDownFunc func()

func StartTestServer(t *testing.T) (tearDownForCaller TearDownFunc, err error) {
	stopCh := make(chan struct{})
	tearDown := func() {
		close(stopCh)
	}
	defer func() {
		if tearDownForCaller == nil {
			tearDown()
		}
	}()

	s := options.NewServerRunOptions()
	s.InsecureServing.BindPort = 0
	s.SecureServing.BindPort = freePort()
	s.Backend.BackendConfig.ServerList = []string{"http://localhost:3333"}

	t.Logf("Starting apiserver...")
	runErrCh := make(chan error, 1)
	server, err := app.CreateServerChain(s, stopCh)
	if err != nil {
		return nil, fmt.Errorf("Failed to create server chain: %v", err)
	}
	go func(stopCh <-chan struct{}) {
		if err := server.PrepareRun().Run(stopCh); err != nil {
			t.Logf("apiserver exited uncleanly: %v", err)
			runErrCh <- err
		}
	}(stopCh)

	return tearDown, nil
}

func StartTestServerOrDie(t *testing.T) TearDownFunc {
	for retry := 0; retry < 5 && !t.Failed(); retry++ {
		td, err := StartTestServer(t)
		if err == nil {
			return td
		}
	}
	t.Fatalf("Failed to launch server")
	return nil
}

func freePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
