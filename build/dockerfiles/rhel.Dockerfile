# Copyright (c) 2019 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation
#

# UPSTREAM: use devtools/go-toolset-rhel7 image so we're not required to authenticate with registry.redhat.io
# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/rhel8/go-toolset
FROM registry.redhat.io/rhel8/go-toolset:1.13.15-1 as builder
USER root
ENV PATH=/opt/rh/go-toolset-1.13/root/usr/bin:$PATH \
    GOPATH=/go/ \
    CGO_ENABLED=0 \
    GOOS=linux
WORKDIR /go/src/github.com/che-incubator/configbump
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . /go/src/github.com/che-incubator/configbump
RUN adduser appuser && \
    go test -v  ./... && \
    export ARCH="$(uname -m)" && if [[ ${ARCH} == "x86_64" ]]; then export ARCH="amd64"; elif [[ ${ARCH} == "aarch64" ]]; then export ARCH="arm64"; fi && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -a -ldflags '-w -s' -a -installsuffix cgo -o configbump cmd/configbump/main.go

# now collect assets into a tarball, and in brew.Dockerfile, extract and use them
# if doing steps locally, run ./build/dockerfiles/rhel.Dockerfile.extract.assets.sh
# if running in Jenkins, see https://github.com/redhat-developer/codeready-workspaces/tree/crw-2.5-rhel-8/dependencies/configbump/Jenkinsfile (or newer branch) for script
