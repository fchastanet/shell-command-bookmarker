package styles

const (
	// MinHeight is the minimum height of the TUI.
	MinHeight = 24
	// Height of prompt including borders
	PromptHeight = 3
	// FooterHeight is the height of the footer at the bottom of the TUI.
	FooterHeight = 1
	// Height of help widget, including borders
	HelpWidgetHeight = 12
	// MinContentHeight is the minimum height of content above the footer.
	MinContentHeight = MinHeight - FooterHeight
	// MinContentWidth is the minimum width of the content.
	MinContentWidth = 80
	// minimum height of each pane
	MinPaneHeight = 4
	// minimum width of each pane
	MinPaneWidth = 20
	// defaultTopRightPaneHeight is the default height of the top right pane.
	DefaultTopRightPaneHeight = 15
	// defaultLeftPaneWidth is the default width of the left pane.
	DefaultLeftPaneWidth = 40
)

func checkDimension() {
	if (MinPaneHeight*2)+PromptHeight+HelpWidgetHeight > MinContentHeight {
		panic("minimum heights of panes, prompt, footer, and help cannot exceed overall minimum height")
	}
	if MinPaneWidth*2 > MinContentWidth {
		panic("minimum width of panes must be no more than half of the minimum content width")
	}
	if MinPaneHeight > DefaultTopRightPaneHeight {
		panic("default top right pane height must not be lower than the overall minimum height")
	}
	if MinPaneWidth > DefaultLeftPaneWidth {
		panic("default left pane width must not be lower than the overall minimum width")
	}
}
