package structure

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
)

type ChildModel interface {
	Init() tea.Cmd
	Update(tea.Msg) tea.Cmd
	View() string
	BeforeSwitchPane() tea.Cmd
}

// Page identifies an instance of a model
type Page struct {
	// The model kind. Identifies the model maker to construct the page.
	Kind resource.Kind
	// ID of resource for a model. If the model does not have a single resource
	// but is say a listing of resources, then this is nil.
	ID resource.ID
}

// ModelHelpBindings is implemented by models that surface further help bindings
// specific to the model.
type ModelHelpBindings interface {
	HelpBindings() []*key.Binding
}
