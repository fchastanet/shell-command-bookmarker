#!/usr/bin/env bash

# In foreground, continuously run app
mkdir -p _build
while true; do
  DEBUG=1 HISTFILE="${HOME}/.bash_history" _build/shell-command-bookmarker
done
