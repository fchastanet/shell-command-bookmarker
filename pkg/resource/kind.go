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
func (k DefaultKind) IsKind()     {}

// KindRegistry defines operations for registering and checking kinds
type KindRegistry interface {
	Register(kind Kind)
	IsRegistered(kind Kind) bool
}

// kindRegistryImpl is the concrete implementation of KindRegistry
type kindRegistryImpl struct {
	kinds map[string]bool
}

// NewKindRegistry creates a new kind registry
func NewKindRegistry() KindRegistry {
	return &kindRegistryImpl{
		kinds: make(map[string]bool),
	}
}

// Register adds a kind to the registry
func (r *kindRegistryImpl) Register(kind Kind) {
	r.kinds[kind.Key()] = true
}

// IsRegistered checks if a kind is registered
func (r *kindRegistryImpl) IsRegistered(kind Kind) bool {
	_, exists := r.kinds[kind.Key()]
	return exists
}

// RegisterDefaultKinds registers the default kinds into the provided registry
func RegisterDefaultKinds(registry KindRegistry) {
	registry.Register(DefaultKind{key: "default"})
}
