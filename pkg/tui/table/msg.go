package table

import "github.com/fchastanet/shell-command-bookmarker/pkg/resource"

// RowDefaultActionMsg is sent when a row is selected (usually by pressing Enter)
type RowDefaultActionMsg[V resource.Identifiable] struct {
	Row   V             // The selected row
	Kind  resource.Kind // The kind of resource this row represents
	RowID resource.ID   // The ID of the selected row
}

type ReloadMsg[V resource.Identifiable] struct{}
