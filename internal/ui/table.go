package ui

import (
	"fmt"
	"strings"

	"github.com/akif/gobrew/internal/brew"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TableModel renders the main package list as a full-width table.
type TableModel struct {
	packages       []brew.Package
	cursor         int
	offset         int
	visibleRows    int
	confirmPending bool
	width          int
	height         int
}

func newTableModel() TableModel {
	return TableModel{}
}

func (t *TableModel) setPackages(pkgs []brew.Package) {
	t.packages = pkgs
	if t.cursor >= len(pkgs) {
		t.cursor = clamp(len(pkgs)-1, 0, len(pkgs)-1)
	}
}

func (t *TableModel) setSize(w, h int) {
	t.width = w
	t.height = h
}

func (t *TableModel) setConfirm(v bool) {
	t.confirmPending = v
}

func (t *TableModel) selectedPackage() *brew.Package {
	if len(t.packages) == 0 || t.cursor < 0 || t.cursor >= len(t.packages) {
		return nil
	}
	return &t.packages[t.cursor]
}

// Update handles navigation keys — called directly by the root model handler.
func (t TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && !t.confirmPending {
		switch key.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
			}
		case "down", "j":
			if t.cursor < len(t.packages)-1 {
				t.cursor++
			}
		case "pgup":
			t.cursor -= t.visibleRows
			if t.cursor < 0 {
				t.cursor = 0
			}
		case "pgdown":
			t.cursor += t.visibleRows
			if t.cursor >= len(t.packages) {
				t.cursor = max(0, len(t.packages)-1)
			}
		case "home", "g":
			t.cursor = 0
		case "end", "G":
			if len(t.packages) > 0 {
				t.cursor = len(t.packages) - 1
			}
		}
	}
	return t, nil
}

// View renders the table into the given height.
func (t TableModel) View(height int) string {
	// 2 = col header + divider; 1 = stat bar at bottom
	t.visibleRows = height - 3
	if t.visibleRows < 1 {
		t.visibleRows = 1
	}

	// Clamp scroll offset so cursor is always visible
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+t.visibleRows {
		t.offset = t.cursor - t.visibleRows + 1
	}

	// Dynamic column widths based on terminal width
	// Layout: │ # │ Name │ Installed │ Latest │ Status │
	//          3    dyn    12          12        12
	colIdx := 4
	colStatus := 14
	colVer := 13
	colLatest := 13
	colName := t.width - colIdx - colVer - colLatest - colStatus - 2 // 2 = padding
	if colName < 10 {
		colName = 10
	}

	var sb strings.Builder

	// Column header row
	sb.WriteString(t.renderHeader(colIdx, colName, colVer, colLatest, colStatus))
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", t.width)))
	sb.WriteString("\n")

	if len(t.packages) == 0 {
		empty := lipgloss.Place(t.width, t.visibleRows,
			lipgloss.Center, lipgloss.Center,
			MutedStyle.Render("No packages found. Press r to refresh."),
		)
		sb.WriteString(empty)
	} else {
		end := t.offset + t.visibleRows
		if end > len(t.packages) {
			end = len(t.packages)
		}
		for i := t.offset; i < end; i++ {
			sb.WriteString(t.renderRow(i, colIdx, colName, colVer, colLatest, colStatus))
			sb.WriteString("\n")
		}
		// Pad remaining lines so the stat bar stays at the bottom
		rendered := end - t.offset
		for i := rendered; i < t.visibleRows; i++ {
			sb.WriteString("\n")
		}
	}

	// Stat bar
	sb.WriteString(t.renderStatBar())

	return sb.String()
}

func (t TableModel) renderHeader(colIdx, colName, colVer, colLatest, colStatus int) string {
	idx := padRight("#", colIdx)
	name := padRight("NAME", colName)
	ver := padRight("INSTALLED", colVer)
	latest := padRight("LATEST", colLatest)
	status := padRight("STATUS", colStatus)
	return TableHeaderStyle.Width(t.width).Render(idx + name + ver + latest + status)
}

func (t TableModel) renderRow(i, colIdx, colName, colVer, colLatest, colStatus int) string {
	pkg := t.packages[i]
	selected := i == t.cursor

	idxStr := padRight(fmt.Sprintf("%d", i+1), colIdx)
	name := padRight(pkg.Name, colName)
	ver := padRight(pkg.InstalledVersion, colVer)
	latest := padRight(pkg.LatestVersion, colLatest)

	// Build the text parts
	textRow := idxStr + name + ver + latest

	// Status badge — rendered inline, full row gets background from style
	var statusText string
	if pkg.Outdated {
		statusText = "↑ outdated"
	} else {
		statusText = "✓ ok      "
	}

	switch {
	case selected && pkg.Outdated:
		// Full-width background for selected row
		rowStyle := RowSelectedOutdatedStyle.Width(t.width)
		row := textRow + padRight(statusText, colStatus)
		return rowStyle.Render(row)
	case selected:
		rowStyle := RowSelectedStyle.Width(t.width)
		row := textRow + padRight(statusText, colStatus)
		return rowStyle.Render(row)
	case pkg.Outdated:
		// For unselected rows, color the status badge separately
		plain := RowOutdatedStyle.Render(textRow)
		badge := BadgeOutdatedStyle.Render(statusText)
		return plain + badge
	default:
		plain := RowNormalStyle.Render(textRow)
		badge := BadgeOkStyle.Render(statusText)
		return plain + badge
	}
}

func (t TableModel) renderStatBar() string {
	total := len(t.packages)
	outdated := 0
	for _, p := range t.packages {
		if p.Outdated {
			outdated++
		}
	}

	totalStr := StatHighlightStyle.Render(fmt.Sprintf("%d", total)) +
		StatBarStyle.Render(" packages")

	var outdatedStr string
	if outdated > 0 {
		outdatedStr = StatBarStyle.Render("  ·  ") +
			StatWarningStyle.Render(fmt.Sprintf("%d", outdated)) +
			StatBarStyle.Render(" outdated")
	} else {
		outdatedStr = StatBarStyle.Render("  ·  ") +
			StatBarStyle.Render("all up to date")
	}

	// Scroll position on the right
	scrollStr := ""
	if total > t.visibleRows && t.visibleRows > 0 {
		end := t.offset + t.visibleRows
		if end > total {
			end = total
		}
		scrollStr = StatBarStyle.Render(
			fmt.Sprintf("%d–%d / %d ", t.offset+1, end, total),
		)
	}

	left := totalStr + outdatedStr
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(scrollStr)
	gap := t.width - leftW - rightW - 2
	if gap < 0 {
		gap = 0
	}

	return StatBarStyle.Width(t.width).Render(
		left + strings.Repeat(" ", gap) + scrollStr,
	)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// padRight pads or truncates a string to exactly n runes.
func padRight(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		if n > 1 {
			runes = runes[:n-1]
			runes = append(runes, '…')
		} else {
			runes = runes[:n]
		}
	}
	for len(runes) < n {
		runes = append(runes, ' ')
	}
	return string(runes)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
