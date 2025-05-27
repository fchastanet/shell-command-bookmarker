#!/usr/bin/env bash
set -o pipefail -o errexit

mkdir -pv "${HOME}/go/bin" || true
go env -w "GOBIN=${HOME}/go/bin"

echo >&2 "Check dependencies ..."
go mod download

echo >&2 "Building ..."
go build -tags "sqlite_fts5" -ldflags="-w -s" ./...

echo >&2 "Installing ..."
# Build with a specific output name and move it to GOBIN
go build -tags "sqlite_fts5" -ldflags="-w -s" -o "${HOME}/go/bin/shell-command-bookmarker" ./app

if [[ -f ${HOME}/go/bin/shell-command-bookmarker ]]; then
  echo >&2 "you can run ${HOME}/go/bin/shell-command-bookmarker"
else
  echo >&2 "${HOME}/go/bin/shell-command-bookmarker has not been generated"
  # List available executables to help troubleshoot
  echo >&2 "Available executables in ${HOME}/go/bin:"
  ls -la "${HOME}/go/bin"
fi
