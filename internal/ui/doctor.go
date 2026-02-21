package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// DoctorModel shows the output of `brew doctor`.
type DoctorModel struct {
	output string
	lines  []string
	scroll int
	width  int
	height int
	hasErr bool
}

func (d *DoctorModel) setSize(w, h int) {
	d.width = w
	d.height = h
}

func (d *DoctorModel) setOutput(out string, err error) {
	d.output = out
	d.scroll = 0
	d.hasErr = err != nil
	d.lines = strings.Split(out, "\n")
}

func (d DoctorModel) Update(msg tea.Msg) (DoctorModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		visibleLines := d.height - 4
		if visibleLines < 1 {
			visibleLines = 1
		}
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
		case "home", "g":
			d.scroll = 0
		case "end", "G":
			d.scroll = max(0, len(d.lines)-visibleLines)
		}
	}
	return d, nil
}

func (d DoctorModel) View(height int) string {
	var sb strings.Builder

	// Title
	if d.hasErr {
		sb.WriteString(ErrorStyle.Render("  brew doctor — issues found"))
	} else {
		sb.WriteString(SuccessStyle.Render("  brew doctor — your system is ready to brew!"))
	}
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", d.width)))
	sb.WriteString("\n")

	headerLines := 2
	visibleLines := height - headerLines
	if visibleLines < 1 {
		visibleLines = 1
	}

	end := d.scroll + visibleLines
	if end > len(d.lines) {
		end = len(d.lines)
	}

	for _, line := range d.lines[d.scroll:end] {
		// Colorize warnings/errors in the output
		switch {
		case strings.HasPrefix(line, "Warning:"):
			sb.WriteString(WarningStyle.Render("  "+line) + "\n")
		case strings.HasPrefix(line, "Error:"):
			sb.WriteString(ErrorStyle.Render("  "+line) + "\n")
		case strings.HasPrefix(line, "==>"):
			sb.WriteString(LabelStyle.Render("  "+line) + "\n")
		default:
			sb.WriteString(ValueStyle.Render("  "+line) + "\n")
		}
	}

	// Scroll hint
	if len(d.lines) > visibleLines {
		hint := strings.Repeat("─", d.width)
		sb.WriteString(DividerStyle.Render(hint) + "\n")
		sb.WriteString(MutedStyle.Render(
			"  ↑↓/j/k scroll · pgup/pgdn page · g/G top/bottom · esc back",
		))
	}

	return sb.String()
}
