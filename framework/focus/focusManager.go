package focus

import (
	tea "github.com/charmbracelet/bubbletea"
)

type FocusManager struct {
	focusedHierarchy   []*Focusable
	focusableHierarchy []*Focusable
}

func NewFocusManager() *FocusManager {
	return &FocusManager{
		focusedHierarchy:   []*Focusable{},
		focusableHierarchy: []*Focusable{},
	}
}

type Focusable interface {
	IsFocusable() bool
	GetInnerFocusableComponents() []*Focusable
	//GetNextFocusableInnerComponent(currentFocusedComponent *string) string
	//GetPreviousFocusableInnerComponent(currentFocusedComponent *string) string
}

type ComponentFocusMsg struct {
	tea.FocusMsg
	Target *Focusable
}

type ComponentBlurMsg struct {
	tea.BlurMsg
	Target *Focusable
}

// FocusNextComponent returns a command to focus the next component in the hierarchy
func (fm *FocusManager) FocusNextComponent() tea.Cmd {
	nextComponent := fm.findNextFocusableComponent()
	if nextComponent == nil {
		return nil
	}

	return func() tea.Msg {
		return ComponentFocusMsg{Target: nextComponent}
	}
}

// findNextFocusableComponent uses recursion to find the next component to focus
func (fm *FocusManager) findNextFocusableComponent() *Focusable {
	return fm.findNextInLevel(fm.focusableHierarchy, 0, fm.focusedHierarchy)
}

// findNextInLevel recursively searches for the next focusable component
func (fm *FocusManager) findNextInLevel(components []*Focusable, depth int, currentPath []*Focusable) *Focusable {
	// Base case 1: No components at this level
	if len(components) == 0 {
		return nil
	}

	// If we're not at the current focus depth yet, go deeper
	if depth < len(currentPath)-1 {
		currentFocused := currentPath[depth]
		// Find current component in this level
		for i, comp := range components {
			if comp == currentFocused {
				// Recurse to next level with current component's inner elements
				inner := (*currentFocused).GetInnerFocusableComponents()
				// Check inner components
				nextInner := fm.findNextInLevel(inner, depth+1, currentPath)
				if nextInner != nil {
					return nextInner
				}

				// If nothing found inner, remove children at depth+1
				fm.focusedHierarchy = fm.focusedHierarchy[:depth+1]

				// If nothing found inner, try next sibling
				if i < len(components)-1 {
					// Try next sibling
					for j := i + 1; j < len(components); j++ {
						if (*components[j]).IsFocusable() {
							// Update focus path
							fm.focusedHierarchy = currentPath[:depth+1]
							fm.focusedHierarchy[depth] = components[j]
							return components[j]
						}
					}
				}

				// Nothing found at this level, go up and try next
				return nil
			}
		}
	}

	// We're at the right depth, or no specific focus yet
	currentIndex := -1

	// If we have current focus at this depth, find its index
	if depth < len(currentPath) {
		currentComponent := currentPath[depth]
		for i, comp := range components {
			if comp == currentComponent {
				currentIndex = i
				break
			}
		}
	}

	// remove component at this depth
	fm.focusedHierarchy = currentPath[:depth]

	// Look for next component at this level
	startIndex := 0
	if currentIndex >= 0 {
		startIndex = currentIndex + 1
	}

	// Check next components at this level
	for i := startIndex; i < len(components); i++ {
		if (*components[i]).IsFocusable() {
			// Check inner components first
			innerComponents := (*components[i]).GetInnerFocusableComponents()
			if len(innerComponents) > 0 {
				var innerFocusable *Focusable = nil
				for j := 0; j < len(innerComponents); j++ {
					if (*innerComponents[j]).IsFocusable() {
						innerFocusable = innerComponents[j]
						break
					}
				}
				if innerFocusable != nil {
					// add current component to the path
					fm.focusedHierarchy = append(fm.focusedHierarchy, components[i])
					return fm.findNextInLevel(innerComponents, depth+1, fm.focusedHierarchy)
				}
			} else {
				// add current component to the path
				fm.focusedHierarchy = append(fm.focusedHierarchy, components[i])
				return components[i]
			}

			// No focusable inner components, skip this one
			return nil
		}
	}

	// Nothing focusable at this level
	fm.focusedHierarchy = currentPath[:depth]
	return nil
}

// findFirstFocusable finds the first focusable component in a list
func (fm *FocusManager) findFirstFocusable(components []*Focusable) *Focusable {
	for _, comp := range components {
		if !(*comp).IsFocusable() {
			continue
		}

		// Check inner components
		innerComps := (*comp).GetInnerFocusableComponents()
		if len(innerComps) > 0 {
			// add current component to the path
			fm.focusedHierarchy = append(fm.focusedHierarchy, comp)
			inner := fm.findFirstFocusable(innerComps)
			if inner != nil {
				return inner
			}
		} else {
			// No inner components, focus this one
			return comp
		}
	}
	return nil
}
