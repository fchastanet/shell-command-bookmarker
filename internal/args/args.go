package args

import (
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

const maxScreenSize = 80

type Cli struct {
	DBPath   FilePath    `arg:""    name:"db-path"   optional:"" type:"path" help:"Path to the SQLite database file"`            //nolint:tagalign //avoid reformat annotations
	Version  VersionFlag `short:"v" name:"version"                           help:"Print version information and quit"`          //nolint:tagalign //avoid reformat annotations
	MaxTasks int         `short:"t" name:"max-tasks" default:"1"             help:"Maximum number of tasks to run concurrently"` //nolint:tagalign //avoid reformat annotations
	Debug    bool        `short:"d"                                          help:"Set log in debug level"`                      //nolint:tagalign //avoid reformat annotations
}

type FilePath string

func (f *FilePath) Decode(_ *kong.DecodeContext) error {
	if *f == "" {
		return ErrMissingDBPath
	}

	// Check if file exists
	if _, err := os.Stat(string(*f)); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrFileDoesNotExist, *f)
	}

	// Check if permission is granted
	if _, err := os.Stat(string(*f)); errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("%w: %s", ErrPermissionDenied, *f)
	}

	if _, err := os.Stat(string(*f)); err != nil {
		return fmt.Errorf("%w: %s", ErrAccessingFile, *f)
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
) error {
	fmt.Printf("Bash command bookmarker version %s\n", vars["version"])
	app.Exit(0)
	return nil
}

func ParseArgs(cli *Cli) (err error) {
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
