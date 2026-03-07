package tui

import (
	"database/sql"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type view int

const (
	topView view = iota
	anomalyView
	miscView
)

type Model struct {
	db          *sql.DB
	activeView  view
	top         TopModel
	anomaly     AnomalyModel
	misc        MiscModel
	lastRefresh time.Time
	width       int
	height      int
}

func New(db *sql.DB) Model {
	return Model{
		db:         db,
		activeView: topView,
		top:        NewTopModel(db),
		anomaly:    NewAnomalyModel(db),
		misc:       NewMiscModel(db),
	}
}

func (m Model) Init() tea.Cmd {
	return m.top.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "t":
			m.activeView = topView
			return m, m.top.Init()
		case "a":
			m.activeView = anomalyView
			return m, m.anomaly.Init()
		case "m":
			m.activeView = miscView
			return m, m.misc.Init()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case refreshMsg:
		m.lastRefresh = time.Now()
	}

	// Delegate to active view
	var cmd tea.Cmd
	switch m.activeView {
	case topView:
		m.top, cmd = m.top.Update(msg)
	case anomalyView:
		m.anomaly, cmd = m.anomaly.Update(msg)
	case miscView:
		m.misc, cmd = m.misc.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	var content string
	switch m.activeView {
	case topView:
		content = m.top.View()
	case anomalyView:
		content = m.anomaly.View()
	case miscView:
		content = m.misc.View()
	}

	return fmt.Sprintf("%s\n%s", content, m.statusBar())
}

func (m Model) statusBar() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(0, 1).
		Width(m.width)

	refresh := "never"
	if !m.lastRefresh.IsZero() {
		refresh = m.lastRefresh.Format("15:04:05")
	}

	return style.Render(fmt.Sprintf(
		"[t]op  [a]nomalies  [m]isc  [q]uit  │  last refresh: %s",
		refresh,
	))
}

type refreshMsg struct{}
