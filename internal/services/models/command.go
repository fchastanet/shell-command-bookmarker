package models

import (
	"time"

	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
)

type CommandStatus string

const (
	CommandStatusImported   CommandStatus = "IMPORTED"
	CommandStatusBookmarked CommandStatus = "BOOKMARKED"
)

type Command struct {
	ID                   uint
	Title                string
	Description          string
	Script               string
	Status               CommandStatus
	LintIssues           string
	LintStatus           LintStatus
	Elapsed              int
	CreationDatetime     time.Time
	ModificationDatetime time.Time
}

type CommandStatusEnum struct {
	Imported   CommandStatus
	Bookmarked CommandStatus
}

type LintStatus string

const (
	LintStatusNotAvailable     LintStatus = "NOT_AVAILABLE"
	LintStatusOK               LintStatus = "OK"
	LintStatusWarning          LintStatus = "WARNING"
	LintStatusError            LintStatus = "ERROR"
	LintStatusShellcheckFailed LintStatus = "SHELLCHECK_FAILED"
)

func (c *Command) GetID() resource.ID {
	return resource.ID(c.ID)
}

func CommandSorter(i, j *Command) int {
	switch {
	case i.ID < j.ID:
		return -1
	case i.ID > j.ID:
		return 1
	default:
		return 0
	}
}
