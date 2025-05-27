#!/usr/bin/env bash
set -e -o pipefail -o errexit

echo "Cleaning ..."
rm -rvf bin logs || true
go mod tidy || true
docker image rm -f scrasnups/shell-command-bookmarker || true
