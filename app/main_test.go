//go:build sqlite_fts5 || fts5

package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/fchastanet/shell-command-bookmarker/internal/args"
	"github.com/stretchr/testify/assert"
)

func RunMainImpl(shellOption string) (error, string) {
	// Set up a temporary CLI object with bash flag
	oldOsArgs := os.Args
	os.Args = []string{"cmd", shellOption}
	defer func() { os.Args = oldOsArgs }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Run mainImpl
	err := mainImpl()

	// Close the writer and get the output
	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	return err, output
}

// TestMainImplWithBashFlag tests that the mainImpl function correctly outputs
// a Bash script when the --bash flag is provided
func TestMainImplWithBashFlag(t *testing.T) {
	err, output := RunMainImpl("--bash")

	// Check results
	assert.Nil(t, err)
	assert.Contains(t, output, "#!/usr/bin/env bash")
	assert.Contains(t, output, "shell_command_bookmarker_paste()")
}

// TestMainImplWithZshFlag tests that the mainImpl function correctly outputs
// a Zsh script when the --zsh flag is provided
func TestMainImplWithZshFlag(t *testing.T) {
	err, output := RunMainImpl("--zsh")

	// Check results
	assert.Nil(t, err)
	assert.Contains(t, output, "#!/usr/bin/env zsh")
	assert.Contains(t, output, "shell_command_bookmarker_paste()")
}

// TestArgsGenerateFlags tests the CLI argument parsing for the integration flags
func TestArgsGenerateFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		expectZsh  bool
		expectBash bool
	}{
		{
			name:       "zsh flag",
			args:       []string{"cmd", "--zsh"},
			expectZsh:  true,
			expectBash: false,
		},
		{
			name:       "bash flag",
			args:       []string{"cmd", "--bash"},
			expectZsh:  false,
			expectBash: true,
		},
		{
			name:       "no flags",
			args:       []string{"cmd"},
			expectZsh:  false,
			expectBash: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up temporary CLI
			oldArgs := os.Args
			os.Args = tt.args
			defer func() { os.Args = oldArgs }()

			// Parse args
			cli := &args.Cli{}
			err := args.ParseArgs(cli)

			// Check results
			assert.Nil(t, err)
			assert.Equal(t, tt.expectZsh, cli.GenerateZsh)
			assert.Equal(t, tt.expectBash, cli.GenerateBash)
		})
	}
}
