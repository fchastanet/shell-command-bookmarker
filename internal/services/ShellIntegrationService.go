package services

import (
	_ "embed"
)

//go:embed templates/bash-integration.sh
var bashIntegrationTemplate string

//go:embed templates/zsh-integration.zsh
var zshIntegrationTemplate string

// ShellIntegrationService provides shell integration scripts
type ShellIntegrationService struct{}

// NewShellIntegrationService creates a new instance of ShellIntegrationService
func NewShellIntegrationService() *ShellIntegrationService {
	return &ShellIntegrationService{}
}

// GenerateBashIntegration returns a bash script for shell integration
func (s *ShellIntegrationService) GenerateBashIntegration() string {
	return bashIntegrationTemplate
}

// GenerateZshIntegration returns a zsh script for shell integration
func (s *ShellIntegrationService) GenerateZshIntegration() string {
	return zshIntegrationTemplate
}
