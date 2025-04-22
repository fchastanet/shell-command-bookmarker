package top

import (
	"github.com/charmbracelet/bubbles/spinner"

	"github.com/fchastanet/shell-command-bookmarker/internal/models"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/command"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
)

// Maker makes new models
type Maker interface {
	Make(id resource.ID, width, height int) (structure.ChildModel, error)
}

// DefaultMaker is the default model maker
type DefaultMaker struct{}

func (m *DefaultMaker) Make(_ resource.ID, _, _ int) (structure.ChildModel, error) {
	return nil, nil
}

// makeMakers makes model makers for making models
func makeMakers(
	app *services.AppService,
	myStyles *styles.Styles,
	spinnerObj *spinner.Model,
) func(kind resource.Kind) models.Maker {
	makers := make(map[string]models.Maker)
	makers["commandList"] = &command.ListMaker{
		App:     app,
		Styles:  myStyles,
		Spinner: spinnerObj,
	}
	makers["search"] = &command.SearchMaker{
		App:     app,
		Styles:  myStyles,
		Spinner: spinnerObj,
	}
	return func(kind resource.Kind) models.Maker {
		key := kind.Key()
		maker, ok := makers[key]
		if !ok {
			return nil
		}
		return maker
	}
}
