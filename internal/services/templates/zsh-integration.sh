#!/usr/bin/env zsh

# Shell Command Bookmarker - Zsh Integration
# This function allows you to select a command and paste it directly at your prompt

shell_command_bookmarker_paste() {
  local tmp_file
  tmp_file=$(mktemp)
  trap 'rm -f "${tmp_file}"' EXIT # Ensure cleanup on exit

  # Run the application with output redirection
  shell-command-bookmarker --output-file="${tmp_file}"

  # Check if command was selected
  if [ -s "${tmp_file}" ]; then
    BUFFER=$(cat "${tmp_file}")
    CURSOR=${#BUFFER}
  else
    BUFFER="" # Clear the buffer if no command was selected
    CURSOR=0  # Reset cursor position
  fi
  zle reset-prompt
}

# Register widget and bind to Ctrl+G (you can change this to your preference)
zle -N shell_command_bookmarker_paste
bindkey '^g' shell_command_bookmarker_paste

echo "Shell Command Bookmarker zsh integration loaded."
echo "Press Ctrl+G or type 'bookmark' to insert a saved command."
