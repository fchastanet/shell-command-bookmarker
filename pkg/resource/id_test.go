package resource

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID_String(t *testing.T) {
	cmd := NewMonotonicID(Command)

	t.Run("string", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(cmd.String(), "#"))
	})
}
