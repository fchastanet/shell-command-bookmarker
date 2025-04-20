package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

const maxScreenSize = 80

type cli struct {
	MaxTasks int         `short:"t" name:"max-tasks" default:"1"             help:"Maximum number of tasks to run concurrently"` //nolint:tagalign //avoid reformat annotations
	DBPath   FilePath    `arg:""    name:"db-path"   optional:"" type:"path" help:"Path to the SQLite database file"`            //nolint:tagalign //avoid reformat annotations
	Version  VersionFlag `short:"v" name:"version"                           help:"Print version information and quit"`          //nolint:tagalign //avoid reformat annotations
	Debug    bool        `short:"d"                                          help:"Set log in debug level"`                      //nolint:tagalign //avoid reformat annotations
}

type FilePath string

func (f *FilePath) Decode(_ *kong.DecodeContext) error {
	if *f == "" {
		return fmt.Errorf("missing required argument: db-path")
	}
	if _, err := os.Stat(string(*f)); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", *f)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied to access file: %s", *f)
		}
		return fmt.Errorf("error accessing file: %s", *f)
	}
	return nil
}

type (
	VersionFlag string
)

func (VersionFlag) Decode(_ *kong.DecodeContext) error { return nil }
func (VersionFlag) IsBool() bool                       { return true }
func (VersionFlag) BeforeApply(
	app *kong.Kong, vars kong.Vars,
) error { //nolint:unparam // need to conform to interface
	fmt.Printf("Bash command bookmarker version %s\n", vars["version"])
	app.Exit(0)
	return nil
}

func parseArgs(cli *cli) (err error) {
	// just need the yaml file, from which all the dependencies will be deduced
	kong.Parse(cli,
		kong.Name("shell-command-bookmarker"),
		kong.Description("A command line tool to bookmark shell commands"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppSummary:        false,
			Summary:             true,
			Compact:             true,
			Tree:                false,
			FlagsLast:           true,
			Indenter:            kong.LineIndenter,
			NoExpandSubcommands: true,
			WrapUpperBound:      maxScreenSize,
		}),
		kong.Vars{
			"version": "1.0.0",
		},
	)

	if cli.DBPath == "" {
		cli.DBPath = "db/shell-command-bookmarker.db"
		if os.Getenv("SHELL_CMD_BOOK_DB") != "" {
			cli.DBPath = FilePath(os.Getenv("SHELL_CMD_BOOK_DB"))
		}
	}

	return nil
}
