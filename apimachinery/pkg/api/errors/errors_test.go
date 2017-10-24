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
package errors

import (
	"fmt"
	"testing"
	"reflect"
	"net/http"

	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	"github.com/rantuttl/cloudops/apimachinery/pkg/util/validation/field"
)

func resource(resource string) schema.GroupResource {
	return schema.GroupResource{Group: "Tests", Resource: resource}
}

func kind(kind string) schema.GroupKind {
	return schema.GroupKind{Group: "Tests", Kind: kind}
}

func TestErrorNew(t *testing.T) {
	var err *StatusError
	var foobarErr error = fmt.Errorf("Foobar error")

	err = NewNotFound(resource("errors"), "NewNotFound")
	if !IsNotFound(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonNotFound, string(reasonForError(err)))
	}

	err = NewAlreadyExists(resource("errors"), "NewAlreadyExists")
	if !IsAlreadyExists(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonAlreadyExists, string(reasonForError(err)))
	}

	err = NewUnauthorized("")
	if !IsUnauthorized(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonUnauthorized, string(reasonForError(err)))
	}

	err = NewForbidden(resource("errors"), "NewForbidden", foobarErr)
	if !IsForbidden(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonForbidden, string(reasonForError(err)))
	}

	err = NewConflict(resource("errors"), "NewConflict", foobarErr)
	if !IsConflict(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonConflict, string(reasonForError(err)))
	}

	err = NewBadRequest("bad request")
	if !IsBadRequest(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonBadRequest, string(reasonForError(err)))
	}

	err = NewServiceUnavailable("new service unavailable")
	if !IsServiceUnavailable(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonServiceUnavailable, string(reasonForError(err)))
	}

	err = NewMethodNotSupported(resource("errors"), "create")
	if !IsMethodNotSupported(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonMethodNotAllowed, string(reasonForError(err)))
	}

	err = NewServerTimeout(resource("errors"), "foobar", 30)
	if !IsServerTimeout(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonServerTimeout, string(reasonForError(err)))
	}

	err = NewServerTimeoutForKind(kind("errors"), "foobar", 30)
	if !IsServerTimeout(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonServerTimeout, string(reasonForError(err)))
	}

	err = NewInternalError(foobarErr)
	if !IsInternalError(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonInternalError, string(reasonForError(err)))
	}

	err = NewTimeoutError("foobar", 30)
	if !IsTimeout(err) {
		t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonTimeout, string(reasonForError(err)))
	}

}

func TestNewInvalid(t *testing.T) {
	testCases := []struct {
		Err     *field.Error
		Details *metav1.StatusDetails
	}{
		{
			field.Duplicate(field.NewPath("field[0].name"), "bar"),
			&metav1.StatusDetails{
				Group: "Tests",
				Kind: "errors",
				Name: "name",
				Causes: []metav1.StatusCause{{
					Type:  metav1.CauseTypeFieldValueDuplicate,
					Field: "field[0].name",
				}},
			},
		},
		{
			field.Invalid(field.NewPath("field[0].name"), "bar", "detail"),
			&metav1.StatusDetails{
				Group: "Tests",
				Kind: "errors",
				Name: "name",
				Causes: []metav1.StatusCause{{
					Type:  metav1.CauseTypeFieldValueInvalid,
					Field: "field[0].name",
				}},
			},
		},
		{
			field.NotFound(field.NewPath("field[0].name"), "bar"),
			&metav1.StatusDetails{
				Group: "Tests",
				Kind: "errors",
				Name: "name",
				Causes: []metav1.StatusCause{{
					Type:  metav1.CauseTypeFieldValueNotFound,
					Field: "field[0].name",
				}},
			},
		},
		{
			field.NotSupported(field.NewPath("field[0].name"), "bar", nil),
			&metav1.StatusDetails{
				Group: "Tests",
				Kind: "errors",
				Name: "name",
				Causes: []metav1.StatusCause{{
					Type:  metav1.CauseTypeFieldValueNotSupported,
					Field: "field[0].name",
				}},
			},
		},
		{
			field.Required(field.NewPath("field[0].name"), ""),
			&metav1.StatusDetails{
				Group: "Tests",
				Kind: "errors",
				Name: "name",
				Causes: []metav1.StatusCause{{
					Type:  metav1.CauseTypeFieldValueRequired,
					Field: "field[0].name",
				}},
			},
		},
	}
	for i, testCase := range testCases {
		vErr, expected := testCase.Err, testCase.Details
		expected.Causes[0].Message = vErr.ErrorBody()
		err := NewInvalid(kind("errors"), "name", field.ErrorList{vErr})
		status := err.ErrStatus
		if status.Code != 422 || status.Reason != metav1.StatusReasonInvalid {
			t.Errorf("%d: unexpected status: %#v", i, status)
		}
		if !reflect.DeepEqual(expected, status.Details) {
			t.Errorf("%d: expected %#v, got %#v", i, expected, status.Details)
		}
	}
}

type verbCodeType map[string][]int

var verbCodeMap = verbCodeType{
	"POST": {
		http.StatusConflict,
		http.StatusNotFound,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusMethodNotAllowed,
		StatusUnprocessableEntity,
		StatusServerTimeout,
		StatusTooManyRequests,
		500,
	},
	"GET": {
		http.StatusConflict,
		http.StatusNotFound,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusMethodNotAllowed,
		StatusUnprocessableEntity,
		StatusServerTimeout,
		StatusTooManyRequests,
		500,
	},
}

func TestNewGenericServerResponse(t *testing.T) {
	r := resource("errors")
	for verb, codes := range verbCodeMap {
		for _, code := range codes {
			err := NewGenericServerResponse(code, verb, r, "name", "foobar error", 30, false)
			switch code {
			case http.StatusConflict:
				if verb == "POST" {
					if !IsAlreadyExists(err) {
						t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonAlreadyExists, string(reasonForError(err)))
					}
				} else {
					if !IsConflict(err) {
						t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonConflict, string(reasonForError(err)))
					}
				}
			case http.StatusNotFound:
				if !IsNotFound(err) {
					t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonNotFound, string(reasonForError(err)))
				}
			case http.StatusBadRequest:
				if !IsBadRequest(err) {
					t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonBadRequest, string(reasonForError(err)))
				}
			case http.StatusUnauthorized:
				if !IsUnauthorized(err) {
					t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonUnauthorized, string(reasonForError(err)))
				}
			case http.StatusForbidden:
				if !IsForbidden(err) {
					t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonForbidden, string(reasonForError(err)))
				}
			case http.StatusMethodNotAllowed:
				if !IsMethodNotSupported(err) {
					t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonMethodNotAllowed, string(reasonForError(err)))
				}
			case StatusUnprocessableEntity:
				if !IsInvalid(err) {
					t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonInvalid, string(reasonForError(err)))
				}
			case StatusServerTimeout:
				if !IsServerTimeout(err) {
					t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonServerTimeout, string(reasonForError(err)))
				}
			case StatusTooManyRequests:
				if !IsTimeout(err) {
					t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonTimeout, string(reasonForError(err)))
				}
			default:
				if code >= 500 {
					if !IsInternalError(err) {
						t.Errorf("Expected \"err\" to be %s, but received %s", metav1.StatusReasonInternalError, string(reasonForError(err)))
					}
				}
			}
		}
	}
}

type TestType struct{}

func (obj *TestType) GetObjectKind() schema.ObjectKind { return schema.EmptyObjectKind }

func TestFromObject(t *testing.T) {
	table := []struct {
		obj     runtime.Object
		message string
	}{
		{&metav1.Status{Message: "foobar"}, "foobar"},
		{&TestType{}, "unexpected object: &{}"},
	}

	for _, item := range table {
		if e, a := item.message, FromObject(item.obj).Error(); e != a {
			t.Errorf("Expected %v, got %v", e, a)
		}
	}
}
