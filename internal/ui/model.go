package ui

import (
	"github.com/akif/go-brew/internal/brew"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// View represents which screen is currently active.
type View int

const (
	ViewTable  View = iota // main package list
	ViewDetail             // package detail panel
	ViewSearch             // search & install
	ViewDoctor             // brew doctor output
)

// --- Messages ---

// PackagesLoadedMsg is sent when the initial package list has been fetched.
type PackagesLoadedMsg struct {
	Packages []brew.Package
	Err      error
}

// InfoLoadedMsg carries the result of `brew info`.
type InfoLoadedMsg struct {
	Info *brew.InfoResult
	Err  error
}

// OperationDoneMsg is sent when an upgrade / uninstall / install finishes.
type OperationDoneMsg struct {
	Output string
	Err    error
}

// SearchResultMsg carries search results.
type SearchResultMsg struct {
	Results []brew.Package
	Err     error
}

// DoctorDoneMsg carries the output of `brew doctor`.
type DoctorDoneMsg struct {
	Output string
	Err    error
}

// --- Root Model ---

// Model is the top-level Bubble Tea model.
type Model struct {
	// Current active view
	activeView View

	// Terminal dimensions
	width  int
	height int

	// Sub-models
	tableModel  TableModel
	detailModel DetailModel
	searchModel SearchModel
	doctorModel DoctorModel

	// Global loading spinner (used during initial load)
	spinner    spinner.Model
	loading    bool
	loadingMsg string

	// Global status message shown in the footer
	statusMsg   string
	statusIsErr bool
}

// New creates and initialises the root model.
func New() Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	m := Model{
		activeView: ViewTable,
		spinner:    sp,
		loading:    true,
		loadingMsg: "Loading packages...",
	}
	m.tableModel = newTableModel()
	m.searchModel = newSearchModel()
	return m
}

// Init starts the spinner and triggers the initial package load.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadPackagesCmd(),
	)
}

// Update is the central message dispatcher.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tableModel.setSize(m.width, m.height)
		m.detailModel.setSize(m.width, m.height)
		m.searchModel.setSize(m.width, m.height)
		m.doctorModel.setSize(m.width, m.height)

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	// --- data loading ---

	case PackagesLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.statusMsg = "Error loading packages: " + msg.Err.Error()
			m.statusIsErr = true
		} else {
			m.tableModel.setPackages(msg.Packages)
			m.statusMsg = ""
		}

	case InfoLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.statusMsg = "Error loading info: " + msg.Err.Error()
			m.statusIsErr = true
			m.activeView = ViewTable
		} else {
			m.detailModel.setInfo(msg.Info)
			m.activeView = ViewDetail
		}

	case OperationDoneMsg:
		m.loading = false
		if msg.Err != nil {
			m.statusMsg = "Error: " + msg.Err.Error()
			m.statusIsErr = true
		} else {
			m.statusMsg = "Done."
			m.statusIsErr = false
		}
		// Refresh the package list after any mutation
		cmds = append(cmds, loadPackagesCmd())
		m.loading = true
		m.loadingMsg = "Refreshing packages..."
		cmds = append(cmds, m.spinner.Tick)

	case SearchResultMsg:
		m.loading = false
		if msg.Err != nil {
			m.statusMsg = "Search error: " + msg.Err.Error()
			m.statusIsErr = true
		} else {
			m.searchModel.setResults(msg.Results)
		}

	case DoctorDoneMsg:
		m.loading = false
		m.doctorModel.setOutput(msg.Output, msg.Err)
		m.activeView = ViewDoctor

	// --- keyboard ---

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Escape always goes back to table
		if msg.String() == "esc" && m.activeView != ViewTable {
			m.activeView = ViewTable
			m.statusMsg = ""
			return m, nil
		}

		switch m.activeView {
		case ViewTable:
			return m.handleTableKey(msg, cmds)
		case ViewDetail:
			return m.handleDetailKey(msg, cmds)
		case ViewSearch:
			return m.handleSearchKey(msg, cmds)
		case ViewDoctor:
			return m.handleDoctorKey(msg, cmds)
		}
	}

	// Propagate updates to active sub-model
	switch m.activeView {
	case ViewDetail:
		var cmd tea.Cmd
		m.detailModel, cmd = m.detailModel.Update(msg)
		cmds = append(cmds, cmd)
	case ViewSearch:
		var cmd tea.Cmd
		m.searchModel, cmd = m.searchModel.Update(msg)
		cmds = append(cmds, cmd)
	case ViewDoctor:
		var cmd tea.Cmd
		m.doctorModel, cmd = m.doctorModel.Update(msg)
		cmds = append(cmds, cmd)
	default:
		var cmd tea.Cmd
		m.tableModel, cmd = m.tableModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleTableKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "r":
		m.loading = true
		m.loadingMsg = "Refreshing packages..."
		cmds = append(cmds, m.spinner.Tick, loadPackagesCmd())

	case "enter":
		if pkg := m.tableModel.selectedPackage(); pkg != nil {
			m.loading = true
			m.loadingMsg = "Loading info for " + pkg.Name + "..."
			cmds = append(cmds, m.spinner.Tick, loadInfoCmd(pkg.Name))
		}

	case "u":
		if pkg := m.tableModel.selectedPackage(); pkg != nil && pkg.Outdated {
			m.loading = true
			m.loadingMsg = "Upgrading " + pkg.Name + "..."
			cmds = append(cmds, m.spinner.Tick, upgradeCmd(pkg.Name))
		}

	case "U":
		m.loading = true
		m.loadingMsg = "Upgrading all outdated packages..."
		cmds = append(cmds, m.spinner.Tick, upgradeAllCmd())

	case "x":
		if pkg := m.tableModel.selectedPackage(); pkg != nil {
			m.tableModel.setConfirm(true)
		}

	case "y":
		if m.tableModel.confirmPending {
			pkg := m.tableModel.selectedPackage()
			m.tableModel.setConfirm(false)
			if pkg != nil {
				m.loading = true
				m.loadingMsg = "Uninstalling " + pkg.Name + "..."
				cmds = append(cmds, m.spinner.Tick, uninstallCmd(pkg.Name))
			}
		}

	case "n":
		m.tableModel.setConfirm(false)

	case "/":
		m.activeView = ViewSearch
		m.searchModel.focus()

	case "d":
		m.loading = true
		m.loadingMsg = "Running brew doctor..."
		cmds = append(cmds, m.spinner.Tick, doctorCmd())
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleDetailKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "u":
		if m.detailModel.info != nil && m.tableModel.selectedPackage() != nil {
			pkg := m.tableModel.selectedPackage()
			if pkg.Outdated {
				m.loading = true
				m.loadingMsg = "Upgrading " + pkg.Name + "..."
				m.activeView = ViewTable
				cmds = append(cmds, m.spinner.Tick, upgradeCmd(pkg.Name))
			}
		}
	case "x":
		if m.detailModel.info != nil {
			name := m.detailModel.info.Name
			m.loading = true
			m.loadingMsg = "Uninstalling " + name + "..."
			m.activeView = ViewTable
			cmds = append(cmds, m.spinner.Tick, uninstallCmd(name))
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) handleSearchKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		if !m.searchModel.inputFocused() {
			return m, tea.Quit
		}
	case "enter":
		if m.searchModel.inputFocused() {
			query := m.searchModel.query()
			if query != "" {
				m.loading = true
				m.loadingMsg = "Searching for \"" + query + "\"..."
				cmds = append(cmds, m.spinner.Tick, searchCmd(query))
			}
		} else {
			// Install selected result
			if pkg := m.searchModel.selectedPackage(); pkg != nil {
				m.loading = true
				m.loadingMsg = "Installing " + pkg.Name + "..."
				m.activeView = ViewTable
				cmds = append(cmds, m.spinner.Tick, installCmd(pkg.Name))
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) handleDoctorKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	}
	return m, tea.Batch(cmds...)
}

// View renders the current screen.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	// Reserve rows for header (1) + footer (1)
	contentHeight := m.height - 2

	var body string
	if m.loading {
		body = m.renderLoading(contentHeight)
	} else {
		switch m.activeView {
		case ViewTable:
			body = m.tableModel.View(contentHeight)
		case ViewDetail:
			body = m.detailModel.View(contentHeight)
		case ViewSearch:
			body = m.searchModel.View(contentHeight)
		case ViewDoctor:
			body = m.doctorModel.View(contentHeight)
		}
	}

	return header + "\n" + body + "\n" + footer
}

func (m Model) renderHeader() string {
	title := AppTitleStyle.Render(" go-brew ")
	version := MutedStyle.Render("homebrew manager")
	space := m.width - lipglossWidth(title) - lipglossWidth(version) - 4
	if space < 1 {
		space = 1
	}
	line := title + repeatStr(" ", space) + version + "  "
	return HeaderStyle.Width(m.width).Render(line)
}

func (m Model) renderFooter() string {
	var hint string

	if m.loading {
		hint = m.spinner.View() + " " + m.loadingMsg
		return FooterStyle.Width(m.width).Render(hint)
	}

	if m.statusMsg != "" {
		if m.statusIsErr {
			hint = ErrorStyle.Render("✗ " + m.statusMsg)
		} else {
			hint = SuccessStyle.Render("✓ " + m.statusMsg)
		}
		return FooterStyle.Width(m.width).Render(hint)
	}

	switch m.activeView {
	case ViewTable:
		if m.tableModel.confirmPending {
			hint = WarningStyle.Render("Uninstall? ") +
				FooterKeyStyle.Render("y") + FooterDescStyle.Render(" yes  ") +
				FooterKeyStyle.Render("n") + FooterDescStyle.Render(" no")
		} else {
			hint = key("↑↓", "navigate") + sep() +
				key("enter", "info") + sep() +
				key("u", "upgrade") + sep() +
				key("U", "upgrade all") + sep() +
				key("x", "uninstall") + sep() +
				key("/", "search") + sep() +
				key("d", "doctor") + sep() +
				key("r", "refresh") + sep() +
				key("q", "quit")
		}
	case ViewDetail:
		hint = key("esc", "back") + sep() +
			key("u", "upgrade") + sep() +
			key("x", "uninstall") + sep() +
			key("q", "quit")
	case ViewSearch:
		hint = key("esc", "back") + sep() +
			key("tab", "toggle input/list") + sep() +
			key("enter", "search/install") + sep() +
			key("q", "quit")
	case ViewDoctor:
		hint = key("esc", "back") + sep() +
			key("↑↓", "scroll") + sep() +
			key("q", "quit")
	}

	return FooterStyle.Width(m.width).Render(hint)
}

func (m Model) renderLoading(height int) string {
	// Center vertically
	pad := height / 2
	line := SpinnerStyle.Render(m.spinner.View()) + "  " + m.loadingMsg
	out := repeatStr("\n", pad) + centerStr(line, m.width)
	return out
}

// --- tea.Cmd factories ---

func loadPackagesCmd() tea.Cmd {
	return func() tea.Msg {
		pkgs, err := brew.ListInstalled()
		return PackagesLoadedMsg{Packages: pkgs, Err: err}
	}
}

func loadInfoCmd(name string) tea.Cmd {
	return func() tea.Msg {
		info, err := brew.GetInfo(name)
		return InfoLoadedMsg{Info: info, Err: err}
	}
}

func upgradeCmd(name string) tea.Cmd {
	return func() tea.Msg {
		out, err := brew.Upgrade(name)
		return OperationDoneMsg{Output: out, Err: err}
	}
}

func upgradeAllCmd() tea.Cmd {
	return func() tea.Msg {
		out, err := brew.UpgradeAll()
		return OperationDoneMsg{Output: out, Err: err}
	}
}

func uninstallCmd(name string) tea.Cmd {
	return func() tea.Msg {
		out, err := brew.Uninstall(name)
		return OperationDoneMsg{Output: out, Err: err}
	}
}

func searchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := brew.Search(query)
		return SearchResultMsg{Results: results, Err: err}
	}
}

func installCmd(name string) tea.Cmd {
	return func() tea.Msg {
		out, err := brew.Install(name)
		return OperationDoneMsg{Output: out, Err: err}
	}
}

func doctorCmd() tea.Cmd {
	return func() tea.Msg {
		out, err := brew.Doctor()
		return DoctorDoneMsg{Output: out, Err: err}
	}
}

// --- misc helpers ---

func key(k, desc string) string {
	return FooterKeyStyle.Render(k) + FooterDescStyle.Render(" "+desc)
}

func sep() string {
	return MutedStyle.Render("  ·  ")
}

func repeatStr(s string, n int) string {
	if n <= 0 {
		return ""
	}
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func centerStr(s string, width int) string {
	w := lipglossWidth(s)
	pad := (width - w) / 2
	if pad < 0 {
		pad = 0
	}
	return repeatStr(" ", pad) + s
}

func lipglossWidth(s string) int {
	return len([]rune(stripANSI(s)))
}

// stripANSI removes ANSI escape codes for width calculation.
func stripANSI(s string) string {
	var result []rune
	inEscape := false
	for _, r := range s {
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		if r == '\x1b' {
			inEscape = true
			continue
		}
		result = append(result, r)
	}
	return string(result)
}
