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

package responsewriters

import (
	"fmt"
	"net/http"
	"strconv"
	"encoding/json"

	utilruntime "github.com/rantuttl/cloudops/apimachinery/pkg/util/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/request"
	"github.com/rantuttl/cloudops/apiserver/pkg/endpoints/handlers/negotiation"
)

// WriteObject renders a returned runtime.Object to the response as a stream or an encoded object. If the object
// returned by the response implements rest.ResourceStreamer that interface will be used to render the
// response. The Accept header and current API version will be passed in, and the output will be copied
// directly to the response body. If content type is returned it is used, otherwise the content type will
// be "application/octet-stream". All other objects are sent to standard JSON serialization.
func WriteObject(ctx request.Context, statusCode int, gv schema.GroupVersion, s runtime.NegotiatedSerializer, object runtime.Object, w http.ResponseWriter, req *http.Request) {
	// FIXME (rantuttl): Implement streaming resources
        //stream, ok := object.(rest.ResourceStreamer)
        //if ok {
        //        StreamObject(ctx, statusCode, gv, s, stream, w, req)
        //        return
        //}
        WriteObjectNegotiated(ctx, s, gv, w, req, statusCode, object)
}


// SerializeObject renders an object in the content type negotiated by the client using the provided encoder.
// The context is optional and can be nil.
func SerializeObject(mediaType string, encoder runtime.Encoder, w http.ResponseWriter, req *http.Request, statusCode int, object runtime.Object) {
	w.Header().Set("Content-Type", mediaType)
	w.WriteHeader(statusCode)

	if err := encoder.Encode(object, w); err != nil {
		errorJSONFatal(err, encoder, w)
	}
}

// WriteObjectNegotiated renders an object in the content type negotiated by the client.
// The context is optional and can be nil.
func WriteObjectNegotiated(ctx request.Context, s runtime.NegotiatedSerializer, gv schema.GroupVersion, w http.ResponseWriter, req *http.Request, statusCode int, object runtime.Object) {
	serializer, err := negotiation.NegotiateOutputSerializer(req, s)
	if err != nil {
		status := ErrorToAPIStatus(err)
		WriteRawJSON(int(status.Code), status, w)
		return
	}

	encoder := s.EncoderForVersion(serializer.Serializer, gv)
	SerializeObject(serializer.MediaType, encoder, w, req, statusCode, object)
}

// ErrorNegotiated renders an error to the response. Returns the HTTP status code of the error.
// The context is optional and may be nil.
func ErrorNegotiated(ctx request.Context, err error, s runtime.NegotiatedSerializer, gv schema.GroupVersion, w http.ResponseWriter, req *http.Request) int {
	status := ErrorToAPIStatus(err)
	code := int(status.Code)
	// when writing an error, check to see if the status indicates a retry after period
	if status.Details != nil && status.Details.RetryAfterSeconds > 0 {
		delay := strconv.Itoa(int(status.Details.RetryAfterSeconds))
		w.Header().Set("Retry-After", delay)
	}

	if code == http.StatusNoContent {
		w.WriteHeader(code)
		return code
	}

	WriteObjectNegotiated(ctx, s, gv, w, req, code, status)
	return code
}

// errorJSONFatal renders an error to the response, and if codec fails will render plaintext.
// Returns the HTTP status code of the error.
func errorJSONFatal(err error, codec runtime.Encoder, w http.ResponseWriter) int {
	utilruntime.HandleError(fmt.Errorf("apiserver was unable to write a JSON response: %v", err))
	status := ErrorToAPIStatus(err)
	code := int(status.Code)
	output, err := runtime.Encode(codec, status)
	if err != nil {
		w.WriteHeader(code)
		fmt.Fprintf(w, "%s: %s", status.Reason, status.Message)
		return code
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(output)
	return code
}

// WriteRawJSON writes a non-API object in JSON.
func WriteRawJSON(statusCode int, object interface{}, w http.ResponseWriter) {
	output, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(output)
}
