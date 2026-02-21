package ui

import (
	"fmt"
	"strings"

	"github.com/akif/gobrew/internal/brew"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetailModel shows detailed information about a selected package.
type DetailModel struct {
	info   *brew.InfoResult
	width  int
	height int
	scroll int
	lines  []string
}

func (d *DetailModel) setInfo(info *brew.InfoResult) {
	d.info = info
	d.scroll = 0
	d.lines = nil // rebuilt in View when width is known
}

func (d *DetailModel) setSize(w, h int) {
	d.width = w
	d.height = h
	d.lines = nil // reset so lines are rebuilt at new width
}

func (d DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		visibleLines := d.height - headerHeight - footerHeight - 2
		if visibleLines < 1 {
			visibleLines = 1
		}
		maxScroll := len(d.lines) - visibleLines
		if maxScroll < 0 {
			maxScroll = 0
		}
		switch key.String() {
		case "up", "k":
			if d.scroll > 0 {
				d.scroll--
			}
		case "down", "j":
			if d.scroll < maxScroll {
				d.scroll++
			}
		case "pgup":
			d.scroll -= visibleLines
			if d.scroll < 0 {
				d.scroll = 0
			}
		case "pgdown":
			d.scroll += visibleLines
			if d.scroll > maxScroll {
				d.scroll = maxScroll
			}
		case "home":
			d.scroll = 0
		case "end":
			d.scroll = maxScroll
		}
	}
	return d, nil
}

func (d DetailModel) View(height int) string {
	if d.info == nil {
		return lipgloss.Place(d.width, height,
			lipgloss.Center, lipgloss.Center,
			MutedStyle.Render("No package selected."),
		)
	}

	// Build lines lazily (or after a resize)
	if len(d.lines) == 0 {
		d.lines = d.buildLines()
	}

	visibleLines := height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	maxScroll := len(d.lines) - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if d.scroll > maxScroll {
		d.scroll = maxScroll
	}

	end := d.scroll + visibleLines
	if end > len(d.lines) {
		end = len(d.lines)
	}

	var sb strings.Builder
	for _, line := range d.lines[d.scroll:end] {
		sb.WriteString(line + "\n")
	}

	// Scroll indicator
	if len(d.lines) > visibleLines {
		pct := 0
		if maxScroll > 0 {
			pct = (d.scroll * 100) / maxScroll
		}
		indicator := fmt.Sprintf(" %d%% (%d/%d lines) ", pct, d.scroll+1, len(d.lines))
		sb.WriteString(DividerStyle.Render(strings.Repeat("─", d.width-len(indicator))) +
			MutedStyle.Render(indicator))
	}

	return sb.String()
}

func (d DetailModel) buildLines() []string {
	info := d.info
	w := d.width
	if w < 20 {
		w = 80
	}

	innerW := w - 4 // account for panel padding
	labelW := 14
	valueW := innerW - labelW - 2
	if valueW < 10 {
		valueW = 10
	}

	var lines []string

	// ── Title + badge ───────────────────────────────────────────────────────
	title := PanelTitleStyle.Render(" " + info.Name + " ")
	var badge string
	// Determine if outdated by checking installed vs latest
	installed := ""
	if len(info.Installed) > 0 {
		installed = info.Installed[len(info.Installed)-1].Version
	}
	if installed != "" && info.Version != "" && installed != info.Version {
		badge = DetailBadgeOutdatedStyle.Render(" ↑ OUTDATED ")
	} else {
		badge = DetailBadgeUpToDateStyle.Render(" ✓ UP TO DATE ")
	}
	titleGap := w - lipgloss.Width(title) - lipgloss.Width(badge) - 2
	if titleGap < 0 {
		titleGap = 0
	}
	lines = append(lines, title+strings.Repeat(" ", titleGap)+badge)
	lines = append(lines, DividerStyle.Render(strings.Repeat("─", w)))
	lines = append(lines, "")

	// ── Field helper ────────────────────────────────────────────────────────
	field := func(label, value string) string {
		if value == "" {
			value = MutedStyle.Render("—")
		}
		l := LabelStyle.Width(labelW).Render(label)
		v := ValueStyle.Width(valueW).Render(value)
		return "  " + lipgloss.JoinHorizontal(lipgloss.Top, l, v)
	}

	// ── Info fields ─────────────────────────────────────────────────────────
	lines = append(lines, SectionTitleStyle.Render("  Info"))
	lines = append(lines, DividerStyle.Render("  "+strings.Repeat("·", w-4)))
	lines = append(lines, field("Description", info.Desc))
	lines = append(lines, field("Homepage", info.Homepage))
	lines = append(lines, field("License", info.License))
	lines = append(lines, field("Tap", info.Tap))
	lines = append(lines, field("Formula", info.FullName))
	lines = append(lines, "")

	// ── Versions ────────────────────────────────────────────────────────────
	lines = append(lines, SectionTitleStyle.Render("  Versions"))
	lines = append(lines, DividerStyle.Render("  "+strings.Repeat("·", w-4)))
	lines = append(lines, field("Latest", info.Version))

	ivs := make([]string, 0, len(info.Installed))
	for _, iv := range info.Installed {
		ivs = append(ivs, iv.Version)
	}
	if len(ivs) > 0 {
		lines = append(lines, field("Installed", strings.Join(ivs, ", ")))
	}
	if info.InstalledSize != "" {
		lines = append(lines, field("Size", info.InstalledSize))
	}
	lines = append(lines, "")

	// ── Dependencies ────────────────────────────────────────────────────────
	lines = append(lines, SectionTitleStyle.Render("  Dependencies"))
	lines = append(lines, DividerStyle.Render("  "+strings.Repeat("·", w-4)))
	if len(info.Dependencies) == 0 {
		lines = append(lines, MutedStyle.Render("  None"))
	} else {
		// Wrap deps into rows of 3
		row := ""
		for i, dep := range info.Dependencies {
			row += ValueStyle.Render("• " + dep + "  ")
			if (i+1)%4 == 0 {
				lines = append(lines, "  "+row)
				row = ""
			}
		}
		if row != "" {
			lines = append(lines, "  "+row)
		}
	}
	lines = append(lines, "")

	// ── Conflicts ───────────────────────────────────────────────────────────
	if len(info.Conflicts) > 0 {
		lines = append(lines, SectionTitleStyle.Render("  Conflicts With"))
		lines = append(lines, DividerStyle.Render("  "+strings.Repeat("·", w-4)))
		for _, c := range info.Conflicts {
			lines = append(lines, "  "+WarningStyle.Render("⚠  "+c))
		}
		lines = append(lines, "")
	}

	// ── Actions hint ────────────────────────────────────────────────────────
	lines = append(lines, DividerStyle.Render(strings.Repeat("─", w)))
	lines = append(lines,
		MutedStyle.Render("  ")+
			FooterKeyStyle.Render("u")+MutedStyle.Render(" upgrade  ")+
			FooterKeyStyle.Render("x")+MutedStyle.Render(" uninstall  ")+
			FooterKeyStyle.Render("esc")+MutedStyle.Render(" back"),
	)

	return lines
}
