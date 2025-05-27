package services

import (
	"testing"

	"github.com/fchastanet/shell-command-bookmarker/internal/services/models"
	"github.com/stretchr/testify/assert"
)

func TestHistoryService_CreateCommandsString(t *testing.T) {
	tests := []struct {
		name     string
		commands []*models.Command
		want     string
	}{
		{
			name:     "Empty commands list",
			commands: []*models.Command{},
			want:     "",
		},
		{
			name: "Single command",
			commands: []*models.Command{
				{
					Script: "echo 'hello world'",
				},
			},
			want: "echo 'hello world'",
		},
		{
			name: "Multiple commands",
			commands: []*models.Command{
				{
					Script: "echo 'hello'",
				},
				{
					Script: "echo 'world'",
				},
				{
					Script: "ls -la",
				},
			},
			want: "echo 'hello'\necho 'world'\nls -la",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal HistoryService for testing
			//nolint:exhaustruct // Only testing CreateCommandsString which doesn't use struct fields
			s := &HistoryService{}
			got := s.CreateCommandsString(tt.commands)
			assert.Equal(t, tt.want, got)
		})
	}
}
