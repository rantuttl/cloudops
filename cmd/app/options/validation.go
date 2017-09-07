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

// Validate checks ServerRunOptions and return a slice of found errors.
func (options *ServerRunOptions) Validate() []error {
	var errors []error
	if errs := options.SecureServing.Validate(); len(errs) > 0 {
		errors = append(errors, errs...)
	}
	if errs := options.Authentication.Validate(); len(errs) > 0 {
		errors = append(errors, errs...)
	}
	if errs := options.InsecureServing.Validate("insecure-port"); len(errs) > 0 {
		errors = append(errors, errs...)
	}
	return errors
}
