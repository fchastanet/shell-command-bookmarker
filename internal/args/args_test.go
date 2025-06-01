package args

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func defaultCli() *Cli {
	return &Cli{
		MaxTasks:     1,
		DBPath:       "db/shell-command-bookmarker.db",
		Version:      "",
		Debug:        false,
		OutputFile:   "",
		GenerateZsh:  false,
		GenerateBash: false,
		AutoDetect:   false,
	}
}

func defaultCase(t *testing.T, args []string) {
	os.Args = args
	expectedCli := defaultCli()
	cli := &Cli{} //nolint:exhaustruct //test
	err := ParseArgs(cli)
	assert.Nil(t, err)
	assert.Equal(t, expectedCli, cli)
}

func TestArgs(t *testing.T) {
	t.Run("no arg", func(t *testing.T) {
		defaultCase(t, []string{"cmd"})
	})

	t.Run("debug", func(t *testing.T) {
		expectedCli := defaultCli()
		expectedCli.Debug = true
		os.Args = []string{"cmd", "-d"}
		cli := &Cli{} //nolint:exhaustruct //test
		err := ParseArgs(cli)
		assert.Nil(t, err)
		assert.Equal(t, expectedCli, cli)
	})

	t.Run("custom max-tasks", func(t *testing.T) {
		expectedCli := defaultCli()
		expectedCli.MaxTasks = 5
		os.Args = []string{"cmd", "-t", "5"}
		cli := &Cli{} //nolint:exhaustruct //test
		err := ParseArgs(cli)
		assert.Nil(t, err)
		assert.Equal(t, expectedCli, cli)
	})

	t.Run("output-file", func(t *testing.T) {
		expectedCli := defaultCli()
		expectedCli.OutputFile = "/tmp/output.txt"
		os.Args = []string{"cmd", "--output-file", "/tmp/output.txt"}
		cli := &Cli{} //nolint:exhaustruct //test
		err := ParseArgs(cli)
		assert.Nil(t, err)
		assert.Equal(t, expectedCli, cli)
	})

	t.Run("zsh flag", func(t *testing.T) {
		expectedCli := defaultCli()
		expectedCli.GenerateZsh = true
		os.Args = []string{"cmd", "--zsh"}
		cli := &Cli{} //nolint:exhaustruct //test
		err := ParseArgs(cli)
		assert.Nil(t, err)
		assert.Equal(t, expectedCli, cli)
	})

	t.Run("bash flag", func(t *testing.T) {
		expectedCli := defaultCli()
		expectedCli.GenerateBash = true
		os.Args = []string{"cmd", "--bash"}
		cli := &Cli{} //nolint:exhaustruct //test
		err := ParseArgs(cli)
		assert.Nil(t, err)
		assert.Equal(t, expectedCli, cli)
	})

	t.Run("auto flag", func(t *testing.T) {
		expectedCli := defaultCli()
		expectedCli.AutoDetect = true
		os.Args = []string{"cmd", "--auto"}
		cli := &Cli{} //nolint:exhaustruct //test
		err := ParseArgs(cli)
		assert.Nil(t, err)
		assert.Equal(t, expectedCli, cli)
	})

	t.Run("short auto flag", func(t *testing.T) {
		expectedCli := defaultCli()
		expectedCli.AutoDetect = true
		os.Args = []string{"cmd", "-a"}
		cli := &Cli{} //nolint:exhaustruct //test
		err := ParseArgs(cli)
		assert.Nil(t, err)
		assert.Equal(t, expectedCli, cli)
	})
}

// TestArgsGenerateFlags tests the CLI argument parsing for the integration flags
func TestArgsGenerateFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		expectZsh  bool
		expectBash bool
		expectAuto bool
	}{
		{
			name:       "zsh flag",
			args:       []string{"cmd", "--zsh"},
			expectZsh:  true,
			expectBash: false,
			expectAuto: false,
		},
		{
			name:       "bash flag",
			args:       []string{"cmd", "--bash"},
			expectZsh:  false,
			expectBash: true,
			expectAuto: false,
		},
		{
			name:       "auto flag",
			args:       []string{"cmd", "--auto"},
			expectZsh:  false,
			expectBash: false,
			expectAuto: true,
		},
		{
			name:       "short auto flag",
			args:       []string{"cmd", "-a"},
			expectZsh:  false,
			expectBash: false,
			expectAuto: true,
		},
		{
			name:       "no flags",
			args:       []string{"cmd"},
			expectZsh:  false,
			expectBash: false,
			expectAuto: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up temporary CLI
			oldArgs := os.Args
			os.Args = tt.args
			defer func() { os.Args = oldArgs }()

			// Parse args
			cli := &Cli{} //nolint:exhaustruct //test
			err := ParseArgs(cli)

			// Check results
			assert.Nil(t, err)
			assert.Equal(t, tt.expectZsh, cli.GenerateZsh)
			assert.Equal(t, tt.expectBash, cli.GenerateBash)
			assert.Equal(t, tt.expectAuto, cli.AutoDetect)
		})
	}
}
