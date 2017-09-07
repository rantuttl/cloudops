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
	"log"
	"math/rand"
	"time"
	"fmt"
	"os"
	"net/http"

	"github.com/rantuttl/cloudops/resources"
	"github.com/rantuttl/cloudops/apiserver/pkg/util/logs"
	"github.com/rantuttl/cloudops/apiserver/pkg/util/flag"
	"github.com/rantuttl/cloudops/cmd/app"
	"github.com/rantuttl/cloudops/apimachinery/pkg/util/wait"
	runoptions "github.com/rantuttl/cloudops/cmd/app/options"

	"github.com/emicklei/go-restful"
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

	wsContainer := restful.NewContainer()

	a := resources.AccountResource{map[string]resources.Account{}}
	a.Register(wsContainer)

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	//config := swagger.Config{
	//	WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
	//	WebServicesUrl: "http://localhost:8080",
	//	ApiPath:        "/apidocs.json",

		// Optionally, specify where the UI is located
	//	SwaggerPath:     "/apidocs/",
	//	SwaggerFilePath: "/Users/emicklei/xProjects/swagger-ui/dist"}
	//swagger.RegisterSwaggerService(config, wsContainer)

	log.Print("Start listening on localhost:8080")
	server := &http.Server{Addr: ":8080", Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
