package styles

import "github.com/charmbracelet/lipgloss"

// Basic colors
const (
	Black           = lipgloss.Color("#000000")
	Blue            = lipgloss.Color("63")
	BurntOrange     = lipgloss.Color("214")
	Cyan            = lipgloss.Color("#00FFFF")
	DarkGreen       = lipgloss.Color("#325451")
	DarkGrey        = lipgloss.Color("#606362")
	DarkRed         = lipgloss.Color("#FF0000")
	DeepBlue        = lipgloss.Color("39")
	EvenLighterGrey = lipgloss.Color("253")
	Green           = lipgloss.Color("34")
	LighterGrey     = lipgloss.Color("250")
	GreenBlue       = lipgloss.Color("#00A095")
	Grey            = lipgloss.Color("#737373")
	HotPink         = lipgloss.Color("200")
	LightBlue       = lipgloss.Color("81")
	LightGreen      = lipgloss.Color("47")
	LightGrey       = lipgloss.Color("245")
	LightishBlue    = lipgloss.Color("75")
	Magenta         = lipgloss.Color("#FF00FF")
	OffWhite        = lipgloss.Color("#a8a7a5")
	Orange          = lipgloss.Color("214")
	Purple          = lipgloss.Color("135")
	Red             = lipgloss.Color("#FF5353")
	Turquoise       = lipgloss.Color("86")
	Violet          = lipgloss.Color("13")
	White           = lipgloss.Color("#ffffff")
	Yellow          = lipgloss.Color("#DBBD70")
)

// ColorTheme defines a collection of related colors used by the application
type ColorTheme struct {
	// Log levels
	DebugLogLevel         lipgloss.AdaptiveColor
	InfoLogLevel          lipgloss.AdaptiveColor
	ErrorLogLevel         lipgloss.AdaptiveColor
	WarnLogLevel          lipgloss.AdaptiveColor
	LogRecordAttributeKey lipgloss.AdaptiveColor

	// Help colors
	HelpKey  lipgloss.AdaptiveColor
	HelpDesc lipgloss.AdaptiveColor

	// Border colors
	InactivePreviewBorder lipgloss.AdaptiveColor
	ActivePreviewBorder   lipgloss.AdaptiveColor

	// Row states
	CurrentBackground            lipgloss.Color
	CurrentForeground            lipgloss.Color
	SelectedBackground           lipgloss.Color
	SelectedForeground           lipgloss.Color
	CurrentAndSelectedBackground lipgloss.Color
	CurrentAndSelectedForeground lipgloss.Color

	// Other UI elements
	TitleColor                 lipgloss.AdaptiveColor
	GroupReportBackgroundColor lipgloss.Color
	TaskSummaryBackgroundColor lipgloss.Color
	ScrollPercentageBackground lipgloss.AdaptiveColor
}

// NewDefaultColorTheme returns a new color theme with default colors
func NewDefaultColorTheme() *ColorTheme {
	return &ColorTheme{
		DebugLogLevel: lipgloss.AdaptiveColor{Dark: string(Blue), Light: string(LightBlue)},
		InfoLogLevel:  lipgloss.AdaptiveColor{Dark: string(Turquoise), Light: string(Green)},
		ErrorLogLevel: lipgloss.AdaptiveColor{Dark: string(Red), Light: string(Red)},
		WarnLogLevel:  lipgloss.AdaptiveColor{Dark: string(Yellow), Light: string(Yellow)},

		LogRecordAttributeKey: lipgloss.AdaptiveColor{Dark: string(LightGrey), Light: string(LightGrey)},

		HelpKey: lipgloss.AdaptiveColor{
			Dark:  "99",
			Light: "240",
		},

		HelpDesc: lipgloss.AdaptiveColor{
			Dark:  "248",
			Light: "244",
		},

		InactivePreviewBorder: lipgloss.AdaptiveColor{
			Dark:  "238",
			Light: "250",
		},
		ActivePreviewBorder: lipgloss.AdaptiveColor{
			Dark:  string(Blue),
			Light: string(Cyan),
		},

		CurrentBackground:            Grey,
		CurrentForeground:            White,
		SelectedBackground:           lipgloss.Color("110"),
		SelectedForeground:           Black,
		CurrentAndSelectedBackground: lipgloss.Color("117"),
		CurrentAndSelectedForeground: Black,

		TitleColor: lipgloss.AdaptiveColor{
			Dark:  "15",
			Light: "8",
		},

		GroupReportBackgroundColor: EvenLighterGrey,
		TaskSummaryBackgroundColor: EvenLighterGrey,

		ScrollPercentageBackground: lipgloss.AdaptiveColor{
			Dark:  "238",
			Light: "250",
		},
	}
}
