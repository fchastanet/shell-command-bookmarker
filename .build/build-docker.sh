#!/usr/bin/env bash
set -e -o pipefail -o errexit

declare image="scrasnups/shell-command-bookmarker"
mkdir -pv logs bin

docker buildx build -t "${image}" .
