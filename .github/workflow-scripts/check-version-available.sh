#!/usr/bin/env bash

VERSION=$1

curl  ${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/git/matching-refs/tags | jq -e --arg VERSION "$VERSION" '.[] | select(.ref | endswith($VERSION))'