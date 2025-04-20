package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func defaultCli() *cli {
	return &cli{
		MaxTasks: 1,
		DBPath:   FilePath("db/shell-command-bookmarker.db"),
		Version:  VersionFlag(""),
		Debug:    false,
	}
}

func defaultCase(t *testing.T, args []string) {
	os.Args = args
	expectedCli := defaultCli()
	cli := &cli{} //nolint:exhaustruct //test
	err := parseArgs(cli)
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
		cli := &cli{} //nolint:exhaustruct //test
		err := parseArgs(cli)
		assert.Nil(t, err)
		assert.Equal(t, expectedCli, cli)
	})

	t.Run("custom max-tasks", func(t *testing.T) {
		expectedCli := defaultCli()
		expectedCli.MaxTasks = 5
		os.Args = []string{"cmd", "-t", "5"}
		cli := &cli{} //nolint:exhaustruct //test
		err := parseArgs(cli)
		assert.Nil(t, err)
		assert.Equal(t, expectedCli, cli)
	})

}
