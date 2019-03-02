#!/bin/sh

die() { echo "oh noes! $*"; exit 1; }

h() {
        if [ "$TRAVIS" = true ]; then
                printf "travis_fold:start:%s\033[33;1m%s\033[0m" "$1" "$2"
        else
                printf "# %s\n" "$2"
        fi
}

f() {
        if [ "$TRAVIS" = true ]; then
                printf "\ntravis_fold:end:%s\r" "$1"
        else
                printf "\n"
        fi
}

if ! [ -r jam.go ]; then
        die "$0 must be run from the repo root"
fi

h test "run tests"
go test -v -covermode=count -coverprofile=c.out . || die "test failed"
f test

h cov "coverage"
go tool cover -func=c.out
if [ "$TRAVIS" = true ] && hash goveralls >/dev/null 2>&1; then
        goveralls -coverprofile=c.out -service=travis-ci || echo "warning: goveralls failed!"
fi
f cov

h init "init"
version="devel"
if hash hubr >/dev/null 2>&1 && hubr now; then
        version=$(head -n 1 VERSION)
fi

cd cmd/jam || die "could not cd to cmd dir"
mkdir -pv dist || die "could not create dist dir"
f init

for os in linux darwin; do
        h "b-$os" "build $os"
        rm -rf jam
        if ! GOOS="$os" go build -ldflags "-X main.version=$version"; then
                die "$os build failed"
        fi
        zip -j "dist/jam-$os.zip" jam || die "$os dist failed"
        f "b-$os"

        h "b-$os-p" "build $os pretty"
        rm -rf jam
        if ! GOOS="$os" go build -ldflags "-X main.version=$version" -tags pretty; then
                die "$os pretty build failed"
        fi
        zip -j "dist/jam-$os-pretty.zip" jam || die "$os pretty dist failed"
        f "b-$os-p"
done

if hash hubr >/dev/null 2>&1; then
        h release release
        hubr push tr-d/jam dist/*.zip || die "failed to push release"
        f release
fi
