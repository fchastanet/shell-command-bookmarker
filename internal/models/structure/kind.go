package structure

import "github.com/fchastanet/shell-command-bookmarker/pkg/resource"

type KindType struct {
	key string
}

func (k KindType) Key() string {
	return k.key
}
func (k KindType) IsKind() {}

//nolint:gochecknoglobals // no way to use enum for this part
var (
	GlobalKind      = KindType{key: "global"}
	CommandKind     = KindType{key: "command"}
	TaskKind        = KindType{key: "task"}
	FolderKind      = KindType{key: "folder"}
	CommandListKind = KindType{key: "commandList"}
	SearchKind      = KindType{key: "search"}
)

func RegisterKinds() {
	// Register all kinds when the package is imported
	kindRegistry := resource.NewKindRegistry()
	kindRegistry.Register(GlobalKind)
	kindRegistry.Register(CommandKind)
	kindRegistry.Register(TaskKind)
	kindRegistry.Register(FolderKind)
	kindRegistry.Register(CommandListKind)
	kindRegistry.Register(SearchKind)
}
