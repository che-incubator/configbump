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

# this container build continues from rhel.Dockerfile and rhel.Dockefile.extract.assets.sh
# assumes you have created asset-*.tar.gz files for all arches, but you'll only unpack the one for your arch

# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/ubi8-minimal
FROM ubi8-minimal:8.2-349
USER appuser
COPY asset-*.tar.gz /tmp/assets/
RUN tar xzf /tmp/asset/asset-configbump-$(uname -m).tar.gz -C / && \
    rm -fr /tmp/assets/
ENTRYPOINT ["configbump"]

# append Brew metadata here
