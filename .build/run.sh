#!/usr/bin/env bash
set -e -o pipefail -o errexit

go run -tags "sqlite_fts5" ./main
