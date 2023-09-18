#!/bin/bash
#
# Copyright (c) 2023 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation
#

# Release process automation script.
# Used to create branch/tag, update VERSION files
# and and trigger release by force pushing changes to the release branch

# set to 1 to actually tag the changes to the release branch
TAG_RELEASE=0
NOCOMMIT=0

while [[ "$#" -gt 0 ]]; do
  case $1 in
    '-t'|'--tag-release') TAG_RELEASE=1; NOCOMMIT=0; shift 0;;
    '-v'|'--version') VERSION="$2"; shift 1;;
    '-n'|'--no-commit') NOCOMMIT=1; TAG_RELEASE=0; shift 0;;
  esac
  shift 1
done

sed_in_place() {
    SHORT_UNAME=$(uname -s)
  if [ "$(uname)" == "Darwin" ]; then
    sed -i '' "$@"
  elif [ "${SHORT_UNAME:0:5}" == "Linux" ]; then
    sed -i "$@"
  fi
}

# search for occurrences of the old version in VERSION file and update to the new version
function update_versioned_files() {
  local VER=$1
  OLD_VER=$(cat VERSION); OLD_VER=${OLD_VER%-*}
  for file in README.md cmd/configbump/main.go; do
    sed_in_place -r -e "s@${OLD_VER}@${VER}@g" $file
  done
  echo "${VER}" > VERSION
}

bump_version () {
  CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

  NEXT_VERSION=$1
  BUMP_BRANCH=$2

  git checkout "${BUMP_BRANCH}"

  echo "Updating project version to ${NEXT_VERSION}"
  update_versioned_files $NEXT_VERSION

  if [[ ${NOCOMMIT} -eq 0 ]]; then
    COMMIT_MSG="chore: release: bump to ${NEXT_VERSION} in ${BUMP_BRANCH}"
    git commit -asm "${COMMIT_MSG}"
    git pull origin "${BUMP_BRANCH}"

    set +e
    PUSH_TRY="$(git push origin "${BUMP_BRANCH}")"
    # shellcheck disable=SC2181
    if [[ $? -gt 0 ]] || [[ $PUSH_TRY == *"protected branch hook declined"* ]]; then
        # create pull request for main branch, as branch is restricted
        PR_BRANCH=pr-${BUMP_BRANCH}-to-${NEXT_VERSION}
        git branch "${PR_BRANCH}"
        git checkout "${PR_BRANCH}"
        git pull origin "${PR_BRANCH}"
        git push origin "${PR_BRANCH}"
        lastCommitComment="$(git log -1 --pretty=%B)"
        hub pull-request -f -m "${lastCommitComment}" -b "${BUMP_BRANCH}" -h "${PR_BRANCH}"
    fi
    set -e
  fi
  git checkout "${CURRENT_BRANCH}"
}

usage ()
{
  echo "Usage: $0 --version [VERSION TO RELEASE] [--tag-release]"
  echo "Example: $0 --version 7.75.0 --tag-release"; echo
}

if [[ ! ${VERSION} ]]; then
  usage
  exit 1
fi

# derive branch from version
BRANCH=${VERSION%.*}.x

# if doing a .0 release, use main; if doing a .z release, use $BRANCH
if [[ ${VERSION} == *".0" ]]; then
  BASEBRANCH="main"
else
  BASEBRANCH="${BRANCH}"
fi

# get sources from ${BASEBRANCH} branch
git fetch origin "${BASEBRANCH}":"${BASEBRANCH}"
git checkout "${BASEBRANCH}"

# create new branch off ${BASEBRANCH} (or check out latest commits if branch already exists), then push to origin
if [[ "${BASEBRANCH}" != "${BRANCH}" ]]; then
  git branch "${BRANCH}" || git checkout "${BRANCH}" && git pull origin "${BRANCH}"
  git push origin "${BRANCH}"
  git fetch origin "${BRANCH}:${BRANCH}"
  git checkout "${BRANCH}"
fi

set -e
set -x

# change VERSION file
echo "${VERSION}" > VERSION

# commit change into branch
if [[ ${NOCOMMIT} -eq 0 ]]; then
  COMMIT_MSG="chore: release: bump to ${VERSION} in ${BRANCH}"
  git status -s || true
  git commit -asm "${COMMIT_MSG}" || git diff || true
  git pull origin "${BRANCH}"
  git push origin "${BRANCH}"
fi

if [[ $TAG_RELEASE -eq 1 ]]; then
  # tag the release
  git checkout "${BRANCH}"
  git tag "${VERSION}"
  git push origin "${VERSION}"
fi

# now update ${BASEBRANCH} to the new version
git checkout "${BASEBRANCH}"

# change VERSION file + commit change into ${BASEBRANCH} branch
if [[ "${BASEBRANCH}" != "${BRANCH}" ]]; then
  # bump the y digit, if it is a major release
  [[ $BRANCH =~ ^([0-9]+)\.([0-9]+)\.x ]] && BASE=${BASH_REMATCH[1]}; NEXT=${BASH_REMATCH[2]}; (( NEXT=NEXT+1 )) # for BRANCH=7.10.x, get BASE=7, NEXT=11
  NEXT_VERSION_Y="${BASE}.${NEXT}.0"
  bump_version "${NEXT_VERSION_Y}" "${BASEBRANCH}"
fi
# bump the z digit
[[ $VERSION =~ ^([0-9]+)\.([0-9]+)\.([0-9]+) ]] && BASE="${BASH_REMATCH[1]}.${BASH_REMATCH[2]}"; NEXT="${BASH_REMATCH[3]}"; (( NEXT=NEXT+1 )) # for VERSION=7.7.1, get BASE=7.7, NEXT=2
NEXT_VERSION_Z="${BASE}.${NEXT}"
bump_version "${NEXT_VERSION_Z}" "${BRANCH}"
