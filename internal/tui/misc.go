package tui

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type miscTab int

const (
	frequentView miscTab = iota
	zombieView
	longRunningView
)

type MiscModel struct {
	db          *sql.DB
	activeView  miscTab
	frequent    table.Model
	zombies     table.Model
	longRunning table.Model
	err         error
}

type frequentProcess struct {
	name     string
	runCount int
	lastSeen string
}

type zombieProcess struct {
	name string
	pid  int32
}

type longRunningProcess struct {
	name     string
	pid      int32
	duration string
}

type frequentLoadedMsg struct{ processes []frequentProcess }
type zombiesLoadedMsg struct{ processes []zombieProcess }
type longRunningLoadedMsg struct{ processes []longRunningProcess }

func NewMiscModel(db *sql.DB) MiscModel {
	frequent := table.New(
		table.WithColumns([]table.Column{
			{Title: "NAME", Width: 24},
			{Title: "RUN COUNT", Width: 12},
			{Title: "LAST SEEN", Width: 20},
		}),
		table.WithFocused(true),
		table.WithHeight(18),
	)
	frequent.SetStyles(tableStyles())

	zombies := table.New(
		table.WithColumns([]table.Column{
			{Title: "NAME", Width: 24},
			{Title: "PID", Width: 10},
		}),
		table.WithFocused(true),
		table.WithHeight(18),
	)
	zombies.SetStyles(tableStyles())

	longRunning := table.New(
		table.WithColumns([]table.Column{
			{Title: "NAME", Width: 24},
			{Title: "PID", Width: 10},
			{Title: "RUNNING FOR", Width: 20},
		}),
		table.WithFocused(true),
		table.WithHeight(18),
	)
	longRunning.SetStyles(tableStyles())

	return MiscModel{
		db:          db,
		activeView:  frequentView,
		frequent:    frequent,
		zombies:     zombies,
		longRunning: longRunning,
	}
}

func (m MiscModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadFrequent(),
		m.loadZombies(),
		m.loadLongRunning(),
		tickCmd(),
	)
}

func (m MiscModel) Update(msg tea.Msg) (MiscModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.activeView = frequentView
		case "2":
			m.activeView = zombieView
		case "3":
			m.activeView = longRunningView
		case "up", "down":
			var cmd tea.Cmd
			switch m.activeView {
			case frequentView:
				m.frequent, cmd = m.frequent.Update(msg)
			case zombieView:
				m.zombies, cmd = m.zombies.Update(msg)
			case longRunningView:
				m.longRunning, cmd = m.longRunning.Update(msg)
			}
			return m, cmd
		}
	case frequentLoadedMsg:
		rows := make([]table.Row, len(msg.processes))
		for i, p := range msg.processes {
			rows[i] = table.Row{p.name, fmt.Sprintf("%d", p.runCount), p.lastSeen}
		}
		m.frequent.SetRows(rows)
	case zombiesLoadedMsg:
		rows := make([]table.Row, len(msg.processes))
		for i, p := range msg.processes {
			rows[i] = table.Row{p.name, fmt.Sprintf("%d", p.pid)}
		}
		m.zombies.SetRows(rows)
	case longRunningLoadedMsg:
		rows := make([]table.Row, len(msg.processes))
		for i, p := range msg.processes {
			rows[i] = table.Row{p.name, fmt.Sprintf("%d", p.pid), p.duration}
		}
		m.longRunning.SetRows(rows)
	case tickMsg:
		return m, tea.Batch(
			m.loadFrequent(),
			m.loadZombies(),
			m.loadLongRunning(),
			tickCmd(),
		)
	case errMsg:
		m.err = msg.err
	}
	return m, nil
}

func (m MiscModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v", m.err)
	}

	tabs := "[1] Frequent  [2] Zombies  [3] Long Running"
	tabStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("213"))

	header := tabStyle.Render(fmt.Sprintf("  Misc — %s", tabs))

	var content string
	switch m.activeView {
	case frequentView:
		content = m.frequent.View()
	case zombieView:
		content = m.zombies.View()
	case longRunningView:
		content = m.longRunning.View()
	}

	return fmt.Sprintf("%s\n%s", header, content)
}

func (m MiscModel) loadFrequent() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.db.Query(`
			SELECT 
				name,
				COUNT(*) as run_count,
				MAX(captured_at) as last_seen
			FROM processes
			WHERE captured_at > datetime('now', '-24 hours')
			GROUP BY name
			ORDER BY run_count DESC
			LIMIT 20
		`)
		if err != nil {
			return errMsg{err}
		}
		defer rows.Close()

		var processes []frequentProcess
		for rows.Next() {
			var p frequentProcess
			if err := rows.Scan(&p.name, &p.runCount, &p.lastSeen); err != nil {
				return errMsg{err}
			}
			processes = append(processes, p)
		}
		return frequentLoadedMsg{processes}
	}
}

func (m MiscModel) loadZombies() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.db.Query(`
			SELECT name, pid
			FROM processes
			WHERE status = 'zombie'
			AND captured_at = (SELECT MAX(captured_at) FROM processes)
		`)
		if err != nil {
			return errMsg{err}
		}
		defer rows.Close()

		var processes []zombieProcess
		for rows.Next() {
			var p zombieProcess
			if err := rows.Scan(&p.name, &p.pid); err != nil {
				return errMsg{err}
			}
			processes = append(processes, p)
		}
		return zombiesLoadedMsg{processes}
	}
}

func (m MiscModel) loadLongRunning() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.db.Query(`
			SELECT 
				name,
				pid,
				MIN(captured_at) as first_seen
			FROM processes
			WHERE captured_at > datetime('now', '-48 hours')
			GROUP BY name, pid
			HAVING COUNT(*) > 10
			ORDER BY first_seen ASC
			LIMIT 20
		`)
		if err != nil {
			return errMsg{err}
		}
		defer rows.Close()

		var processes []longRunningProcess
		for rows.Next() {
			var p longRunningProcess
			var firstSeen string
			if err := rows.Scan(&p.name, &p.pid, &firstSeen); err != nil {
				return errMsg{err}
			}
			p.duration = fmt.Sprintf("since %s", firstSeen)
			processes = append(processes, p)
		}
		return longRunningLoadedMsg{processes}
	}
}
