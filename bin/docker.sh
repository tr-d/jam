#!/bin/sh

die() { echo "oh noes! $*"; exit 1; }

cmd=$1
if [ -z "$cmd" ]; then
        echo "usage: $0 <cmd>"
        exit 2
fi

if ! [ -r jam.go ]; then
        die "$0 must be run from the repo root"
fi

docker pull dedelala/go:latest || die "failed to pull"

docker run --rm \
        -e "TRAVIS_JOB_ID=$TRAVIS_JOB_ID" \
        -e "TRAVIS_BRANCH=$TRAVIS_BRANCH" \
        -e "GITHUB_API_TOKEN=$GITHUB_API_TOKEN" \
        -v "$PWD:/jam" -w /jam dedelala/go:latest "$cmd" || die "$cmd failed"
