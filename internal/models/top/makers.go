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

// makeMakers makes model makers for making models
func makeMakers(
	app *services.AppService,
	myStyles *styles.Styles,
	spinnerObj *spinner.Model,
	keyMaps *KeyMaps,
) func(kind resource.Kind) models.Maker {
	makers := make(map[string]models.Maker)
	makers["commandList"] = &command.ListMaker{
		App:              app,
		Styles:           myStyles,
		Spinner:          spinnerObj,
		NavigationKeyMap: keyMaps.tableNavigation,
		ActionKeyMap:     keyMaps.tableAction,
	}
	makers["search"] = &command.SearchMaker{
		App:     app,
		Styles:  myStyles,
		Spinner: spinnerObj,
	}
	makers["commandEditor"] = &command.EditorMaker{
		App:          app,
		Styles:       myStyles,
		EditorKeyMap: keyMaps.editor,
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
