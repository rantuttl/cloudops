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

package storage

import (
	"fmt"
	"errors"

	"golang.org/x/net/context"

	"github.com/rantuttl/cloudops/apiserver/pkg/backend"
	"github.com/rantuttl/cloudops/apiserver/pkg/backend/cal"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
)

// Account resource wrapper around BackendTransformer (CalTransformer)
type accountTransformer struct {
	resource string
	transformer backend.BackendTransformer
}

func (a *accountTransformer) BackendTransformerInitializer(c backend.Config) error {
	// TODO (rantuttl): When config carries the backend type, switch on it to select the proper transformer
	// Right now, there's only CAL
	t := cal.NewCalResourceTransformer(a.resource)
	verbs := []cal.Verb{cal.CREATE, cal.DELETE, cal.UPDATE, cal.GET}
	for _, v := range verbs {
		gqlBody, err := t.NewGraphQLBody(v)
		if err != nil {
			panic(err)
		}
		t.GraphQLBodies[v] = gqlBody
	}

	// CREATE
	gqlBody := t.GraphQLBodies[cal.CREATE]
	gqlBody.Parameters[cal.KIND] = cal.GqlParameter{GqlType: cal.GQLKIND, GqlTypeNullable: cal.NON_NULLABLE}
	gqlBody.Parameters[cal.APIVERSION] = cal.GqlParameter{GqlType: cal.GQLAPIVERSION, GqlTypeNullable: cal.NON_NULLABLE}
	gqlBody.Parameters[cal.METADATA] = cal.GqlParameter{GqlType: cal.GQLMETADATA, GqlTypeNullable: cal.NON_NULLABLE}
	gqlBody.Parameters[cal.SPEC] = cal.GqlParameter{GqlType: cal.GQLSPEC, GqlTypeNullable: cal.NON_NULLABLE}
	gqlBody.Parameters[cal.STATUS] = cal.GqlParameter{GqlType: cal.GQLSTATUS, GqlTypeNullable: cal.NON_NULLABLE}
	gqlBody.OpBody.Arguments[cal.ARGKIND] = cal.KIND
	gqlBody.OpBody.Arguments[cal.ARGAPIVERSION] = cal.APIVERSION
	gqlBody.OpBody.Arguments[cal.ARGMETADATA] = cal.METADATA
	gqlBody.OpBody.Arguments[cal.ARGSPEC] = cal.SPEC
	gqlBody.OpBody.Arguments[cal.ARGSTATUS] = cal.STATUS
	fieldNames := []cal.Argument{cal.ARGKIND, cal.ARGAPIVERSION, cal.ARGMETADATA, cal.ARGSPEC, cal.ARGSTATUS}

	fields := []*cal.Field{}
	for _, f := range fieldNames {
		field := &cal.Field{
			FieldName:		f,
			FieldArguments:		make(map[string]interface{}),
			GraphQLDirective:	make(map[cal.GqlDirective]cal.Variable),
		}
		switch f {
		case cal.ARGMETADATA:
			field.FieldArguments["uid"] = 123
			subFields := []*cal.Field{}
			for _, v := range cal.MetadataMap {
				sfield := &cal.Field{
					FieldName:	cal.Argument(v),
				}
				subFields = append(subFields, sfield)
			}
			field.SubFields = subFields
		case cal.ARGSTATUS:
			field.GraphQLDirective[cal.INCLUDE] = cal.STATUS
		}
		fields = append(fields, field)
	}
	gqlBody.OpBody.Fields = fields
	fieldNames = []cal.Argument{cal.ARGKIND, cal.ARGAPIVERSION, cal.ARGMETADATA, cal.ARGSPEC, cal.ARGSTATUS}
	fragfields := []*cal.Field{}
	for _, f := range fieldNames {
		field := &cal.Field{
			FieldName:		f,
			FieldArguments:		make(map[string]interface{}),
			GraphQLDirective:	make(map[cal.GqlDirective]cal.Variable),
		}
		switch f {
		case cal.ARGMETADATA:
			field.FieldArguments["uid"] = 246
			subFields := []*cal.Field{}
			for _, v := range cal.MetadataMap {
				sfield := &cal.Field{
					FieldName:	cal.Argument(v),
				}
				subFields = append(subFields, sfield)
			}
			field.SubFields = subFields
		case cal.ARGSTATUS:
			field.GraphQLDirective[cal.INCLUDE] = cal.STATUS
		}
		fragfields = append(fragfields, field)
	}
	frags := make(map[cal.FragName]*cal.Fragment)
	frags["fieldList"] = &cal.Fragment{GqlTypeRef: cal.GQLACCOUNT, FragFields: fragfields}
	gqlBody.OpBody.FragRefs = frags

	a.transformer = t
	return a.transformer.BackendTransformerInitializer(c)
}

func (a *accountTransformer) TransformToBackend(ctx context.Context, data string) (string, error) {
	req, ok := request.RequestInfoFrom(ctx)
	if !ok {
		return data, errors.New("Failed to retrieve request info from context")
	}
	if req.Resource != "accounts" {
		return data, errors.New(fmt.Sprintf("Transformation of \"accounts\" resource not permitted with resource: %s", req.Resource))
	}
	return a.transformer.TransformToBackend(ctx, data)
}

func (a *accountTransformer) TransformFromBackend(ctx context.Context, data string) (string, error) {
	return a.transformer.TransformFromBackend(ctx, data)
}
