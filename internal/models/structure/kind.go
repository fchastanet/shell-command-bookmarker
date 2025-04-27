package structure

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
