#!/usr/bin/env bash

# Watch code changes, trigger re-build, and kill process
while true; do
  echo "Watching for changes..."
  rm logs/*.log
  go build -o bin/shell-command-bookmarker -tags=sqlite_fts5 ./app/main.go && pkill -f '_build/shell-command-bookmarker'
  mapfile -t files < <(find . -name '*.go')
  # Watch for modify, create, and move events on Go files (not just attribute changes)
  inotifywait -e modify -e create -e move "${files[@]}" || exit
done
