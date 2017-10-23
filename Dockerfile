# Copyright (c) 2016-2017 CloudPerceptions, LLC. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
FROM golang:1.7.5
MAINTAINER CloudPerceptions <support@cloudperceptions.com>

RUN go get -d -v github.com/emicklei/go-restful && \
    go get -d -v github.com/ghodss/yaml && \
    go get -d -v github.com/golang/glog && \
    go get -d -v github.com/go-openapi/spec && \
    go get -d -v github.com/gophercloud/gophercloud && \
    go get -d -v github.com/gophercloud/gophercloud/openstack && \
    go get -d -v github.com/pborman/uuid && \
    go get -d -v github.com/pkg/errors && \
    go get -d -v github.com/spf13/pflag && \
    go get -d -v github.com/ugorji/go/codec && \
    go get -d -v golang.org/x/net/context && \
    go get -d -v golang.org/x/net/http2 && \
    go get -d -v bitbucket.org/ww/goautoneg
