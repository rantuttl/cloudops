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

package cal

import (
	"fmt"
	"strings"
	"strconv"
	"errors"
	"reflect"
	"encoding/json"

	"golang.org/x/net/context"
	"github.com/golang/glog"

	"github.com/rantuttl/cloudops/apiserver/pkg/backend"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
)

func NewCalResourceTransformer(resource string) *Transformer {
	t := &Transformer{
		Resource:       resource,
		GraphQLBodies:  make(map[Verb]*GraphQLBody),
	}
	if resource[len(resource) - 1] == 's' {
		t.SingularResource = resource[:len(resource) - 1]
	}
	return t
}

func (t *Transformer) NewGraphQLBody(verb Verb) (*GraphQLBody, error) {
	gqlBody := &GraphQLBody{}
	switch verb {
	case "create", "update", "delete", "patch":
		gqlBody.OpKeyword = mutatonKeyword
	default:
		gqlBody.OpKeyword = queryKeyword
	}
	gqlBody.FuncName = string(verb) + strings.Title(t.SingularResource)
	gqlBody.Parameters = make(map[Variable]GqlParameter)
	gqlBody.OpBody.ObjName = t.SingularResource
	gqlBody.OpBody.Arguments = make(map[Argument]Variable)
	return gqlBody, nil
}

func (t *Transformer) BackendTransformerInitializer(c backend.Config) error {
	return nil
}

func (t *Transformer) TransformToBackend(ctx context.Context, data string) (string, error) {
	glog.Info("Hit TransformToBackend")
	glog.Infof("Encoded & string-a-fied obj: %s", data)
	req, ok := request.RequestInfoFrom(ctx)
	if !ok {
		return data, errors.New("Failed to retrieve request info from context")
	}
	glog.Infof("Context: %v", ctx)
	glog.Infof("Context Request: %v", req)
	glog.Infof("Context.Resource: %s", req.Resource)
	glog.Infof("Context.Verb: %s", req.Verb)
	gqlBody, ok := t.GraphQLBodies[Verb(req.Verb)]
	if !ok {
		return data, errors.New(fmt.Sprintf("Did not find a GraphQL body for verb \"%s\"", req.Verb))
	}
	op := string(gqlBody.OpKeyword) + " " + gqlBody.FuncName
	obj := ""
	// Function Parameters
	for variable, gqlparm := range gqlBody.Parameters {
		obj = fmt.Sprint(obj + string(variable) + " : " + string(gqlparm.GqlType) + string(gqlparm.GqlTypeNullable) + ", ")
	}
	if len(obj) > 0 {
		obj = "(" + obj[:len(obj)-2] + ")"
	}
	op = op + obj + " {\n"
	// Object Alias and Name
	alias := gqlBody.OpBody.Alias
	if alias != "" {
		alias = alias + ": "
	}
	op = op + "\t" + alias + gqlBody.OpBody.ObjName
	// Object Arguments
	obj = ""
	for arg, variable := range gqlBody.OpBody.Arguments {
		obj = fmt.Sprint(obj + string(arg) + " : " + string(variable) + ", ")
	}
	if len(obj) > 0 {
		obj = "(" + obj[:len(obj)-2] + ")"
	}
	op = op + obj
	// Object Fields, Field Arguments, Field Directives, Inline Fragments, Nested Fields
	obj, _ = processObjectFields(gqlBody.OpBody.Fields, 2)
	// Object Fragments
	frags := []string{}
	for fragName, frag := range gqlBody.OpBody.FragRefs {
		obj = fmt.Sprint(obj + "\t\t" + "..." + string(fragName) + "\n")
		fragStr := fmt.Sprint("fragment " + string(fragName) + " on " + string(frag.GqlTypeRef) + " ")
		ff, _ := processObjectFields(frag.FragFields, 1)
		fragStr = fmt.Sprint("\n" + fragStr + "{\n" + ff + "}\n")
		frags = append(frags, fragStr)
	}
	if len(obj) > 0 {
		obj = " {\n" + obj + "\t}\n"
	}
	op = op + obj + "\n}"
	// TODO: Handle field variables here when we support them

	// Insert Fragments
	for _, frag := range frags {
		op = op + fmt.Sprint(frag + "\n")
	}
	glog.V(5).Infof("GraphQL query: \n%s", op)

	gqlQuery := qraphqlQuery{
		Query:	op,
		Vars:	data,
	}
	b, err := json.Marshal(gqlQuery)
	if err != nil {
		return data, errors.New("Unable to marshal GraphQL query to JSON")
	}
	return string(b), nil
}

func (a *Transformer) TransformFromBackend(ctx context.Context, data string) (string, error) {
	return data, nil
}

func processObjectFields(fields []*Field, tabstops int) (obj string, errors []error) {
	var tabs string
	for t := 1; t <= tabstops; t++ {
		tabs = tabs + "\t"
	}
	// Object Fields
	for _, f := range fields {
		obj = fmt.Sprint(obj + tabs + string(f.FieldName))
		// Arguments on fields
		args := ""
		for karg, varg := range f.FieldArguments {
			var value string
			if s, ok := varg.(string); ok {
				value = s
			} else if s, ok := varg.(int); ok {
				value = strconv.Itoa(s)
			} else if s, ok := varg.(bool); ok {
				value = strconv.FormatBool(s)
			} else {
				errors = append(errors, fmt.Errorf("Unable to convert field value to string. Unhandled type.", reflect.TypeOf(varg)))
				continue
			}
			args = fmt.Sprint(string(karg) + " : " + value + ", ")
		}
		if len(args) > 0 {
			args = fmt.Sprint("(" + args[:len(args)-2] + ")")
		}
		obj = obj + args
		// Directives on fields
		dir := ""
		for kdir, vdir := range f.GraphQLDirective {
			dir = fmt.Sprint(dir + " " + string(kdir) + "(" + "if: " + string(vdir) + ")")
		}
		obj = obj + dir
		// TODO: Inline Fragments on fields

		// Nested fields (SubFields) on fields. This is recursive
		sf, _ := processObjectFields(f.SubFields, tabstops + 1)
		if len(sf) > 0 {
			obj = obj + " {\n" + sf + "\n" + tabs+ "}"
		}

		// put it all together
		if len(obj) > 0 {
			obj = obj + "\n"
		}
	}
	return
}
