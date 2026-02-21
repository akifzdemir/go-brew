package ui

import (
	"fmt"
	"strings"

	"github.com/akif/gobrew/internal/brew"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchModel handles the search & install view.
type SearchModel struct {
	input      textinput.Model
	results    []brew.Package
	cursor     int
	offset     int
	width      int
	height     int
	focusInput bool // true = typing in input; false = navigating results
}

func newSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Type a formula name and press Enter..."
	ti.CharLimit = 128
	ti.Width = 40
	ti.PromptStyle = SearchPromptStyle
	ti.TextStyle = ValueStyle

	return SearchModel{
		input:      ti,
		focusInput: true,
	}
}

func (s *SearchModel) setSize(w, h int) {
	s.width = w
	s.height = h
	s.input.Width = w - 16 // leave room for label + border padding
}

func (s *SearchModel) setResults(pkgs []brew.Package) {
	s.results = pkgs
	s.cursor = 0
	s.offset = 0
	s.focusInput = false
}

func (s *SearchModel) focus() {
	s.focusInput = true
	s.input.Focus()
	s.results = nil
	s.cursor = 0
	s.offset = 0
}

func (s SearchModel) inputFocused() bool { return s.focusInput }
func (s SearchModel) query() string      { return strings.TrimSpace(s.input.Value()) }

func (s SearchModel) selectedPackage() *brew.Package {
	if len(s.results) == 0 || s.cursor < 0 || s.cursor >= len(s.results) {
		return nil
	}
	return &s.results[s.cursor]
}

// Update — navigation handled here; textinput update delegated below.
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

	// ── Title ────────────────────────────────────────────────────────────────
	sb.WriteString(PanelTitleStyle.Render("  Search & Install"))
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", s.width)))
	sb.WriteString("\n")

	// ── Input row ────────────────────────────────────────────────────────────
	var inputBoxStyle lipgloss.Style
	if s.focusInput {
		inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(0, 1).
			Width(s.width - 4)
	} else {
		inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			Width(s.width - 4)
	}
	sb.WriteString("  " + inputBoxStyle.Render(s.input.View()))
	sb.WriteString("\n")

	// ── Hint line ────────────────────────────────────────────────────────────
	hint := MutedStyle.Render("  ") +
		FooterKeyStyle.Render("Enter") + MutedStyle.Render(" search  ") +
		FooterKeyStyle.Render("Tab") + MutedStyle.Render(" switch focus  ") +
		FooterKeyStyle.Render("Enter") + MutedStyle.Render(" on result to install  ") +
		FooterKeyStyle.Render("Esc") + MutedStyle.Render(" back")
	sb.WriteString(hint + "\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", s.width)) + "\n")

	// ── Results ──────────────────────────────────────────────────────────────
	headerLines := 5
	visibleRows := height - headerLines
	if visibleRows < 1 {
		visibleRows = 1
	}

	if len(s.results) == 0 {
		empty := lipgloss.Place(s.width, visibleRows,
			lipgloss.Center, lipgloss.Center,
			MutedStyle.Render("Type a formula name above and press Enter to search."),
		)
		sb.WriteString(empty)
		return sb.String()
	}

	// Result count badge
	countBadge := SearchResultCountStyle.Render(
		fmt.Sprintf("  %d result(s)", len(s.results)),
	)
	sb.WriteString(countBadge + "\n")
	visibleRows-- // one row used by count badge

	// Scroll offset
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

	colName := 32
	colDesc := s.width - colName - 4
	if colDesc < 10 {
		colDesc = 10
	}

	for i := s.offset; i < end; i++ {
		pkg := s.results[i]
		name := padRight(pkg.Name, colName)
		desc := pkg.LatestVersion // repurposed as description in search results
		if len([]rune(desc)) > colDesc {
			rr := []rune(desc)
			desc = string(rr[:colDesc-1]) + "…"
		}

		if i == s.cursor && !s.focusInput {
			row := "  " + padRight(pkg.Name, colName) + padRight(desc, colDesc)
			sb.WriteString(RowSelectedStyle.Width(s.width).Render(row) + "\n")
		} else {
			row := "  " + name + MutedStyle.Render(desc)
			sb.WriteString(row + "\n")
		}
	}

	// Scroll hint
	if len(s.results) > visibleRows {
		sb.WriteString(MutedStyle.Render(
			fmt.Sprintf("  showing %d–%d of %d", s.offset+1, end, len(s.results)),
		))
	}

	return sb.String()
}
