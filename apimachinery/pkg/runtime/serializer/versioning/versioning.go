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

package versioning

import (
	"io"
	"github.com/golang/glog"

	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime"
	"github.com/rantuttl/cloudops/apimachinery/pkg/runtime/schema"
	utilruntime "github.com/rantuttl/cloudops/apimachinery/pkg/util/runtime"
)

// NewCodecForScheme is a convenience method for callers that are using a scheme.
func NewCodecForScheme(
	// TODO: I should be a scheme interface?
	scheme *runtime.Scheme,
	encoder runtime.Encoder,
	decoder runtime.Decoder,
	encodeVersion runtime.GroupVersioner,
	decodeVersion runtime.GroupVersioner,
) runtime.Codec {
	return NewCodec(encoder, decoder, runtime.UnsafeObjectConvertor(scheme), scheme, scheme, scheme, nil, encodeVersion, decodeVersion)
}

// NewDefaultingCodecForScheme is a convenience method for callers that are using a scheme.
func NewDefaultingCodecForScheme(
	// TODO: I should be a scheme interface?
	scheme *runtime.Scheme,
	encoder runtime.Encoder,
	decoder runtime.Decoder,
	encodeVersion runtime.GroupVersioner,
	decodeVersion runtime.GroupVersioner,
) runtime.Codec {
	return NewCodec(encoder, decoder, runtime.UnsafeObjectConvertor(scheme), scheme, scheme, scheme, scheme, encodeVersion, decodeVersion)
}

// NewCodec takes objects in their internal versions and converts them to external versions before
// serializing them. It assumes the serializer provided to it only deals with external versions.
// This class is also a serializer, but is generally used with a specific version.
func NewCodec(
	encoder runtime.Encoder,
	decoder runtime.Decoder,
	convertor runtime.ObjectConvertor,
	creater runtime.ObjectCreater,
	copier runtime.ObjectCopier,
	typer runtime.ObjectTyper,
	defaulter runtime.ObjectDefaulter,
	encodeVersion runtime.GroupVersioner,
	decodeVersion runtime.GroupVersioner,
) runtime.Codec {
	internal := &codec{
		encoder:   encoder,
		decoder:   decoder,
		convertor: convertor,
		creater:   creater,
		copier:    copier,
		typer:     typer,
		defaulter: defaulter,

		encodeVersion: encodeVersion,
		decodeVersion: decodeVersion,
	}
	return internal
}

type codec struct {
	encoder   runtime.Encoder
	decoder   runtime.Decoder
	convertor runtime.ObjectConvertor
	creater   runtime.ObjectCreater
	copier    runtime.ObjectCopier
	typer     runtime.ObjectTyper
	defaulter runtime.ObjectDefaulter

	encodeVersion runtime.GroupVersioner
	decodeVersion runtime.GroupVersioner
}

// Decode attempts a decode of the object, then tries to convert it to the internal version. If into is provided and the decoding is
// successful, the returned runtime.Object will be the value passed as into. Note that this may bypass conversion if you pass an
// into that matches the serialized version.
func (c *codec) Decode(data []byte, defaultGVK *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	glog.Info("Stage: 0 Decode")
	versioned, isVersioned := into.(*runtime.VersionedObjects)
	if isVersioned {
		into = versioned.Last()
	}

	glog.Info("Decode Stage A")
	obj, gvk, err := c.decoder.Decode(data, defaultGVK, into)
	glog.Infof("Decode Stage A: obj: %v, gvk: %v, into: %v", obj, gvk, into)
	if err != nil {
		return nil, gvk, err
	}

	glog.Info("Decode Stage B")
	if d, ok := obj.(runtime.NestedObjectDecoder); ok {
		glog.Info("Decode Stage B1")
		if err := d.DecodeNestedObjects(DirectDecoder{c.decoder}); err != nil {
			return nil, gvk, err
		}
	}

	glog.Info("Decode Stage C")
	// if we specify a target, use generic conversion.
	if into != nil {
		glog.Info("Decode Stage C1")
		if into == obj {
			// API input decoded directly to our registered type
			glog.Info("Decode Stage C1a")
			if isVersioned {
				return versioned, gvk, nil
			}
			glog.Info("Decode Stage C1b")
			return into, gvk, nil
		}

		glog.Info("Decode Stage C2")
		// API input did not decoded directly to our registered type, so
		// perform defaulting if requested
		if c.defaulter != nil {
			// create a copy to ensure defaulting is not applied to the original versioned objects
			if isVersioned {
				glog.Info("Decode Stage C2a")
				copied, err := c.copier.Copy(obj)
				if err != nil {
					utilruntime.HandleError(err)
					copied = obj
				}
				versioned.Objects = []runtime.Object{copied}
			}
			glog.Info("Decode Stage C2b")
			c.defaulter.Default(obj)
		} else {
			if isVersioned {
				versioned.Objects = []runtime.Object{obj}
			}
		}

		glog.Info("Decode Stage C3")
		if err := c.convertor.Convert(obj, into, c.decodeVersion); err != nil {
			return nil, gvk, err
		}

		glog.Info("Decode Stage C4")
		if isVersioned {
			versioned.Objects = append(versioned.Objects, into)
			return versioned, gvk, nil
		}
		glog.Infof("Decode Stage C5: gvk: %v, into: %v", gvk, into)
		return into, gvk, nil
	}

	glog.Info("Decode Stage D")
	// Convert if needed.
	if isVersioned {
		// create a copy, because ConvertToVersion does not guarantee non-mutation of objects
		copied, err := c.copier.Copy(obj)
		if err != nil {
			utilruntime.HandleError(err)
			copied = obj
		}
		versioned.Objects = []runtime.Object{copied}
	}

	glog.Info("Decode Stage E")
	// perform defaulting if requested
	if c.defaulter != nil {
		glog.Info("Stage: 2 Defaults")
		c.defaulter.Default(obj)
	}

	out, err := c.convertor.ConvertToVersion(obj, c.decodeVersion)
	if err != nil {
		return nil, gvk, err
	}
	glog.Info("Decode Stage F")
	if isVersioned {
		if versioned.Last() != out {
			versioned.Objects = append(versioned.Objects, out)
		}
		return versioned, gvk, nil
	}
	glog.Info("Decode Stage G")
	return out, gvk, nil
}

// Encode ensures the provided object is output in the appropriate group and version, invoking
// conversion if necessary. Unversioned objects (according to the ObjectTyper) are output as is.
func (c *codec) Encode(obj runtime.Object, w io.Writer) error {
	glog.Info("Encode Stage A")
	switch obj.(type) {
	case *runtime.Unknown, runtime.Unstructured:
		return c.encoder.Encode(obj, w)
	}

	gvks, isUnversioned, err := c.typer.ObjectKinds(obj)
	if err != nil {
		return err
	}

	glog.Info("Encode Stage B")
	if c.encodeVersion == nil || isUnversioned {
		glog.Info("Encode Stage B1")
		if e, ok := obj.(runtime.NestedObjectEncoder); ok {
			if err := e.EncodeNestedObjects(DirectEncoder{Encoder: c.encoder, ObjectTyper: c.typer}); err != nil {
				return err
			}
		}
		objectKind := obj.GetObjectKind()
		old := objectKind.GroupVersionKind()
		objectKind.SetGroupVersionKind(gvks[0])
		err = c.encoder.Encode(obj, w)
		objectKind.SetGroupVersionKind(old)
		return err
	}

	glog.Info("Encode Stage C")
	// Perform a conversion if necessary
	objectKind := obj.GetObjectKind()
	old := objectKind.GroupVersionKind()
	out, err := c.convertor.ConvertToVersion(obj, c.encodeVersion)
	if err != nil {
		return err
	}

	glog.Info("Encode Stage D")
	if e, ok := out.(runtime.NestedObjectEncoder); ok {
		glog.Info("Encode Stage D1")
		if err := e.EncodeNestedObjects(DirectEncoder{Version: c.encodeVersion, Encoder: c.encoder, ObjectTyper: c.typer}); err != nil {
			return err
		}
	}

	glog.Info("Encode Stage E")
	// Conversion is responsible for setting the proper group, version, and kind onto the outgoing object
	err = c.encoder.Encode(out, w)
	// restore the old GVK, in case conversion returned the same object
	objectKind.SetGroupVersionKind(old)
	return err
}

// DirectEncoder serializes an object and ensures the GVK is set.
type DirectEncoder struct {
	Version runtime.GroupVersioner
	runtime.Encoder
	runtime.ObjectTyper
}

// Encode does not do conversion. It sets the gvk during serialization.
func (e DirectEncoder) Encode(obj runtime.Object, stream io.Writer) error {
	gvks, _, err := e.ObjectTyper.ObjectKinds(obj)
	if err != nil {
		if runtime.IsNotRegisteredError(err) {
			return e.Encoder.Encode(obj, stream)
		}
		return err
	}
	kind := obj.GetObjectKind()
	oldGVK := kind.GroupVersionKind()
	gvk := gvks[0]
	if e.Version != nil {
		preferredGVK, ok := e.Version.KindForGroupVersionKinds(gvks)
		if ok {
			gvk = preferredGVK
		}
	}
	kind.SetGroupVersionKind(gvk)
	err = e.Encoder.Encode(obj, stream)
	kind.SetGroupVersionKind(oldGVK)
	return err
}

// DirectDecoder clears the group version kind of a deserialized object.
type DirectDecoder struct {
	runtime.Decoder
}

// Decode does not do conversion. It removes the gvk during deserialization.
func (d DirectDecoder) Decode(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	obj, gvk, err := d.Decoder.Decode(data, defaults, into)
	if obj != nil {
		kind := obj.GetObjectKind()
		// clearing the gvk is just a convention of a codec
		kind.SetGroupVersionKind(schema.GroupVersionKind{})
	}
	return obj, gvk, err
}
