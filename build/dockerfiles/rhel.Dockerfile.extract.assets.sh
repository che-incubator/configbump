#!/bin/bash
set -xe

# Copyright (c) 2020 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation
#
# script to build rhel.Dockerfile and extract relevant assets for reuse in Brew

TMPIMG=configbump:local
PODMAN=podman; if [[ ! $(which podman) ]]; then PODMAN=docker;fi
${PODMAN} build . -f build/dockerfiles/rhel.Dockerfile -t ${TMPIMG}

# create asset-* files
TMPDIR=$(mktemp -d)
rm -fr ${TMPDIR} ${WORKSPACE}/asset-configbump-*.tar.gz

for d in usr/local/bin/configbump etc/passwd; do
    mkdir -p ${TMPDIR}/${d%/*}
    ${PODMAN} run --rm --entrypoint cat $TMPIMG /${d} > ${TMPDIR}/${d}
done
tree ${TMPDIR}

pushd ${TMPDIR} >/dev/null || exit 1
    tar cvzf "${WORKSPACE}/asset-configbump-$(uname -m).tar.gz" ./
popd >/dev/null || exit 1

${PODMAN} rmi -f ${TMPIMG}
rm -fr ${TMPDIR}
