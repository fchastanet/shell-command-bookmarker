#!/usr/bin/env bash
set -e -o pipefail -o errexit

go run -tags "fts5" ./main
