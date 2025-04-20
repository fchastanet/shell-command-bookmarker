package top

import (
	"github.com/charmbracelet/bubbles/spinner"

	"github.com/fchastanet/shell-command-bookmarker/internal/models"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/command"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
)

// Maker makes new models
type Maker interface {
	Make(id resource.ID, width, height int) (models.ChildModel, error)
}

// DefaultMaker is the default model maker
type DefaultMaker struct{}

func (m *DefaultMaker) Make(id resource.ID, width, height int) (models.ChildModel, error) {
	return nil, nil
}

// makeMakers makes model makers for making models
func makeMakers(
	app *services.AppService,
	spinner *spinner.Model,
) func(kind models.Kind) models.Maker {
	makers := map[models.Kind]Maker{
		models.CommandListKind: &command.CommandListMaker{
			App:     app,
			Spinner: spinner,
		},
		models.CommandKind: &command.CommandMaker{
			App:     app,
			Spinner: spinner,
		},
	}
	return func(kind models.Kind) models.Maker {
		maker, ok := makers[models.Kind(kind)]
		if !ok {
			return &DefaultMaker{}
		}
		return maker
	}
}
