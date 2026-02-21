package main

import (
	"fmt"
	"os"

	"github.com/akif/gobrew/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := ui.New()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "go-brew: %v\n", err)
		os.Exit(1)
	}
}
