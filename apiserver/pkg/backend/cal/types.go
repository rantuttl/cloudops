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

// TODO (rantuttl): Move to CAL client???
type qraphqlQuery struct {
	Query	string	`json:"query"`
	OpName	string	`json:"operationName,omitempty"`
	Vars	string	`json:"variables"`
}

type opKeyword string

const (
        queryKeyword    opKeyword = "query"
        mutatonKeyword  opKeyword = "mutation"
)

type graphQLType string

const (
        GQLKIND graphQLType = "Kind"
        GQLAPIVERSION graphQLType = "ApiVersion"
        GQLMETADATA graphQLType = "Metadata"
        GQLSPEC graphQLType = "Spec"
        GQLSTATUS graphQLType = "Status"
        GQLACCOUNT graphQLType = "Account"
)

type Variable string

const (
        KIND Variable = "$kind"
        APIVERSION Variable = "$apiVersion"
        METADATA Variable = "$metadata"
        SPEC Variable = "$spec"
        STATUS Variable = "$status"
)

type graphqlEnum int64

// Sub-fields of GraphQL type METADATA; aligns closely to ObjectMeta in apimachinery/pkg/apigroups/meta/v1/types.go
const (
        NAME graphqlEnum = iota
        NAMESPACE
        UID
        RESOURCEVERSION
        CREATETIMESTAMP
        DELETETIMESTAMP
        LABELS
        ANNOTATIONS
        CLUSTERNAME
)

type gqlStringMap map[graphqlEnum]string

var MetadataMap = gqlStringMap{
        NAME:                   "Name",
        NAMESPACE:              "Namespace",
        UID:                    "Uid",
        RESOURCEVERSION:        "ResourceVersion",
        CREATETIMESTAMP:        "CreateTimestamp",
        DELETETIMESTAMP:        "DeleteTimestamp",
        LABELS:                 "Labels",
        ANNOTATIONS:            "Annotations",
        CLUSTERNAME:            "ClusterName",
}

type Verb string

const (
	CREATE Verb = "create"
	GET Verb = "get"
	DELETE Verb = "delete"
	UPDATE Verb = "update"
)

type Transformer struct {
	Resource		string	  // registered API resource
	SingularResource	string
	GraphQLBodies		map[Verb]*GraphQLBody
}

type GraphQLBody struct {
        OpKeyword       opKeyword
        FuncName	string
	Parameters	map[Variable]GqlParameter
	OpBody		ObjectBody
}

type GqlParameter struct {
        GqlType         graphQLType
        GqlTypeNullable nullable
}

type nullable string

const (
	NULLABLE nullable = ""
	NON_NULLABLE nullable = "!"
)

type Argument string

const (
        ARGKIND Argument = "kind"
        ARGAPIVERSION Argument = "apiVersion"
        ARGMETADATA Argument = "metadata"
        ARGSPEC Argument = "spec"
        ARGSTATUS Argument = "status"
)

type FragName string

type ObjectBody struct {
	// ObjectName
	Alias           string
	ObjName		string
	Arguments       map[Argument]Variable
	// Field Fragments
	Fields		[]*Field
	FragRefs	map[FragName]*Fragment
}

type Fragment struct {
	GqlTypeRef	graphQLType
	FragFields	[]*Field
}

type Field struct {
	FieldName		Argument
	FieldArguments		map[string]interface{}
	GraphQLDirective	map[GqlDirective]Variable
	SubFields		[]*Field
	InlineFrags		map[graphQLType][]graphqlEnum
}

type GqlDirective string

const (
        INCLUDE GqlDirective = "@include"
        SKIP GqlDirective = "@skip"
)

type GraphQLDirective struct {
        Directive       GqlDirective
        DirectiveBool   Variable // referenced value must resolve to boolean type
}
