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

# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/ubi8-minimal
FROM registry.access.redhat.com/ubi8-minimal:8.4-200.1622548483 as builder
RUN microdnf install -y golang

ENV GOPATH=/go/ \
    CGO_ENABLED=0 \
    GOOS=linux
USER root
WORKDIR /go/src/github.com/che-incubator/configbump
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . /go/src/github.com/che-incubator/configbump
RUN adduser appuser && \
    go test -v  ./... && \
    export ARCH="$(uname -m)" && if [[ ${ARCH} == "x86_64" ]]; then export ARCH="amd64"; elif [[ ${ARCH} == "aarch64" ]]; then export ARCH="arm64"; fi && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -a -ldflags '-w -s' -a -installsuffix cgo -o configbump cmd/configbump/main.go && \
    cp /go/src/github.com/che-incubator/configbump/configbump /usr/local/bin/configbump

# now collect assets into a tarball, and in brew.Dockerfile, extract and use them
# if doing steps locally, run ./build/dockerfiles/rhel.Dockerfile.extract.assets.sh
# if running in Jenkins, see https://gitlab.cee.redhat.com/codeready-workspaces/crw-jenkins/-/blob/master/jobs/CRW_CI/crw-configbump_2.6.jenkinsfile (or newer) for script
