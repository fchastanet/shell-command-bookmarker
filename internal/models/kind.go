package models

import "github.com/fchastanet/shell-command-bookmarker/pkg/tui"

type Kind tui.Kind

const (
	CommandKind Kind = iota
	CommandListKind
	FolderKind
	TaskKind
)

func (i Kind) String() string {
	switch i {
	case CommandKind:
		return "Command"
	case CommandListKind:
		return "CommandList"
	case FolderKind:
		return "Folder"
	case TaskKind:
		return "Task"
	}
	return "Unknown"
}
