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

package validation

import (
	"math/rand"
	"strings"
	"testing"

	metav1 "github.com/rantuttl/cloudops/apimachinery/pkg/apigroups/meta/v1"
	"github.com/rantuttl/cloudops/apimachinery/pkg/util/validation/field"
)

const (
	maxLengthErrMsg = "must be no more than"
	namePartErrMsg  = "name part must consist of"
	nameErrMsg      = "a qualified name must consist of"
)

// Ensure custom name functions are allowed
func TestValidateObjectMetaCustomName(t *testing.T) {
	errs := ValidateObjectMeta(
		&metav1.ObjectMeta{Name: "test", GenerateName: "foo"},
		false,
		func(s string, prefix bool) []string {
			if s == "test" {
				return nil
			}
			return []string{"name-gen"}
		},
		field.NewPath("field"))
	if len(errs) != 1 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !strings.Contains(errs[0].Error(), "name-gen") {
		t.Errorf("unexpected error message: %v", errs)
	}
}

// Ensure namespace names follow dns label format
func TestValidateObjectMetaNamespaces(t *testing.T) {
	errs := ValidateObjectMeta(
		&metav1.ObjectMeta{Name: "test", Namespace: "foo.bar"},
		true,
		func(s string, prefix bool) []string {
			return nil
		},
		field.NewPath("field"))
	if len(errs) != 1 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !strings.Contains(errs[0].Error(), `Invalid value: "foo.bar"`) {
		t.Errorf("unexpected error message: %v", errs)
	}
	maxLength := 63
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, maxLength+1)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	errs = ValidateObjectMeta(
		&metav1.ObjectMeta{Name: "test", Namespace: string(b)},
		true,
		func(s string, prefix bool) []string {
			return nil
		},
		field.NewPath("field"))
	if len(errs) != 2 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !strings.Contains(errs[0].Error(), "Invalid value") || !strings.Contains(errs[1].Error(), "Invalid value") {
		t.Errorf("unexpected error message: %v", errs)
	}
}

// Ensure trailing slash is allowed in generate name
func TestValidateObjectMetaTrimsTrailingSlash(t *testing.T) {
	errs := ValidateObjectMeta(
		&metav1.ObjectMeta{Name: "test", GenerateName: "foo-"},
		false,
		NameIsDNSSubdomain,
		field.NewPath("field"))
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestValidateAnnotations(t *testing.T) {
	successCases := []map[string]string{
		{"simple": "bar"},
		{"now-with-dashes": "bar"},
		{"1-starts-with-num": "bar"},
		{"1234": "bar"},
		{"simple/simple": "bar"},
		{"now-with-dashes/simple": "bar"},
		{"now-with-dashes/now-with-dashes": "bar"},
		{"now.with.dots/simple": "bar"},
		{"now-with.dashes-and.dots/simple": "bar"},
		{"1-num.2-num/3-num": "bar"},
		{"1234/5678": "bar"},
		{"1.2.3.4/5678": "bar"},
		{"UpperCase123": "bar"},
		{"a": strings.Repeat("b", totalAnnotationSizeLimitB-1)},
		{
			"a": strings.Repeat("b", totalAnnotationSizeLimitB/2-1),
			"c": strings.Repeat("d", totalAnnotationSizeLimitB/2-1),
		},
	}
	for i := range successCases {
		errs := ValidateAnnotations(successCases[i], field.NewPath("field"))
		if len(errs) != 0 {
			t.Errorf("case[%d] expected success, got %#v", i, errs)
		}
	}

	nameErrorCases := []struct {
		annotations map[string]string
		expect      string
	}{
		{map[string]string{"nospecialchars^=@": "bar"}, namePartErrMsg},
		{map[string]string{"cantendwithadash-": "bar"}, namePartErrMsg},
		{map[string]string{"only/one/slash": "bar"}, nameErrMsg},
		{map[string]string{strings.Repeat("a", 254): "bar"}, maxLengthErrMsg},
	}
	for i := range nameErrorCases {
		errs := ValidateAnnotations(nameErrorCases[i].annotations, field.NewPath("field"))
		if len(errs) != 1 {
			t.Errorf("case[%d]: expected failure", i)
		} else {
			if !strings.Contains(errs[0].Detail, nameErrorCases[i].expect) {
				t.Errorf("case[%d]: error details do not include %q: %q", i, nameErrorCases[i].expect, errs[0].Detail)
			}
		}
	}
	totalSizeErrorCases := []map[string]string{
		{"a": strings.Repeat("b", totalAnnotationSizeLimitB)},
		{
			"a": strings.Repeat("b", totalAnnotationSizeLimitB/2),
			"c": strings.Repeat("d", totalAnnotationSizeLimitB/2),
		},
	}
	for i := range totalSizeErrorCases {
		errs := ValidateAnnotations(totalSizeErrorCases[i], field.NewPath("field"))
		if len(errs) != 1 {
			t.Errorf("case[%d] expected failure", i)
		}
	}
}
