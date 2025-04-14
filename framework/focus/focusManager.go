package focus

import (
	"iter"
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type settings struct {
	keys keyMap
}

type Manager struct {
	settings         *settings
	lastKey          *string
	focusedHierarchy []Focusable
	rootComponents   []Focusable
	terminalFocused  bool
}

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Tab      key.Binding
	ShiftTab key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("↔", "focus next component"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("Shift-↔", "focus previous component"),
		),
	}
}

func (fm Manager) GetKeyBindings() []key.Binding {
	return []key.Binding{
		fm.settings.keys.Tab, fm.settings.keys.ShiftTab,
	}
}

type Focusable interface {
	IsFocusable() bool
	GetInnerFocusableComponents() []Focusable
	GetFocusableUniqueID() string
	// GetNextFocusableInnerComponent(currentFocusedComponent *string) string
	// GetPreviousFocusableInnerComponent(currentFocusedComponent *string) string
}

type ComponentFocusMsg struct {
	tea.FocusMsg
	Target Focusable
}

type ComponentBlurMsg struct {
	tea.BlurMsg
	Target Focusable
}

func NewFocusManager() *Manager {
	return &Manager{
		settings: &settings{
			keys: defaultKeyMap(),
		},
		lastKey:          nil,
		terminalFocused:  true,
		focusedHierarchy: []Focusable{},
		rootComponents:   []Focusable{},
	}
}

func (fm *Manager) IsTerminalFocused() bool {
	return fm.terminalFocused
}

func (fm *Manager) SetRootComponents(hierarchy []Focusable) {
	fm.rootComponents = hierarchy
}

func (fm *Manager) GetRootComponents() []Focusable {
	return fm.rootComponents
}

func (fm *Manager) SetFocusedHierarchy(hierarchy []Focusable) {
	fm.focusedHierarchy = hierarchy
}

func (fm *Manager) GetFocusedHierarchy() []Focusable {
	return fm.focusedHierarchy
}

func (fm Manager) Init() tea.Cmd {
	return nil
}

func (fm Manager) GetLastKey() string {
	return *fm.lastKey
}

func (fm Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.FocusMsg:
		fm.terminalFocused = true
	case tea.BlurMsg:
		fm.terminalFocused = false
	case tea.KeyMsg:
		lastKey := msg.String()
		fm.lastKey = &lastKey
		switch {
		case key.Matches(msg, fm.settings.keys.Tab):
			cmds = append(cmds, fm.FocusNextComponent())
		case key.Matches(msg, fm.settings.keys.ShiftTab):
			cmds = append(cmds, fm.FocusPreviousComponent())
		}
	}

	return fm, tea.Batch(cmds...)
}

func (fm Manager) View() string {
	return ""
}

func (fm *Manager) FocusPreviousComponent() tea.Cmd {
	return nil
}

// FocusNextComponent returns a command to focus the next component in the hierarchy
func (fm *Manager) FocusNextComponent() tea.Cmd {
	nextComponent := fm.findNextFocusableComponent()
	if nextComponent == nil {
		log.Println("No next focusable component found")
		return nil
	}
	log.Printf("Found next component: %v\n", nextComponent.GetFocusableUniqueID())

	return func() tea.Msg {
		return ComponentFocusMsg{
			FocusMsg: tea.FocusMsg{},
			Target:   nextComponent,
		}
	}
}

// findNextFocusableComponent uses recursion to find the next component to focus
func (fm *Manager) findNextFocusableComponent() Focusable {
	return fm.findNextInLevel(fm.rootComponents, 0, fm.focusedHierarchy)
}

func findComponentIndex(components []Focusable, component Focusable) int {
	for i, comp := range components {
		if comp.GetFocusableUniqueID() == component.GetFocusableUniqueID() {
			return i
		}
	}
	return -1
}

// findNextInLevel recursively searches for the next focusable component
func (fm *Manager) findNextInLevel(components []Focusable, depth int, currentPath []Focusable) Focusable {
	// Base case: No components at this level
	if len(components) == 0 {
		return nil
	}

	// Case 1: We're still navigating down to current focus depth
	if depth < len(currentPath)-1 {
		return fm.navigateCurrentBranch(components, depth, currentPath)
	}

	// Case 2: We're at the target depth or no specific focus yet
	return fm.findNextAtCurrentLevel(components, depth, currentPath)
}

// navigateCurrentBranch handles navigation within the current focus branch
func (fm *Manager) navigateCurrentBranch(components []Focusable, depth int, currentPath []Focusable) Focusable {
	currentFocused := currentPath[depth]
	currentIndex := findComponentIndex(components, currentFocused)

	if currentIndex == -1 {
		return nil // Component not found at this level
	}

	// First try inner components of the focused component
	inner := currentFocused.GetInnerFocusableComponents()
	nextInner := fm.findNextInLevel(inner, depth+1, currentPath)
	if nextInner != nil {
		return nextInner
	}

	// No focusable inner components found, truncate hierarchy and try siblings
	fm.focusedHierarchy = fm.focusedHierarchy[:depth+1]

	// Look for next focusable sibling
	for _, focusableComponent := range iterateFocusableComponents(components, currentIndex+1) {
		fm.focusedHierarchy[depth] = focusableComponent
		return focusableComponent
	}

	return nil
}

// findNextAtCurrentLevel finds the next component at the current level
func (fm *Manager) findNextAtCurrentLevel(components []Focusable, depth int, currentPath []Focusable) Focusable {
	currentIndex := -1

	// If we have focus at this depth, find the current component's index
	if depth < len(currentPath) {
		currentComponent := currentPath[depth]
		currentIndex = findComponentIndex(components, currentComponent)
	}

	// Truncate hierarchy to this depth
	fm.focusedHierarchy = currentPath[:depth]

	// Start searching from next component or beginning
	startIndex := 0
	if currentIndex >= 0 {
		startIndex = currentIndex + 1
	}

	// Look for next focusable component
	for _, focusableComponent := range iterateFocusableComponents(components, startIndex) {
		innerComponents := focusableComponent.GetInnerFocusableComponents()

		if len(innerComponents) > 0 {
			// Has inner components, add to hierarchy and continue deeper
			fm.focusedHierarchy = append(fm.focusedHierarchy, focusableComponent)

			// Try to find focusable inner component
			innerFocusable := firstFocusable(innerComponents)
			if innerFocusable != nil {
				return fm.findNextInLevel(innerComponents, depth+1, fm.focusedHierarchy)
			}
		} else {
			// No inner components, focus this component
			fm.focusedHierarchy = append(fm.focusedHierarchy, focusableComponent)
			return focusableComponent
		}

		return nil // No focusable inner components, skip
	}

	// Nothing focusable found, preserve truncated hierarchy
	fm.focusedHierarchy = currentPath[:depth]
	return nil
}

func firstFocusable(components []Focusable) Focusable {
	for _, comp := range components {
		if comp.IsFocusable() {
			return comp
		}
	}
	return nil
}

func iterateFocusableComponents(components []Focusable, startIndex int) iter.Seq2[int, Focusable] {
	return func(yield func(int, Focusable) bool) {
		for i := startIndex; i < len(components); i++ {
			comp := components[i]
			if !comp.IsFocusable() {
				continue
			}
			if !yield(i, comp) {
				break
			}
		}
	}
}
