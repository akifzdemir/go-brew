package ui

import (
	"fmt"
	"strings"

	"github.com/akif/gobrew/internal/brew"
	tea "github.com/charmbracelet/bubbletea"
)

// column widths
const (
	colName    = 28
	colVersion = 14
	colLatest  = 14
	colStatus  = 14
)

// TableModel renders the main package list as a fixed-column table.
type TableModel struct {
	packages       []brew.Package
	cursor         int
	offset         int // scroll offset
	visibleRows    int
	confirmPending bool
	width          int
}

func newTableModel() TableModel {
	return TableModel{}
}

func (t *TableModel) setPackages(pkgs []brew.Package) {
	t.packages = pkgs
	if t.cursor >= len(pkgs) {
		t.cursor = max(0, len(pkgs)-1)
	}
}

func (t *TableModel) setSize(w, h int) {
	t.width = w
}

func (t *TableModel) setConfirm(v bool) {
	t.confirmPending = v
}

func (t *TableModel) selectedPackage() *brew.Package {
	if len(t.packages) == 0 {
		return nil
	}
	if t.cursor < 0 || t.cursor >= len(t.packages) {
		return nil
	}
	return &t.packages[t.cursor]
}

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
				t.cursor = len(t.packages) - 1
			}
		case "home", "g":
			t.cursor = 0
		case "end", "G":
			t.cursor = len(t.packages) - 1
		}
	}
	return t, nil
}

func (t TableModel) View(height int) string {
	// Header row + 1 divider + rows
	headerLines := 2
	t.visibleRows = height - headerLines
	if t.visibleRows < 1 {
		t.visibleRows = 1
	}

	// Scroll offset
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+t.visibleRows {
		t.offset = t.cursor - t.visibleRows + 1
	}

	var sb strings.Builder

	// --- Header row ---
	sb.WriteString(t.renderHeader())
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", t.width)))
	sb.WriteString("\n")

	if len(t.packages) == 0 {
		sb.WriteString(MutedStyle.Render("  No packages found. Press r to refresh."))
		return sb.String()
	}

	// --- Data rows ---
	end := t.offset + t.visibleRows
	if end > len(t.packages) {
		end = len(t.packages)
	}

	for i := t.offset; i < end; i++ {
		sb.WriteString(t.renderRow(i))
		sb.WriteString("\n")
	}

	// Scroll indicator
	if len(t.packages) > t.visibleRows {
		shown := fmt.Sprintf("  %d–%d of %d", t.offset+1, end, len(t.packages))
		sb.WriteString(MutedStyle.Render(shown))
	}

	return sb.String()
}

func (t TableModel) renderHeader() string {
	name := padRight("NAME", colName)
	ver := padRight("INSTALLED", colVersion)
	latest := padRight("LATEST", colLatest)
	status := padRight("STATUS", colStatus)
	return TableHeaderStyle.Render(name + ver + latest + status)
}

func (t TableModel) renderRow(i int) string {
	pkg := t.packages[i]
	selected := i == t.cursor

	name := padRight(pkg.Name, colName)
	ver := padRight(pkg.InstalledVersion, colVersion)
	latest := padRight(pkg.LatestVersion, colLatest)

	var statusStr string
	if pkg.Outdated {
		statusStr = "↑ outdated"
	} else {
		statusStr = "✓ ok"
	}
	statusCol := padRight(statusStr, colStatus)

	row := name + ver + latest + statusCol

	switch {
	case selected && pkg.Outdated:
		return RowSelectedOutdatedStyle.Render(row)
	case selected:
		return RowSelectedStyle.Render(row)
	case pkg.Outdated:
		return RowOutdatedStyle.Render(row)
	default:
		return RowNormalStyle.Render(row)
	}
}

// padRight pads or truncates a string to exactly n runes.
func padRight(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		runes = runes[:n-1]
		runes = append(runes, '…')
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
