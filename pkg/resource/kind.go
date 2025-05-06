package resource

// Kind interface defines common behavior across all kinds
type Kind interface {
	// Key returns the unique string key for this kind
	Key() string
	IsKind()
}

type DefaultKind struct {
	key string
}

func (k DefaultKind) Key() string { return k.key }
func (DefaultKind) IsKind()       {}
