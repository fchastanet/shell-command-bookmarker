package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellIntegrationService_GenerateBashIntegration(t *testing.T) {
	service := NewShellIntegrationService()
	script := service.GenerateBashIntegration()

	// Check essential components of the script
	assert.Contains(t, script, "#!/usr/bin/env bash")
	assert.Contains(t, script, "shell_command_bookmarker_paste()")
	assert.Contains(t, script, "--output-file=")
	assert.Contains(t, script, "READLINE_LINE=")
	assert.Contains(t, script, "bind -x '\"\\C-g\": shell_command_bookmarker_paste'")
	assert.Contains(t, script, "alias bookmark='shell_command_bookmarker_paste'")
}

func TestShellIntegrationService_GenerateZshIntegration(t *testing.T) {
	service := NewShellIntegrationService()
	script := service.GenerateZshIntegration()

	// Check essential components of the script
	assert.Contains(t, script, "#!/usr/bin/env zsh")
	assert.Contains(t, script, "shell_command_bookmarker_paste()")
	assert.Contains(t, script, "--output-file=")
	assert.Contains(t, script, "BUFFER=$(cat")
	assert.Contains(t, script, "zle -N shell_command_bookmarker_paste")
	assert.Contains(t, script, "bindkey '^g' shell_command_bookmarker_paste")
}
