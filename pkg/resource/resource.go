package resource

// Resource is a unique Shell Command Bookmarker entity
type Resource interface {
	GetID() ID
	// String is a human-readable identifier for the resource. Not necessarily
	// unique across Shell Command Bookmarker.
	String() string
}
