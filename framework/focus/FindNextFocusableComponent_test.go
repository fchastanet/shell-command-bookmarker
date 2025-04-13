package focus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockFocusable implements the Focusable interface for testing
type MockFocusable struct {
	name            string
	focusable       bool
	innerComponents []Focusable
}

func (m *MockFocusable) GetFocusableUniqueId() string {
	return m.name
}

func (m *MockFocusable) IsFocusable() bool {
	return m.focusable
}

func (m *MockFocusable) GetInnerFocusableComponents() []Focusable {
	return m.innerComponents
}

func TestFindNextFocusableComponent(t *testing.T) {
	// Helper to create focusable components
	createFocusable := func(name string, focusable bool, inner []Focusable) Focusable {
		var f Focusable = &MockFocusable{
			name:            name,
			focusable:       focusable,
			innerComponents: inner,
		}
		return f
	}

	t.Run("empty hierarchy", func(t *testing.T) {
		fm := NewFocusManager()
		next := fm.findNextFocusableComponent()
		assert.Nil(t, next)
	})

	t.Run("single component", func(t *testing.T) {
		fm := NewFocusManager()
		root := createFocusable("root", true, nil)
		fm.rootComponents = []Focusable{root}

		next := fm.findNextFocusableComponent()
		assert.Equal(t, root, next)
		assert.Equal(t, []Focusable{root}, fm.focusedHierarchy)
	})

	t.Run("first component when none focused", func(t *testing.T) {
		fm := NewFocusManager()
		root1 := createFocusable("root1", true, nil)
		root2 := createFocusable("root2", true, nil)
		fm.rootComponents = []Focusable{root1, root2}

		next := fm.findNextFocusableComponent()
		assert.Equal(t, root1, next)
		assert.Equal(t, []Focusable{root1}, fm.focusedHierarchy)
	})

	t.Run("next component when one focused", func(t *testing.T) {
		fm := NewFocusManager()
		root1 := createFocusable("root1", true, nil)
		root2 := createFocusable("root2", true, nil)
		fm.rootComponents = []Focusable{root1, root2}
		fm.focusedHierarchy = []Focusable{root1}

		next := fm.findNextFocusableComponent()
		assert.Equal(t, root2, next)
		assert.Equal(t, []Focusable{root2}, fm.focusedHierarchy)
	})

	t.Run("skip non-focusable components", func(t *testing.T) {
		fm := NewFocusManager()
		root1 := createFocusable("root1", true, nil)
		root2 := createFocusable("root2", false, nil) // Not focusable
		root3 := createFocusable("root3", true, nil)
		fm.rootComponents = []Focusable{root1, root2, root3}
		fm.focusedHierarchy = []Focusable{root1}

		next := fm.findNextFocusableComponent()
		assert.Equal(t, root3, next)
		assert.Equal(t, []Focusable{root3}, fm.focusedHierarchy)
	})

	t.Run("focus after last component should be nil", func(t *testing.T) {
		fm := NewFocusManager()
		root1 := createFocusable("root1", true, nil)
		root2 := createFocusable("root2", true, nil)
		fm.rootComponents = []Focusable{root1, root2}
		fm.focusedHierarchy = []Focusable{root2}

		next := fm.findNextFocusableComponent()
		assert.Nil(t, next)
		assert.Equal(t, []Focusable{}, fm.focusedHierarchy)
	})

	t.Run("focus inner components first", func(t *testing.T) {
		fm := NewFocusManager()
		child1 := createFocusable("child1", true, nil)
		child2 := createFocusable("child2", true, nil)
		parent := createFocusable("parent", true, []Focusable{child1, child2})
		root2 := createFocusable("root2", true, nil)
		fm.rootComponents = []Focusable{parent, root2}
		fm.focusedHierarchy = []Focusable{}

		next := fm.findNextFocusableComponent()
		assert.Equal(t, child1, next)
		assert.Equal(t, []Focusable{parent, child1}, fm.focusedHierarchy)
	})

	t.Run("navigate between inner components", func(t *testing.T) {
		fm := NewFocusManager()
		child1 := createFocusable("child1", true, nil)
		child2 := createFocusable("child2", true, nil)
		parent := createFocusable("parent", true, []Focusable{child1, child2})
		root2 := createFocusable("root2", true, nil)
		fm.rootComponents = []Focusable{parent, root2}
		fm.focusedHierarchy = []Focusable{parent, child1}

		next := fm.findNextFocusableComponent()
		assert.Equal(t, child2, next)
		assert.Equal(t, []Focusable{parent, child2}, fm.focusedHierarchy)
	})

	t.Run("move from inner to next root component", func(t *testing.T) {
		fm := NewFocusManager()
		child1 := createFocusable("child1", true, nil)
		child2 := createFocusable("child2", true, nil)
		parent := createFocusable("parent", true, []Focusable{child1, child2})
		root2 := createFocusable("root2", true, nil)
		fm.rootComponents = []Focusable{parent, root2}
		fm.focusedHierarchy = []Focusable{parent, child2}

		next := fm.findNextFocusableComponent()
		assert.Equal(t, root2, next)
		assert.Equal(t, []Focusable{root2}, fm.focusedHierarchy)
	})

	t.Run("focus after nil should be grandchild", func(t *testing.T) {
		fm := NewFocusManager()

		// parent -> child1 -> grandchild
		//        -> child2
		// root2
		grandchild := createFocusable("grandchild", true, nil)
		child1 := createFocusable("child1", true, []Focusable{grandchild})
		child2 := createFocusable("child2", true, nil)
		parent := createFocusable("parent", true, []Focusable{child1, child2})
		root2 := createFocusable("root2", true, nil)

		fm.rootComponents = []Focusable{parent, root2}
		fm.focusedHierarchy = []Focusable{}

		next := fm.findNextFocusableComponent()
		assert.Equal(t, grandchild, next)
		assert.Equal(t, []Focusable{parent, child1, grandchild}, fm.focusedHierarchy)
	})

	t.Run("focus after grandchild should be child2", func(t *testing.T) {
		fm := NewFocusManager()

		// parent -> child1 -> grandchild
		//        -> child2
		// root2
		grandchild := createFocusable("grandchild", true, nil)
		child1 := createFocusable("child1", true, []Focusable{grandchild})
		child2 := createFocusable("child2", true, nil)
		parent := createFocusable("parent", true, []Focusable{child1, child2})
		root2 := createFocusable("root2", true, nil)
		fm.rootComponents = []Focusable{parent, root2}
		fm.focusedHierarchy = []Focusable{parent, child1, grandchild}

		next := fm.findNextFocusableComponent()
		assert.Equal(t, child2, next)
		assert.Equal(t, []Focusable{parent, child2}, fm.focusedHierarchy)
	})
}
