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
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/emicklei/go-restful"

	genericmux "github.com/rantuttl/cloudops/apiserver/pkg/genericserver/server/mux"
)

type APIServerHandler struct {
	// FullHandlerChain is the one that is eventually served with.  It should include the full filter
	// chain and then call the Director.
	FullHandlerChain http.Handler
	// The registered APIs.  InstallAPIs uses this.  Other servers probably shouldn't access this directly.
	GoRestfulContainer *restful.Container
	NonGoRestfulMux *genericmux.PathRecorderMux
	Director http.Handler
}

// HandlerChainBuilderFn is used to wrap the GoRestfulContainer handler using the provided handler chain.
// It is normally used to apply filtering like authentication and authorization
type HandlerChainBuilderFn func(apiHandler http.Handler) http.Handler

func NewAPIServerHandler(name string, handlerChainBuilder HandlerChainBuilderFn, notFoundHandler http.Handler) *APIServerHandler {
	nonGoRestfulMux := genericmux.NewPathRecorderMux(name)
	if notFoundHandler != nil {
		nonGoRestfulMux.NotFoundHandler(notFoundHandler)
	}

	gorestfulContainer := restful.NewContainer()
	gorestfulContainer.ServeMux = http.NewServeMux()
	gorestfulContainer.Router(restful.CurlyRouter{})
	// FIXME (rantuttl): finish recovery handling (see ./staging/src/k8s.io/apiserver/pkg/server/handler.go)
	//gorestfulContainer.RecoverHandler(func(panicReason interface{}, ...

	//gorestfulContainer.ServiceErrorHandler(func(serviceErr restful.ServiceError, ...

	director := director{
		name:			name,
		goRestfulContainer:	gorestfulContainer,
		nonGoRestfulMux:	nonGoRestfulMux,
	}

	return &APIServerHandler{
		FullHandlerChain:	handlerChainBuilder(director),
		GoRestfulContainer:	gorestfulContainer,
		Director:		director,
	}
}

type director struct {
        name               string
        goRestfulContainer *restful.Container
        nonGoRestfulMux    *genericmux.PathRecorderMux
}


func (d director) ServeHTTP(w http.ResponseWriter, req *http.Request) {
        path := req.URL.Path

        // check to see if our webservices want to claim this path
        for _, ws := range d.goRestfulContainer.RegisteredWebServices() {
                switch {
                case strings.HasPrefix(path, ws.RootPath()):
                        // ensure an exact match or a path boundary match
                        if len(path) == len(ws.RootPath()) || path[len(ws.RootPath())] == '/' {
                                glog.V(5).Infof("%v: %v %q satisfied by gorestful with webservice %v", d.name, req.Method, path, ws.RootPath())
                                // don't use servemux here because gorestful servemuxes get messed up when removing webservices
                                // TODO fix gorestful, remove TPRs, or stop using gorestful
                                d.goRestfulContainer.Dispatch(w, req)
                                return
                        }
                }
        }

        // if we didn't find a match, then we just skip gorestful altogether
        glog.V(5).Infof("%v: %v %q satisfied by nonGoRestful", d.name, req.Method, path)
        d.nonGoRestfulMux.ServeHTTP(w, req)
}

// ServeHTTP makes it an http.Handler
func (a *APIServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        a.FullHandlerChain.ServeHTTP(w, r)
}
