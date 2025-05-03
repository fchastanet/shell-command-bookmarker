package table

import (
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
)

// RowDefaultActionMsg is sent when a row is selected (usually by pressing Enter)
type RowDefaultActionMsg[V resource.Identifiable] struct {
	Row   V             // The selected row
	Kind  resource.Kind // The kind of resource this row represents
	RowID resource.ID   // The ID of the selected row
}

// RowSelectedActionMsg is sent when a row is selected (usually by navigating with arrow keys)
type RowSelectedActionMsg[V resource.Identifiable] struct {
	Row   V           // The selected row
	RowID resource.ID // The ID of the selected row
}

// RowDeleteActionMsg is sent when a row is requested to be deleted
type RowDeleteActionMsg[V resource.Identifiable] struct {
	Row   V           // The row to delete
	RowID resource.ID // The ID of the row to delete
}

type ReloadMsg[V resource.Identifiable] struct {
	InfoMsg *tui.InfoMsg
	RowID   resource.ID // The ID of the row to go to
}
