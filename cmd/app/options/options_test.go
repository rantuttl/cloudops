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
	"testing"

	"github.com/spf13/pflag"
)

func TestAddFlagsFlag(t *testing.T) {
	// TODO (rantuttl): Only supports testing backend-servers flag for now.
	// Add remaining server options as necessary
	f := pflag.NewFlagSet("addflagstest", pflag.ContinueOnError)
	s := NewServerRunOptions()
	s.AddFlags(f)
	if len(s.Backend.BackendConfig.ServerList) > 0 {
		t.Errorf("Expected s.Backend.BackendConfig.ServerList to be empty")
	}

	args := []string{
		"--backend-servers=http://localhost:3333",
	}
	f.Parse(args)
	if len(s.Backend.BackendConfig.ServerList) == 0 {
		t.Errorf("Expected s.Backend.BackendConfig.ServerList to have one entry")
	}
}
