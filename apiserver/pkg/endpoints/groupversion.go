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

package endpoints

import (
	"time"
	"path"

	"github.com/emicklei/go-restful"

	"github.com/rantuttl/cloudops/apimachinery/pkg/api/meta"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	utilerrors "github.com/rantuttl/cloudops/apimachinery/pkg/util/errors"
	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	apirequest "github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
	"github.com/rantuttl/cloudops/apiserver/pkg/registry/rest"
)

type APIGroupVersion struct {
	Root		string
	Storage		map[string]rest.Storage
	GroupVersion	schema.GroupVersion
	Mapper		meta.RESTMapper
	Serializer	runtime.NegotiatedSerializer
	Typer		runtime.ObjectTyper
	Creater		runtime.ObjectCreater
	Copier		runtime.ObjectCopier // performs deepcopies per API Group type
	Convertor	runtime.ObjectConvertor
	Defaulter	runtime.ObjectDefaulter // for setting our defaults per API Group type
	Linker		runtime.SelfLinker
	Context		apirequest.RequestContextMapper

	// SubresourceGroupVersionKind contains the GroupVersionKind overrides for each subresource that is
	// accessible from this API group version.
	SubresourceGroupVersionKind map[string]schema.GroupVersionKind
	MinRequestTimeout time.Duration
}

func (g *APIGroupVersion) InstallREST(container *restful.Container) error {
	installer := g.newInstaller()
	ws := installer.NewWebService()
	apiResources, registrationErrors := installer.Install(ws)
	// TODO (rantuttl): Figure out discovery, and add all resources to list all resources, i.e., GET "/"
	lister := staticLister{apiResources}
	// make compiler happy
	_ = lister

	container.Add(ws)
	return utilerrors.NewAggregate(registrationErrors)
}

// staticLister implements the APIResourceLister interface
type staticLister struct {
	list []metav1.APIResource
}

func (g *APIGroupVersion) newInstaller() *APIInstaller {
	// /api/<group-name>/<version>
	prefix := path.Join(g.Root, g.GroupVersion.Group, g.GroupVersion.Version)
	installer := &APIInstaller{
		group:			g,
		prefix:			prefix,
		minRequestTimeout:	g.MinRequestTimeout,
	}
	return installer
}
