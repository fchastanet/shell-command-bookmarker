package resource

import (
	"fmt"
	"sync"
)

// IDService defines the interface for ID generation
type IDService interface {
	NewMonotonicID(kind Kind) MonotonicID
}

// MonotonicIDService provides monotonically increasing IDs for resources
type MonotonicIDService struct {
	// nextMonotonicID provides the next monotonic ID for each kind
	nextMonotonicID map[Kind]ID
	mu              sync.Mutex
}

// NewMonotonicIDService creates a new instance of MonotonicIDService
func NewMonotonicIDService() *MonotonicIDService {
	return &MonotonicIDService{
		nextMonotonicID: make(map[Kind]ID),
		mu:              sync.Mutex{},
	}
}

// MonotonicID is an identifier based on an ever-increasing serial number, and a
// kind to differentiate it from other kinds of identifiers.
type MonotonicID struct {
	Serial ID
	Kind   Kind
}

// NewMonotonicID generates a new monotonic ID for the given kind
func (s *MonotonicIDService) NewMonotonicID(kind Kind) MonotonicID {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextMonotonicID[kind]
	s.nextMonotonicID[kind]++

	return MonotonicID{
		Serial: id,
		Kind:   kind,
	}
}

// String provides a human readable representation of the identifier.
func (id MonotonicID) String() string {
	return fmt.Sprintf("#%d", id.Serial)
}
