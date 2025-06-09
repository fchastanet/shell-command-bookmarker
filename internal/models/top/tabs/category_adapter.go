package tabs

import (
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	dbmodels "github.com/fchastanet/shell-command-bookmarker/internal/services/models"
	"github.com/fchastanet/shell-command-bookmarker/pkg/category"
	pkgTabs "github.com/fchastanet/shell-command-bookmarker/pkg/components/tabs"
	"github.com/fchastanet/shell-command-bookmarker/pkg/sort"
)

const (
	// AvailableCommands represents commands that are available for use
	AvailableCommands pkgTabs.CategoryType = iota
	// SavedCommands represents commands that have been saved
	SavedCommands
	// NewCommands represents commands that have been imported but not yet saved
	NewCommands
	// DeletedCommands represents commands that have been marked as deleted
	DeletedCommands
	// AllCommands represents all commands regardless of status
	AllCommands
)

// CategoryAdapter helps translate between UI category types and service-level categories
type CategoryAdapter struct {
	historyService *services.HistoryService
	sortStyles     sort.EditorSortStyles
}

// NewCategoryAdapter creates a new adapter for category conversions
func NewCategoryAdapter(
	historyService *services.HistoryService,
	sortStyles sort.EditorSortStyles,
) *CategoryAdapter {
	return &CategoryAdapter{
		historyService: historyService,
		sortStyles:     sortStyles,
	}
}

func (ca *CategoryAdapter) GetCategoryTabs(
	compareBySortFieldFunc sort.CompareBySortFieldFunc[*dbmodels.Command, string],
) []pkgTabs.CategoryTab[
	*dbmodels.Command,
	dbmodels.CommandStatus,
	string,
] {
	sortFields := []string{
		structure.FieldID,
		structure.FieldTitle,
		structure.FieldScript,
		structure.FieldStatus,
		structure.FieldLintStatus,
		structure.FieldCreationDate,
		structure.FieldModificationDate,
	}

	// Create a function that returns a new sort state for each tab
	createNewSortState := func() *sort.State[*dbmodels.Command, string] {
		return sort.NewDefaultState(
			ca.sortStyles,
			structure.FieldID,
			sortFields,
			compareBySortFieldFunc,
		)
	}

	return []pkgTabs.CategoryTab[
		*dbmodels.Command,
		dbmodels.CommandStatus,
		string,
	]{
		NewCategoryTab(
			"Available",
			createNewSortState(),
			AvailableCommands,
			[]dbmodels.CommandStatus{
				dbmodels.CommandStatusSaved,
				dbmodels.CommandStatusImported,
			},
		),
		NewCategoryTab(
			"Saved",
			createNewSortState(),
			SavedCommands,
			[]dbmodels.CommandStatus{
				dbmodels.CommandStatusSaved,
			},
		),
		NewCategoryTab(
			"New",
			createNewSortState(),
			NewCommands,
			[]dbmodels.CommandStatus{
				dbmodels.CommandStatusImported,
			},
		),
		NewCategoryTab(
			"Deleted",
			createNewSortState(),
			DeletedCommands,
			[]dbmodels.CommandStatus{
				dbmodels.CommandStatusDeleted,
			},
		),
		NewCategoryTab(
			"All",
			createNewSortState(),
			AllCommands,
			[]dbmodels.CommandStatus{
				dbmodels.CommandStatusSaved,
				dbmodels.CommandStatusImported,
				dbmodels.CommandStatusDeleted,
				dbmodels.CommandStatusObsolete,
			},
		),
	}
}

func NewCategoryTab(
	title string,
	sortState *sort.State[*dbmodels.Command, string],
	categoryType pkgTabs.CategoryType,
	commandTypes []dbmodels.CommandStatus,
) pkgTabs.CategoryTab[
	*dbmodels.Command,
	dbmodels.CommandStatus,
	string,
] {
	return pkgTabs.CategoryTab[
		*dbmodels.Command,
		dbmodels.CommandStatus,
		string,
	]{
		Title: title,
		Type:  categoryType,
		Count: 0,
		FilterState: &category.FilterSortState[*dbmodels.Command, string]{
			FilterValue: "",
			SortState:   sortState,
		},
		CommandTypes: commandTypes,
	}
}

// GetCommandTypesByCategory returns command statuses for a UI category type
func (ca *CategoryAdapter) GetCategoryTabConfiguration(
	cat pkgTabs.CategoryType,
	compareBySortFieldFunc sort.CompareBySortFieldFunc[*dbmodels.Command, string],
) pkgTabs.CategoryTab[*dbmodels.Command, dbmodels.CommandStatus, string] {
	return ca.GetCategoryTabs(compareBySortFieldFunc)[cat]
}

// GetCategoryCounts maps service-level category counts to UI category types
func (ca *CategoryAdapter) GetCategoryCounts() (map[pkgTabs.CategoryType]int, error) {
	// Get category counts from service
	serviceCounts, err := ca.historyService.GetCommandCountsByCategory()
	if err != nil {
		return nil, err
	}

	// Map service categories to UI categories
	uiCounts := make(map[pkgTabs.CategoryType]int)
	uiCounts[AvailableCommands] = serviceCounts[services.CommandCategoryAvailable]
	uiCounts[SavedCommands] = serviceCounts[services.CommandCategorySaved]
	uiCounts[NewCommands] = serviceCounts[services.CommandCategoryNew]
	uiCounts[DeletedCommands] = serviceCounts[services.CommandCategoryDeleted]
	uiCounts[AllCommands] = serviceCounts[services.CommandCategoryAll]

	return uiCounts, nil
}
