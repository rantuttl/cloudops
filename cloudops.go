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

package main

import (
	"math/rand"
	"time"
	"fmt"
	"os"

	"github.com/rantuttl/cloudops/apiserver/pkg/util/logs"
	"github.com/rantuttl/cloudops/apiserver/pkg/util/flag"
	"github.com/rantuttl/cloudops/cmd/app"
	"github.com/rantuttl/cloudops/apimachinery/pkg/util/wait"
	runoptions "github.com/rantuttl/cloudops/cmd/app/options"

	"github.com/spf13/pflag"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	// to see what happens in the package, uncomment the following
	//restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshortfile))

        // TODO (rantuttl):
        // 1. Maybe grab a Config (file and/or command line flags) and 'type' it to a config struct
        // 2. call an Install with the Config struct

	// Get the API server with its run options
	s := runoptions.NewServerRunOptions()
	s.AddFlags(pflag.CommandLine)
	flag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	flag.PrintAndExitIfRequested()

	if err := app.Run(s, wait.NeverStop); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
