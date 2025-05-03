package resource

// ID uniquely identifies a Shell Command Bookmarker resource.
type ID int64

// Identifiable is a Shell Command Bookmarker resource with an identity.
type Identifiable interface {
	GetID() ID
}
