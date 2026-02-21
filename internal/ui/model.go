package ui

import (
	"strings"

	"github.com/akif/gobrew/internal/brew"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View represents which screen is currently active.
type View int

const (
	ViewTable View = iota
	ViewDetail
	ViewSearch
	ViewDoctor
)

// ---------------------------------------------------------------------------
// Messages
// ---------------------------------------------------------------------------

type PackagesLoadedMsg struct {
	Packages []brew.Package
	Err      error
}

type InfoLoadedMsg struct {
	Info *brew.InfoResult
	Err  error
}

type OperationDoneMsg struct {
	Output string
	Err    error
}

type SearchResultMsg struct {
	Results []brew.Package
	Err     error
}

type DoctorDoneMsg struct {
	Output string
	Err    error
}

// ---------------------------------------------------------------------------
// Root model
// ---------------------------------------------------------------------------

type Model struct {
	activeView View
	width      int
	height     int

	tableModel  TableModel
	detailModel DetailModel
	searchModel SearchModel
	doctorModel DoctorModel

	spinner    spinner.Model
	loading    bool
	loadingMsg string

	statusMsg   string
	statusIsErr bool
}

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

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadPackagesCmd(),
	)
}

// ---------------------------------------------------------------------------
// Update — BUG FIX: sub-models are always updated via their own handlers
// ---------------------------------------------------------------------------

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tableModel.setSize(m.width, m.height)
		m.detailModel.setSize(m.width, m.height)
		m.searchModel.setSize(m.width, m.height)
		m.doctorModel.setSize(m.width, m.height)
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case PackagesLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.statusMsg = "Error loading packages: " + msg.Err.Error()
			m.statusIsErr = true
		} else {
			m.tableModel.setPackages(msg.Packages)
			m.statusMsg = ""
		}
		return m, nil

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
		return m, nil

	case OperationDoneMsg:
		m.loading = false
		if msg.Err != nil {
			m.statusMsg = "Error: " + msg.Err.Error()
			m.statusIsErr = true
		} else {
			m.statusMsg = "Done."
			m.statusIsErr = false
		}
		// Refresh after mutation
		m.loading = true
		m.loadingMsg = "Refreshing..."
		return m, tea.Batch(m.spinner.Tick, loadPackagesCmd())

	case SearchResultMsg:
		m.loading = false
		if msg.Err != nil {
			m.statusMsg = "Search error: " + msg.Err.Error()
			m.statusIsErr = true
		} else {
			m.searchModel.setResults(msg.Results)
			m.statusMsg = ""
		}
		return m, nil

	case DoctorDoneMsg:
		m.loading = false
		m.doctorModel.setOutput(msg.Output, msg.Err)
		m.activeView = ViewDoctor
		return m, nil

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Escape back to table from any sub-view
		if msg.String() == "esc" && m.activeView != ViewTable {
			m.activeView = ViewTable
			m.statusMsg = ""
			return m, nil
		}
		// While loading, ignore all keys except ctrl+c
		if m.loading {
			return m, nil
		}

		switch m.activeView {
		case ViewTable:
			return m.handleTableKey(msg)
		case ViewDetail:
			return m.handleDetailKey(msg)
		case ViewSearch:
			return m.handleSearchKey(msg)
		case ViewDoctor:
			return m.handleDoctorKey(msg)
		}
	}

	return m, nil
}

// ---------------------------------------------------------------------------
// Key handlers — BUG FIX: each handler forwards msg to its sub-model
// ---------------------------------------------------------------------------

func (m Model) handleTableKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// First let the table model handle navigation keys
	var tableCmd tea.Cmd
	m.tableModel, tableCmd = m.tableModel.Update(msg)

	var cmds []tea.Cmd
	if tableCmd != nil {
		cmds = append(cmds, tableCmd)
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "r":
		m.loading = true
		m.loadingMsg = "Refreshing packages..."
		m.statusMsg = ""
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
		if m.tableModel.selectedPackage() != nil {
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

func (m Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Forward to detail model for scrolling
	var cmd tea.Cmd
	m.detailModel, cmd = m.detailModel.Update(msg)
	var cmds []tea.Cmd
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "u":
		if m.detailModel.info != nil {
			pkg := m.tableModel.selectedPackage()
			if pkg != nil && pkg.Outdated {
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

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Forward to search model for input handling / navigation
	var subCmd tea.Cmd
	m.searchModel, subCmd = m.searchModel.Update(msg)
	var cmds []tea.Cmd
	if subCmd != nil {
		cmds = append(cmds, subCmd)
	}

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

func (m Model) handleDoctorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Forward to doctor model for scrolling
	var cmd tea.Cmd
	m.doctorModel, cmd = m.doctorModel.Update(msg)
	var cmds []tea.Cmd
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if msg.String() == "q" {
		return m, tea.Quit
	}
	return m, tea.Batch(cmds...)
}

// ---------------------------------------------------------------------------
// View — BUG FIX: loading text only appears once, in the body
// ---------------------------------------------------------------------------

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	// Fixed heights
	contentH := m.height - headerHeight - footerHeight
	if contentH < 1 {
		contentH = 1
	}

	header := m.renderHeader()
	footer := m.renderFooter() // never shows loading state — bug fixed

	var body string
	if m.loading {
		// BUG FIX: spinner only rendered here, not also in footer
		body = m.renderLoading(contentH)
	} else {
		switch m.activeView {
		case ViewTable:
			body = m.tableModel.View(contentH)
		case ViewDetail:
			body = m.detailModel.View(contentH)
		case ViewSearch:
			body = m.searchModel.View(contentH)
		case ViewDoctor:
			body = m.doctorModel.View(contentH)
		}
	}

	return header + body + "\n" + footer
}

// ---------------------------------------------------------------------------
// Header — full-width bar with title + subtitle
// ---------------------------------------------------------------------------

func (m Model) renderHeader() string {
	title := HeaderTitleStyle.Render("  gobrew")
	subtitle := HeaderSubtitleStyle.Render("homebrew manager")

	// Fill the remaining space
	used := lipgloss.Width(title) + lipgloss.Width(subtitle)
	gap := m.width - used
	if gap < 0 {
		gap = 0
	}
	filler := HeaderBarStyle.Width(gap).Render("")

	topBar := lipgloss.JoinHorizontal(lipgloss.Top, title, filler, subtitle)
	topBar = HeaderBarStyle.Width(m.width).Render(topBar)

	divider := HeaderDividerStyle.Width(m.width).Render(strings.Repeat("─", m.width))

	return topBar + "\n" + divider + "\n"
}

// ---------------------------------------------------------------------------
// Footer — keybinds on left, view name on right
// ---------------------------------------------------------------------------

func (m Model) renderFooter() string {
	var leftPart string

	// Status message takes priority over keybind hints
	if m.statusMsg != "" {
		if m.statusIsErr {
			leftPart = ErrorStyle.Render("✗ " + m.statusMsg)
		} else {
			leftPart = SuccessStyle.Render("✓ " + m.statusMsg)
		}
	} else if m.tableModel.confirmPending && m.activeView == ViewTable {
		leftPart = WarningStyle.Render("Uninstall? ") +
			FooterKeyStyle.Render("y") + FooterDescStyle.Render(" yes") +
			FooterSepStyle.Render("  /  ") +
			FooterKeyStyle.Render("n") + FooterDescStyle.Render(" no")
	} else {
		leftPart = m.footerHints()
	}

	// Right side: current view name
	var viewName string
	switch m.activeView {
	case ViewTable:
		viewName = "packages"
	case ViewDetail:
		viewName = "detail"
	case ViewSearch:
		viewName = "search"
	case ViewDoctor:
		viewName = "doctor"
	}
	rightPart := FooterSepStyle.Render(viewName + " ")

	leftW := lipgloss.Width(leftPart)
	rightW := lipgloss.Width(rightPart)
	gap := m.width - leftW - rightW - 4 // 4 = padding on both sides
	if gap < 1 {
		gap = 1
	}

	line := leftPart + strings.Repeat(" ", gap) + rightPart
	return FooterStyle.Width(m.width).Render(line)
}

func (m Model) footerHints() string {
	k := func(key, desc string) string {
		return FooterKeyStyle.Render(key) + FooterDescStyle.Render(" "+desc)
	}
	s := FooterSepStyle.Render(" · ")

	switch m.activeView {
	case ViewTable:
		return k("↑↓", "move") + s + k("enter", "info") + s +
			k("u", "upgrade") + s + k("U", "upg.all") + s +
			k("x", "uninstall") + s + k("/", "search") + s +
			k("d", "doctor") + s + k("r", "refresh") + s + k("q", "quit")
	case ViewDetail:
		return k("↑↓", "scroll") + s + k("u", "upgrade") + s +
			k("x", "uninstall") + s + k("esc", "back") + s + k("q", "quit")
	case ViewSearch:
		return k("tab", "focus") + s + k("enter", "search/install") + s +
			k("↑↓", "move") + s + k("esc", "back")
	case ViewDoctor:
		return k("↑↓", "scroll") + s + k("esc", "back") + s + k("q", "quit")
	}
	return ""
}

// ---------------------------------------------------------------------------
// Loading screen — BUG FIX: only one place renders the spinner
// ---------------------------------------------------------------------------

func (m Model) renderLoading(height int) string {
	spinLine := SpinnerStyle.Render(m.spinner.View()) + "  " +
		MutedStyle.Render(m.loadingMsg)

	return lipgloss.Place(
		m.width, height,
		lipgloss.Center, lipgloss.Center,
		spinLine,
	)
}

// ---------------------------------------------------------------------------
// tea.Cmd factories
// ---------------------------------------------------------------------------

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
