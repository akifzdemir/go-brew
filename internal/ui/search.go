package ui

import (
	"fmt"
	"strings"

	"github.com/akif/gobrew/internal/brew"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SearchModel handles the search & install view.
type SearchModel struct {
	input   textinput.Model
	results []brew.Package
	cursor  int
	offset  int
	width   int
	height  int
	// When true, the text input is active; otherwise the results list is active.
	focusInput bool
}

func newSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search for a formula..."
	ti.CharLimit = 128
	ti.Width = 40

	return SearchModel{
		input:      ti,
		focusInput: true,
	}
}

func (s *SearchModel) setSize(w, h int) {
	s.width = w
	s.height = h
	s.input.Width = w - 6
}

func (s *SearchModel) setResults(pkgs []brew.Package) {
	s.results = pkgs
	s.cursor = 0
	s.offset = 0
	s.focusInput = false // move focus to results list
}

func (s *SearchModel) focus() {
	s.focusInput = true
	s.input.Focus()
	s.results = nil
	s.cursor = 0
}

func (s SearchModel) inputFocused() bool {
	return s.focusInput
}

func (s SearchModel) query() string {
	return strings.TrimSpace(s.input.Value())
}

func (s SearchModel) selectedPackage() *brew.Package {
	if len(s.results) == 0 {
		return nil
	}
	if s.cursor < 0 || s.cursor >= len(s.results) {
		return nil
	}
	return &s.results[s.cursor]
}

func (s SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	var cmds []tea.Cmd

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "tab":
			s.focusInput = !s.focusInput
			if s.focusInput {
				s.input.Focus()
			} else {
				s.input.Blur()
			}
			return s, nil

		case "up", "k":
			if !s.focusInput && s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if !s.focusInput && s.cursor < len(s.results)-1 {
				s.cursor++
			}
		}
	}

	if s.focusInput {
		var cmd tea.Cmd
		s.input, cmd = s.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return s, tea.Batch(cmds...)
}

func (s SearchModel) View(height int) string {
	var sb strings.Builder

	// Title
	sb.WriteString(PanelTitleStyle.Render("  Search & Install"))
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", s.width)))
	sb.WriteString("\n")

	// Input box
	inputLabel := LabelStyle.Render("  Formula: ")
	sb.WriteString(inputLabel + s.input.View())
	sb.WriteString("\n")
	sb.WriteString(MutedStyle.Render("  Press Enter to search · Tab to switch focus · Enter on result to install"))
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", s.width)))
	sb.WriteString("\n")

	// Results
	headerLines := 5
	visibleRows := height - headerLines
	if visibleRows < 1 {
		visibleRows = 1
	}

	if len(s.results) == 0 {
		sb.WriteString(MutedStyle.Render("  Type a name above and press Enter to search."))
		return sb.String()
	}

	// Scroll
	if s.cursor < s.offset {
		s.offset = s.cursor
	}
	if s.cursor >= s.offset+visibleRows {
		s.offset = s.cursor - visibleRows + 1
	}

	end := s.offset + visibleRows
	if end > len(s.results) {
		end = len(s.results)
	}

	for i := s.offset; i < end; i++ {
		pkg := s.results[i]
		name := padRight(pkg.Name, 30)
		desc := pkg.LatestVersion // repurposed as description
		if len(desc) > 50 {
			desc = desc[:49] + "…"
		}
		row := "  " + name + MutedStyle.Render(desc)
		if i == s.cursor && !s.focusInput {
			row = RowSelectedStyle.Render("  " + padRight(pkg.Name, 30) + padRight(desc, 50))
		}
		sb.WriteString(row + "\n")
	}

	if len(s.results) > visibleRows {
		hint := fmt.Sprintf("  %d–%d of %d results", s.offset+1, end, len(s.results))
		sb.WriteString(MutedStyle.Render(hint))
	}

	return sb.String()
}
