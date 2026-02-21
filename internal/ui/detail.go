package ui

import (
	"fmt"
	"strings"

	"github.com/akif/gobrew/internal/brew"
	tea "github.com/charmbracelet/bubbletea"
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
	d.lines = buildDetailLines(info)
}

func (d *DetailModel) setSize(w, h int) {
	d.width = w
	d.height = h
}

func (d DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		visibleLines := d.height - 4
		switch key.String() {
		case "up", "k":
			if d.scroll > 0 {
				d.scroll--
			}
		case "down", "j":
			if d.scroll < len(d.lines)-visibleLines {
				d.scroll++
			}
		case "pgup":
			d.scroll -= visibleLines
			if d.scroll < 0 {
				d.scroll = 0
			}
		case "pgdown":
			d.scroll += visibleLines
			if d.scroll > len(d.lines)-visibleLines {
				d.scroll = max(0, len(d.lines)-visibleLines)
			}
		}
	}
	return d, nil
}

func (d DetailModel) View(height int) string {
	if d.info == nil {
		return MutedStyle.Render("  No package selected.")
	}

	visibleLines := height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	end := d.scroll + visibleLines
	if end > len(d.lines) {
		end = len(d.lines)
	}

	visible := d.lines[d.scroll:end]
	body := strings.Join(visible, "\n")

	// Scroll hint
	if len(d.lines) > visibleLines {
		hint := fmt.Sprintf("  line %d/%d  ↑↓ scroll", d.scroll+1, len(d.lines))
		body += "\n" + MutedStyle.Render(hint)
	}

	return body
}

func buildDetailLines(info *brew.InfoResult) []string {
	var lines []string

	add := func(label, value string) {
		lines = append(lines, LabelStyle.Render(label+": ")+ValueStyle.Render(value))
	}

	lines = append(lines, PanelTitleStyle.Render("  "+info.Name))
	lines = append(lines, DividerStyle.Render(strings.Repeat("─", 60)))

	add("  Full name  ", info.FullName)
	add("  Tap        ", info.Tap)
	add("  Description", info.Desc)
	add("  Homepage   ", info.Homepage)
	add("  License    ", info.License)
	add("  Version    ", info.Version)

	// Installed versions
	ivs := make([]string, 0, len(info.Installed))
	for _, iv := range info.Installed {
		ivs = append(ivs, iv.Version)
	}
	if len(ivs) > 0 {
		add("  Installed  ", strings.Join(ivs, ", "))
	}

	// Dependencies
	lines = append(lines, "")
	if len(info.Dependencies) > 0 {
		lines = append(lines, LabelStyle.Render("  Dependencies:"))
		for _, dep := range info.Dependencies {
			lines = append(lines, "    "+ValueStyle.Render("• "+dep))
		}
	} else {
		lines = append(lines, MutedStyle.Render("  No dependencies."))
	}

	// Conflicts
	if len(info.Conflicts) > 0 {
		lines = append(lines, "")
		lines = append(lines, LabelStyle.Render("  Conflicts with:"))
		for _, c := range info.Conflicts {
			lines = append(lines, "    "+WarningStyle.Render("• "+c))
		}
	}

	lines = append(lines, "")
	lines = append(lines, MutedStyle.Render("  esc back  u upgrade  x uninstall"))

	return lines
}
