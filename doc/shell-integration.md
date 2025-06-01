# Shell Integration for Command Bookmarker

This document explains how to integrate Shell Command Bookmarker with Bash or
Zsh for easy command pasting.

- [1. Why Use Shell Integration?](#1-why-use-shell-integration)
- [2. Installation](#2-installation)
  - [2.1. Automatic Shell Detection](#21-automatic-shell-detection)
  - [2.2. Bash Integration](#22-bash-integration)
  - [2.3. Zsh Integration](#23-zsh-integration)
- [3. Usage](#3-usage)
- [4. Manual Integration](#4-manual-integration)
- [5. Customizing the Integration](#5-customizing-the-integration)
- [6. Troubleshooting](#6-troubleshooting)

## 1. Why Use Shell Integration?

Shell integration allows you to:

- Select a command from your saved bookmarks
- Insert it directly at your shell prompt
- Execute it immediately or edit it first

## 2. Installation

### 2.1. Automatic Shell Detection

The simplest way to set up shell integration is to use the `--auto` (or `-a`)
flag, which will automatically detect whether you're using Bash or Zsh and
generate the appropriate integration script:

```bash
(
  echo '# Shell Command Bookmarker - Auto Shell Integration'
  echo 'source <(shell-command-bookmarker --auto)'
) >>~/.bashrc # or ~/.zshrc depending on your shell
```

Restart your shell or run:

```bash
source ~/.bashrc # or ~/.zshrc depending on your shell
```

### 2.2. Bash Integration

Update your shell configuration to include the command bookmarker functionality:

```bash
(
  echo '# Shell Command Bookmarker - Bash Integration'
  echo 'source <(shell-command-bookmarker --bash)'
) >~/.bashrc
```

Restart your shell or run:

```bash
source ~/.bashrc
```

### 2.3. Zsh Integration

Update your shell configuration to include the command bookmarker functionality:

```bash
(
  echo '# Shell Command Bookmarker - Zsh Integration'
  echo 'source <(shell-command-bookmarker --zsh)'
) >~/.zshrc
```

Restart your shell or run:

```bash
source ~/.zshrc
```

## 3. Usage

Once installed, you can:

1. Press `Ctrl+G` to bring up the command bookmarker interface
2. Navigate to select your command
3. Press `Enter` to select it
4. The command will be inserted at your prompt, ready to execute

You can also use the `bookmark` alias to achieve the same functionality.

## 4. Manual Integration

If you prefer to integrate without using the generated scripts, you can:

Create a function in your shell configuration that runs:

```bash
shell-command-bookmarker --output-file=/tmp/cmd.txt
```

Read the file and insert its contents at your prompt

Clean up the temporary file

## 5. Customizing the Integration

You can modify the key binding in the integration file:

- For Bash: Change `bind -x '"\C-g": shell_command_bookmarker_paste'` to use a
  different key
- For Zsh: Change `bindkey '^g' shell_command_bookmarker_paste` to use a
  different key

## 6. Troubleshooting

If you encounter issues:

1. Make sure `shell-command-bookmarker` is in your PATH
2. Check that you have write permissions to the temporary directory
3. Verify that the integration script was sourced correctly
