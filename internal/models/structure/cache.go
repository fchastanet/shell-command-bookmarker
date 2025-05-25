package structure

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

// page Cache: not so much for performance but to retain memory of user actions,
// e.g. a user may select a particular row in a table, navigate away from the
// page and later return to the page, and they would expect the same row still
// to be selected.
type Cache struct {
	cache map[string]ChildModel
}

func NewCache() *Cache {
	return &Cache{
		cache: make(map[string]ChildModel),
	}
}

func (c *Cache) Get(page Page) ChildModel {
	return c.cache[page.Kind.Key()+strconv.FormatInt(int64(page.ID), 10)]
}

func (c *Cache) Put(page Page, model ChildModel) {
	c.cache[page.Kind.Key()+strconv.FormatInt(int64(page.ID), 10)] = model
}

func (c *Cache) UpdateAll(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, len(c.cache))
	var i int
	for k := range c.cache {
		cmds[i] = c.cache[k].Update(msg)
		i++
	}
	return cmds
}

func (c *Cache) Update(key Page, msg tea.Msg) tea.Cmd {
	return c.cache[key.Kind.Key()+strconv.FormatInt(int64(key.ID), 10)].Update(msg)
}
