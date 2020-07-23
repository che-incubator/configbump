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

# UPSTREAM: use devtools/go/-toolset-rhel7 image so we're not required to authenticate with registry.redhat.io
# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/devtools/go-toolset-rhel7
FROM registry.access.redhat.com/devtools/go-toolset-rhel7:1.13.4-18 as builder
ARG CGO_ENABLED=0
ARG GOOS=linux
ENV PATH=/opt/rh/go-toolset-1.13/root/usr/bin:$PATH
# DOWNSTREAM: use rhel8/go-toolset; no path modification needed
# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/rhel8/go-toolset
# FROM registry.redhat.io/rhel8/go-toolset:1.13.4-15 as builder

ENV GOPATH=/go/
USER root
WORKDIR /go/src/github.com/che-incubator/configbump
# copy go.mo go.sum
COPY go.mod go.sum ./
# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download && go mod verify
COPY . /go/src/github.com/che-incubator/configbump
RUN go test -v  ./...
RUN adduser appuser && \
    go build -a -ldflags '-w -s' -a -installsuffix cgo -o configbump cmd/configbump/main.go


# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/ubi8-minimal
FROM registry.access.redhat.com/ubi8-minimal:8.2-267
USER appuser
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/che-incubator/configbump/configbump /usr/local/bin
ENTRYPOINT ["configbump"]

# append Brew metadata here
