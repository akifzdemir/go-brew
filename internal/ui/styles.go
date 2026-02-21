package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	colorPrimary  = lipgloss.Color("#7C3AED") // purple
	colorAccent   = lipgloss.Color("#06B6D4") // cyan
	colorSuccess  = lipgloss.Color("#10B981") // green
	colorWarning  = lipgloss.Color("#F59E0B") // amber
	colorDanger   = lipgloss.Color("#EF4444") // red
	colorMuted    = lipgloss.Color("#6B7280") // gray
	colorFg       = lipgloss.Color("#F9FAFB") // near-white
	colorBg       = lipgloss.Color("#111827") // near-black
	colorSelected = lipgloss.Color("#1E1B4B") // deep indigo
	colorBorder   = lipgloss.Color("#374151") // dark gray
	colorHeaderBg = lipgloss.Color("#1F2937") // dark slate
)

// Header bar at the top of the screen
var HeaderStyle = lipgloss.NewStyle().
	Background(colorPrimary).
	Foreground(colorFg).
	Bold(true).
	Padding(0, 2).
	Width(0) // set dynamically

// AppTitleStyle is the app name inside the header
var AppTitleStyle = lipgloss.NewStyle().
	Foreground(colorFg).
	Bold(true)

// FooterStyle is the keybind hint bar at the bottom
var FooterStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorMuted).
	Padding(0, 2)

// FooterKeyStyle highlights a key name in the footer
var FooterKeyStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true)

// FooterDescStyle is the description next to a key
var FooterDescStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// TableHeaderStyle styles the column header row
var TableHeaderStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorAccent).
	Bold(true).
	Padding(0, 1)

// RowNormalStyle is an unselected, up-to-date row
var RowNormalStyle = lipgloss.NewStyle().
	Foreground(colorFg).
	Padding(0, 1)

// RowOutdatedStyle highlights an outdated package
var RowOutdatedStyle = lipgloss.NewStyle().
	Foreground(colorWarning).
	Padding(0, 1)

// RowSelectedStyle is the cursor row
var RowSelectedStyle = lipgloss.NewStyle().
	Background(colorSelected).
	Foreground(colorFg).
	Bold(true).
	Padding(0, 1)

// RowSelectedOutdatedStyle is the cursor row when the package is outdated
var RowSelectedOutdatedStyle = lipgloss.NewStyle().
	Background(colorSelected).
	Foreground(colorWarning).
	Bold(true).
	Padding(0, 1)

// StatusUpToDate badge
var StatusUpToDate = lipgloss.NewStyle().
	Foreground(colorSuccess).
	Bold(true).
	Render("✓ up-to-date")

// StatusOutdated badge
var StatusOutdated = lipgloss.NewStyle().
	Foreground(colorWarning).
	Bold(true).
	Render("↑ outdated")

// SpinnerStyle
var SpinnerStyle = lipgloss.NewStyle().
	Foreground(colorAccent)

// PanelStyle is the border for detail / doctor panels
var PanelStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(1, 2)

// PanelTitleStyle is the bold title inside a panel
var PanelTitleStyle = lipgloss.NewStyle().
	Foreground(colorPrimary).
	Bold(true).
	MarginBottom(1)

// LabelStyle is a field label inside a panel
var LabelStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true)

// ValueStyle is a field value inside a panel
var ValueStyle = lipgloss.NewStyle().
	Foreground(colorFg)

// MutedStyle is muted / secondary text
var MutedStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// SuccessStyle for success messages
var SuccessStyle = lipgloss.NewStyle().
	Foreground(colorSuccess).
	Bold(true)

// ErrorStyle for error messages
var ErrorStyle = lipgloss.NewStyle().
	Foreground(colorDanger).
	Bold(true)

// WarningStyle for warning text
var WarningStyle = lipgloss.NewStyle().
	Foreground(colorWarning).
	Bold(true)

// ConfirmPromptStyle for confirm dialogs
var ConfirmPromptStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorDanger).
	Padding(1, 3).
	Bold(true).
	Foreground(colorFg)

// SearchInputStyle for the search text input box
var SearchInputStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorAccent).
	Padding(0, 1)

// DividerStyle is a horizontal separator line
var DividerStyle = lipgloss.NewStyle().
	Foreground(colorBorder)
