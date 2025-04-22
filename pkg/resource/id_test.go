package resource

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type GlobalKind struct{}

func (GlobalKind) Key() string { return "global" }
func (GlobalKind) IsKind()     {}

func TestID_String(t *testing.T) {
	monotonicIDService := NewMonotonicIDService()
	cmd := monotonicIDService.NewMonotonicID(GlobalKind{})

	t.Run("string", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(cmd.String(), "#"))
	})

	t.Run("kind", func(t *testing.T) {
		assert.Equal(t, cmd.Kind, GlobalKind{})
	})
}
