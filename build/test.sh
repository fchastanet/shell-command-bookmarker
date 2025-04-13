#!/usr/bin/env bash
set -e -o pipefail -o errexit

mkdir -p logs

echo >&2 "Tests whether the code compiles ..."
go build -o /dev/null ./...

echo >&2 "Runs the tests ..."
go test "$@" ./...

go test -count 1 ./... -coverprofile=logs/cover.out --json | tee "logs/tests.log"
