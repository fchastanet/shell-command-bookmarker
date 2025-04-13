#!/usr/bin/env bash
set -e -o pipefail -o errexit

echo >&2 "Runs Go coverage ..."
# Displays coverage per func on cli
go tool cover -func=logs/cover.out

echo >&2 "Runs HTML coverage ..."
## Displays the coverage results in the browser
go tool cover -html=logs/cover.out
