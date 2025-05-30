package resource

const (
	CreatedEvent EventType = "created"
	UpdatedEvent EventType = "updated"
	DeletedEvent EventType = "deleted"
)

type (
	// EventType identifies the type of event
	EventType string

	// Event represents an event in the lifecycle of a resource
	Event[T any] struct {
		Payload T
		Type    EventType
	}

	Publisher[T any] interface {
		Publish(EventType, T)
	}
)
