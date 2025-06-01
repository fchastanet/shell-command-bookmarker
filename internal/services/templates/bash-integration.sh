#!/usr/bin/env bash

# Shell Command Bookmarker - Bash Integration
# This function allows you to select a command and paste it directly at your prompt

shell_command_bookmarker_paste() {
  local tmp_file
  tmp_file=$(mktemp)
  trap 'rm -f "${tmp_file}"' EXIT # Ensure cleanup on exit

  # Run the application with output redirection
  shell-command-bookmarker --output-file="${tmp_file}"

  # Check if command was selected and file exists
  if [[ -s "${tmp_file}" ]]; then
    # Insert the command at current cursor position
    READLINE_LINE="$(cat "${tmp_file}")"
    READLINE_POINT=${#READLINE_LINE}
  fi
}

# Bind to Ctrl+G (you can change this to your preference)
bind -x '"\C-g": shell_command_bookmarker_paste'

# Add alias for convenience
alias bookmark='shell_command_bookmarker_paste'

echo "Shell Command Bookmarker bash integration loaded."
echo "Press Ctrl+G or type 'bookmark' to insert a saved command."
