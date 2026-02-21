package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DoctorModel shows the output of `brew doctor`.
type DoctorModel struct {
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
	d.scroll = 0
	d.hasErr = err != nil
	// Clean up and split
	d.lines = strings.Split(strings.TrimRight(out, "\n"), "\n")
}

func (d DoctorModel) Update(msg tea.Msg) (DoctorModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		visibleLines := d.height - headerHeight - footerHeight - 3
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
		case "home", "g":
			d.scroll = 0
		case "end", "G":
			d.scroll = maxScroll
		}
	}
	return d, nil
}

func (d DoctorModel) View(height int) string {
	var sb strings.Builder

	// ── Title ─────────────────────────────────────────────────────────────
	if d.hasErr {
		sb.WriteString(WarningStyle.Render("  ⚠  brew doctor — issues found"))
	} else {
		sb.WriteString(SuccessStyle.Render("  ✓  brew doctor — your system is ready to brew!"))
	}
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", d.width)))
	sb.WriteString("\n")

	headerLines := 2
	visibleLines := height - headerLines - 1 // -1 for scroll bar
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

	for _, line := range d.lines[d.scroll:end] {
		switch {
		case strings.HasPrefix(line, "Warning:"):
			sb.WriteString(WarningStyle.Render("  "+line) + "\n")
		case strings.HasPrefix(line, "Error:"):
			sb.WriteString(ErrorStyle.Render("  "+line) + "\n")
		case strings.HasPrefix(line, "==>"):
			sb.WriteString(SectionTitleStyle.Render("  "+line) + "\n")
		case line == "":
			sb.WriteString("\n")
		default:
			sb.WriteString(ValueStyle.Render("  "+line) + "\n")
		}
	}

	// ── Scroll bar ────────────────────────────────────────────────────────
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", d.width)) + "\n")
	if len(d.lines) > visibleLines && maxScroll > 0 {
		pct := (d.scroll * 100) / maxScroll
		indicator := lipgloss.JoinHorizontal(lipgloss.Top,
			MutedStyle.Render("  "),
			FooterKeyStyle.Render("↑↓"), MutedStyle.Render(" scroll  "),
			FooterKeyStyle.Render("g/G"), MutedStyle.Render(" top/bottom  "),
			StatBarStyle.Render(strings.Repeat("─", 10)),
			StatHighlightStyle.Render(padRight("", pct/10)+"▌"),
			MutedStyle.Render(strings.Repeat("─", 10-pct/10)),
			MutedStyle.Render(lipgloss.Place(8, 1, lipgloss.Right, lipgloss.Top,
				StatBarStyle.Render(""),
			)),
		)
		sb.WriteString(indicator)
	} else {
		sb.WriteString(MutedStyle.Render("  end of output"))
	}

	return sb.String()
}
