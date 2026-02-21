package ui

import "github.com/charmbracelet/lipgloss"

// ---------------------------------------------------------------------------
// Color palette (adaptive: works on both dark and light terminals)
// ---------------------------------------------------------------------------
var (
	colorPrimary  = lipgloss.AdaptiveColor{Dark: "#A78BFA", Light: "#6D28D9"}
	colorAccent   = lipgloss.AdaptiveColor{Dark: "#38BDF8", Light: "#0369A1"}
	colorSuccess  = lipgloss.AdaptiveColor{Dark: "#34D399", Light: "#047857"}
	colorWarning  = lipgloss.AdaptiveColor{Dark: "#FBBF24", Light: "#B45309"}
	colorDanger   = lipgloss.AdaptiveColor{Dark: "#F87171", Light: "#B91C1C"}
	colorMuted    = lipgloss.AdaptiveColor{Dark: "#6B7280", Light: "#9CA3AF"}
	colorFg       = lipgloss.AdaptiveColor{Dark: "#F1F5F9", Light: "#0F172A"}
	colorSubtle   = lipgloss.AdaptiveColor{Dark: "#CBD5E1", Light: "#475569"}
	colorBorder   = lipgloss.AdaptiveColor{Dark: "#334155", Light: "#CBD5E1"}
	colorHeaderBg = lipgloss.AdaptiveColor{Dark: "#1E293B", Light: "#E2E8F0"}
	colorSelBg    = lipgloss.AdaptiveColor{Dark: "#1D3461", Light: "#BFDBFE"}
	colorSelFg    = lipgloss.AdaptiveColor{Dark: "#E0F2FE", Light: "#1E3A5F"}
	colorTitleBg  = lipgloss.AdaptiveColor{Dark: "#4C1D95", Light: "#7C3AED"}
)

// ---------------------------------------------------------------------------
// Layout constants
// ---------------------------------------------------------------------------
const (
	headerHeight = 3 // top bar height in lines
	footerHeight = 1 // bottom bar height in lines
)

// ---------------------------------------------------------------------------
// Header
// ---------------------------------------------------------------------------

var HeaderBarStyle = lipgloss.NewStyle().
	Background(colorTitleBg).
	Foreground(colorFg).
	Bold(true)

var HeaderTitleStyle = lipgloss.NewStyle().
	Background(colorTitleBg).
	Foreground(lipgloss.AdaptiveColor{Dark: "#FFFFFF", Light: "#FFFFFF"}).
	Bold(true).
	Padding(0, 2)

var HeaderSubtitleStyle = lipgloss.NewStyle().
	Background(colorTitleBg).
	Foreground(lipgloss.AdaptiveColor{Dark: "#C4B5FD", Light: "#EDE9FE"}).
	Padding(0, 1)

var HeaderDividerStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorBorder)

// ---------------------------------------------------------------------------
// Footer
// ---------------------------------------------------------------------------

var FooterStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorMuted).
	Padding(0, 2)

var FooterKeyStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorAccent).
	Bold(true)

var FooterDescStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorMuted)

var FooterSepStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorBorder)

// ---------------------------------------------------------------------------
// Table
// ---------------------------------------------------------------------------

var TableHeaderStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorAccent).
	Bold(true).
	Padding(0, 1)

var RowNormalStyle = lipgloss.NewStyle().
	Foreground(colorFg).
	Padding(0, 1)

var RowOutdatedStyle = lipgloss.NewStyle().
	Foreground(colorWarning).
	Padding(0, 1)

var RowSelectedStyle = lipgloss.NewStyle().
	Background(colorSelBg).
	Foreground(colorSelFg).
	Bold(true).
	Padding(0, 1)

var RowSelectedOutdatedStyle = lipgloss.NewStyle().
	Background(colorSelBg).
	Foreground(colorWarning).
	Bold(true).
	Padding(0, 1)

// Status badge styles
var BadgeOkStyle = lipgloss.NewStyle().
	Foreground(colorSuccess).
	Bold(true)

var BadgeOutdatedStyle = lipgloss.NewStyle().
	Foreground(colorWarning).
	Bold(true)

var BadgeOkSelectedStyle = lipgloss.NewStyle().
	Background(colorSelBg).
	Foreground(colorSuccess).
	Bold(true)

var BadgeOutdatedSelectedStyle = lipgloss.NewStyle().
	Background(colorSelBg).
	Foreground(colorWarning).
	Bold(true)

// Stat bar at the bottom of the table
var StatBarStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorMuted).
	Padding(0, 1)

var StatHighlightStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorAccent).
	Bold(true)

var StatWarningStyle = lipgloss.NewStyle().
	Background(colorHeaderBg).
	Foreground(colorWarning).
	Bold(true)

// ---------------------------------------------------------------------------
// Spinner
// ---------------------------------------------------------------------------

var SpinnerStyle = lipgloss.NewStyle().
	Foreground(colorAccent)

// ---------------------------------------------------------------------------
// Panels (detail, doctor)
// ---------------------------------------------------------------------------

var PanelStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(0, 1)

var PanelTitleStyle = lipgloss.NewStyle().
	Foreground(colorPrimary).
	Bold(true)

var SectionTitleStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true)

var LabelStyle = lipgloss.NewStyle().
	Foreground(colorSubtle).
	Bold(true)

var ValueStyle = lipgloss.NewStyle().
	Foreground(colorFg)

var MutedStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var SuccessStyle = lipgloss.NewStyle().
	Foreground(colorSuccess).
	Bold(true)

var ErrorStyle = lipgloss.NewStyle().
	Foreground(colorDanger).
	Bold(true)

var WarningStyle = lipgloss.NewStyle().
	Foreground(colorWarning).
	Bold(true)

// Badge for package status inside detail panel
var DetailBadgeUpToDateStyle = lipgloss.NewStyle().
	Background(lipgloss.AdaptiveColor{Dark: "#064E3B", Light: "#D1FAE5"}).
	Foreground(lipgloss.AdaptiveColor{Dark: "#34D399", Light: "#065F46"}).
	Bold(true).
	Padding(0, 1)

var DetailBadgeOutdatedStyle = lipgloss.NewStyle().
	Background(lipgloss.AdaptiveColor{Dark: "#451A03", Light: "#FEF3C7"}).
	Foreground(lipgloss.AdaptiveColor{Dark: "#FBBF24", Light: "#92400E"}).
	Bold(true).
	Padding(0, 1)

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

var SearchPromptStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true)

var SearchResultCountStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// ---------------------------------------------------------------------------
// Confirm dialog
// ---------------------------------------------------------------------------

var ConfirmBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorDanger).
	Padding(1, 4)

// ---------------------------------------------------------------------------
// Divider
// ---------------------------------------------------------------------------

var DividerStyle = lipgloss.NewStyle().
	Foreground(colorBorder)
