#!/bin/sh

die() { echo "oh noes! $*"; exit 1; }

h() {
        if [ "$TRAVIS" = true ]; then
                printf "travis_fold:start:%s\033[33;1m%s\033[0m\n" "$1" "$2"
        else
                printf "# %s\n" "$2"
        fi
}

f() {
        if [ "$TRAVIS" = true ]; then
                printf "\ntravis_fold:end:%s\n" "$1"
        else
                printf "\n"
        fi
}

cmd=$1
if [ -z "$cmd" ]; then
        echo "usage: $0 <cmd>"
        exit 2
fi

if ! [ -r jam.go ]; then
        die "$0 must be run from the repo root"
fi

h d-pull "docker pull dedelala/go"
docker pull dedelala/go:latest || die "failed to pull"
f d-pull

docker run --rm \
        -e "TRAVIS=$TRAVIS" \
        -e "TRAVIS_JOB_ID=$TRAVIS_JOB_ID" \
        -e "TRAVIS_BRANCH=$TRAVIS_BRANCH" \
        -e "GITHUB_API_TOKEN=$GITHUB_API_TOKEN" \
        -v "$PWD:/jam" -w /jam dedelala/go:latest "$cmd" || die "$cmd failed"
