package table

import (
	"slices"
	"testing"

	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

type ResourceTestKind struct{ key string }

func (r ResourceTestKind) Key() string { return r.key }
func (r ResourceTestKind) IsKind()     {}

var (
	resourceTestKind   = &ResourceTestKind{key: "test"}
	monotonicIDService = resource.NewMonotonicIDService()
	resource0          = testResource{n: 0, ID: monotonicIDService.NewMonotonicID(resourceTestKind)}
	resource1          = testResource{n: 1, ID: monotonicIDService.NewMonotonicID(resourceTestKind)}
	resource2          = testResource{n: 2, ID: monotonicIDService.NewMonotonicID(resourceTestKind)}
	resource3          = testResource{n: 3, ID: monotonicIDService.NewMonotonicID(resourceTestKind)}
	resource4          = testResource{n: 4, ID: monotonicIDService.NewMonotonicID(resourceTestKind)}
	resource5          = testResource{n: 5, ID: monotonicIDService.NewMonotonicID(resourceTestKind)}
)

type testResource struct {
	ID resource.MonotonicID
	n  int
}

func (r *testResource) GetID() resource.ID { return r.ID.Serial }

// setupTest sets up a table test with several rows. Each row is keyed with an
// int, and the row item is an int corresponding to the key, for ease of
// testing. The rows are sorted from lowest int to highest int.
func setupTest() Model[*testResource] {
	renderer := func(_ *testResource) RenderedRow { return nil }
	tbl := New(styles.NewStyles().TableStyle, nil, renderer, 0, 0,
		WithSortFunc(func(i, j *testResource) int {
			if i.n < j.n {
				return -1
			}
			return 1
		}),
	)
	tbl.SetItems(
		&resource0,
		&resource1,
		&resource2,
		&resource3,
		&resource4,
		&resource5,
	)
	return tbl
}

func TestTable_CurrentRow(t *testing.T) {
	tbl := setupTest()

	got, ok := tbl.CurrentRow()
	require.True(t, ok)

	assert.Equal(t, &resource0, got)
}

func TestTable_ToggleSelection(t *testing.T) {
	tbl := setupTest()

	tbl.ToggleSelection()

	assert.Len(t, tbl.selected, 1)
	assert.Equal(t, &resource0, tbl.selected[resource0.ID.Serial])
}

func TestTable_SelectRange(t *testing.T) {
	tests := []struct {
		name     string
		selected []resource.ID
		cursor   int
		want     []resource.ID
	}{
		{
			name:     "select no range when nothing is selected, and cursor is on first row",
			selected: []resource.ID{},
			cursor:   0,
			want:     []resource.ID{},
		},
		{
			name:     "select no range when nothing is selected, and cursor is on last row",
			selected: []resource.ID{},
			cursor:   0,
			want:     []resource.ID{},
		},
		{
			name:     "select no range when cursor is on the only selected row",
			selected: []resource.ID{resource0.ID.Serial},
			cursor:   0,
			want:     []resource.ID{resource0.ID.Serial},
		},
		{
			name:     "select all rows between selected top row and cursor on last row",
			selected: []resource.ID{resource0.ID.Serial}, // first row
			cursor:   5,                                  // last row
			want:     []resource.ID{resource0.ID.Serial, resource1.ID.Serial, resource2.ID.Serial, resource3.ID.Serial, resource4.ID.Serial, resource5.ID.Serial},
		},
		{
			name:     "select rows between selected top row and cursor in third row",
			selected: []resource.ID{resource0.ID.Serial}, // first row
			cursor:   2,                                  // third row
			want:     []resource.ID{resource0.ID.Serial, resource1.ID.Serial, resource2.ID.Serial},
		},
		{
			name:     "select rows between selected top row and cursor in third row, ignoring selected last row",
			selected: []resource.ID{resource0.ID.Serial, resource5.ID.Serial}, // first and last row
			cursor:   2,                                                       // third row
			want:     []resource.ID{resource0.ID.Serial, resource1.ID.Serial, resource2.ID.Serial, resource5.ID.Serial},
		},
		{
			name:     "select rows between cursor in third row and selected last row",
			selected: []resource.ID{resource5.ID.Serial}, // last row
			cursor:   2,                                  // third row
			want:     []resource.ID{resource2.ID.Serial, resource3.ID.Serial, resource4.ID.Serial, resource5.ID.Serial},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := setupTest()
			for _, id := range tt.selected {
				tbl.ToggleSelectionByID(id)
			}
			tbl.currentRowIndex = tt.cursor

			tbl.SelectRange()

			got := maps.Keys(tbl.selected)
			slices.SortFunc(got, sortStrings)
			slices.SortFunc(tt.want, sortStrings)
			assert.Equal(t, tt.want, got)
		})
	}
}

func sortStrings(i, j resource.ID) int {
	if i < j {
		return -1
	}
	return 1
}
