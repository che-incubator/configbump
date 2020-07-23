#
# Copyright (c) 2012-2018 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation
#

FROM golang:1.13-alpine3.12 as builder
ARG CGO_ENABLED=0
ARG GOOS=linux
RUN apk add --no-cache ca-certificates
RUN adduser -D -g '' appuser
WORKDIR /go/src/github.com/che-incubator/configbump
# copy go.mo go.sum
COPY go.mod go.sum ./
# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download && go mod verify
COPY . /go/src/github.com/che-incubator/configbump
RUN go test -v  ./...
RUN go build -a -ldflags '-w -s' -a -installsuffix cgo -o configbump cmd/configbump/main.go

FROM alpine:3.12
USER appuser
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/che-incubator/configbump/configbump /usr/local/bin
ENTRYPOINT ["configbump"]
