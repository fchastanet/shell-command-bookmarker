#!/usr/bin/env bash

set -e -o pipefail -o errexit
echo >&2 "Runs deadcode ..."

tempFile=$(mktemp)
#trap 'rm -f "${tempFile}"' EXIT
echo "${tempFile}"
# run deadcode
deadcode \
  -filter "github.com/fchastanet/shell-command-bookmarker" \
  ./app/main.go >"${tempFile}"

# remove the ignored lines
sed -i -E \
  -e '/^pkg\/tui\/table\/table.go:.*: unreachable func: WithNavigation$/d' \
  -e '/^pkg\/tui\/table\/table.go:.*: unreachable func: WithAction$/d' \
  -e '/^pkg\/tui\/table\/table.go:.*: unreachable func: WithSelectable$/d' \
  -e '/^pkg\/tui\/table\/truncation.go:.*: unreachable func: TruncateLeft$/d' \
  "${tempFile}"

if [[ -s "${tempFile}" ]]; then
  echo >&2 "Dead code found:"
  cat "${tempFile}"
  exit 1
else
  echo >&2 "No dead code found."
fi
