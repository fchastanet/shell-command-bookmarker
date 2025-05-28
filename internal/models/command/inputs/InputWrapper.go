package inputs

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
)

// InputWrapper wraps textinput.Model to implement the Input interface
type InputWrapper struct {
	Model    *textinput.Model
	readOnly bool
}

func NewInputWrapper(placeHolder string) *InputWrapper {
	textInput := textinput.New()
	textInput.Placeholder = placeHolder
	return &InputWrapper{
		Model:    &textInput,
		readOnly: false,
	}
}

func (w *InputWrapper) SetCharLimit(charLimit int) {
	w.Model.CharLimit = charLimit
}

// Update implements the Input interface
func (w *InputWrapper) Update(msg tea.Msg) (Input, tea.Cmd) {
	newModel, cmd := w.Model.Update(msg)
	w.Model = &newModel
	return w, cmd
}

func (w *InputWrapper) SetReadOnly(readOnly bool) {
	w.readOnly = readOnly
}

func (w *InputWrapper) SetWidth(width int) {
	w.Model.Width = width
}

func (*InputWrapper) SetHeight(_ int) {
	// do nothing
}

func (w *InputWrapper) Blur() {
	w.Model.Blur()
}

func (w *InputWrapper) Focus() tea.Cmd {
	return w.Model.Focus()
}

func (w *InputWrapper) Value() string {
	return w.Model.Value()
}

func (w *InputWrapper) SetValue(value string) {
	w.Model.SetValue(value)
}

func (w *InputWrapper) View() string {
	if w.readOnly {
		return w.Model.Value()
	}
	txt := w.Model.View()
	if !w.readOnly && w.Model.CharLimit > 0 {
		txt += fmt.Sprintf("\nLength: %d/%d\n", len(w.Model.Value()), w.Model.CharLimit)
	}
	return txt
}
