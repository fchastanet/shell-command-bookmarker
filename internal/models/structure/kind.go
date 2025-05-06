package structure

type KindType struct {
	key string
}

func (k KindType) Key() string {
	return k.key
}
func (KindType) IsKind() {}

//nolint:gochecknoglobals // no way to use enum for this part
var (
	GlobalKind        = KindType{key: "global"}
	CommandKind       = KindType{key: "command"}
	CommandListKind   = KindType{key: "commandList"}
	CommandEditorKind = KindType{key: "commandEditor"}
	TaskKind          = KindType{key: "task"}
	FolderKind        = KindType{key: "folder"}
	SearchKind        = KindType{key: "search"}
)
