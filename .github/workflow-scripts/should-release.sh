#!/usr/bin/env bash

VERSION=$1

! curl -s ${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/git/matching-refs/tags | jq -e --arg VERSION "$VERSION" '.[] | select(.ref | endswith($VERSION))'

if [ $? -eq 0 ]; then
   echo "::set-output name=release::true"
else
   echo "::set-output name=release::false"
fi