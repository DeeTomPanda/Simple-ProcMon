package tui

import (
	"database/sql"
	"fmt"
	"time"

	"simple-procmon/internal/models"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sortBy int

const (
	sortByCPU sortBy = iota
	sortByMem
)

type TopModel struct {
	db        *sql.DB
	table     table.Model
	sortBy    sortBy
	processes []models.Process
	err       error
}

type processesLoadedMsg struct {
	processes []models.Process
}

type errMsg struct {
	err error
}

func NewTopModel(db *sql.DB) TopModel {
	cols := []table.Column{
		{Title: "PID", Width: 8},
		{Title: "NAME", Width: 24},
		{Title: "CPU%", Width: 10},
		{Title: "MEM(MB)", Width: 10},
		{Title: "STATUS", Width: 10},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	t.SetStyles(tableStyles())

	return TopModel{
		db:     db,
		table:  t,
		sortBy: sortByCPU,
	}
}

func (m TopModel) Init() tea.Cmd {
	return tea.Batch(m.loadProcesses(), tickCmd())
}

func (m TopModel) Update(msg tea.Msg) (TopModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "c":
			m.sortBy = sortByCPU
			return m, m.loadProcesses()
		case "r":
			m.sortBy = sortByMem
			return m, m.loadProcesses()
		case "up", "down":
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	case processesLoadedMsg:
		m.processes = msg.processes
		m.table.SetRows(processesToRows(msg.processes))
	case errMsg:
		m.err = msg.err
	case tickMsg:
		return m, tea.Batch(m.loadProcesses(), tickCmd())
	}

	return m, nil
}

func (m TopModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v", m.err)
	}

	sort := "CPU"
	if m.sortBy == sortByMem {
		sort = "MEM"
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Render(fmt.Sprintf("  Top Processes — sorted by %s  [c]pu  [r]am", sort))

	return fmt.Sprintf("%s\n%s", header, m.table.View())
}

func (m TopModel) loadProcesses() tea.Cmd {
	return func() tea.Msg {
		orderBy := "cpu_percent"
		if m.sortBy == sortByMem {
			orderBy = "mem_rss"
		}

		query := fmt.Sprintf(`
			SELECT pid, name, cpu_percent, mem_rss, status
			FROM processes
			WHERE captured_at = (SELECT MAX(captured_at) FROM processes)
			ORDER BY %s DESC
			LIMIT 20`, orderBy)

		rows, err := m.db.Query(query)
		if err != nil {
			return errMsg{err}
		}
		defer rows.Close()

		var processes []models.Process
		for rows.Next() {
			var p models.Process
			if err := rows.Scan(&p.PID, &p.Name, &p.CPUPercent, &p.MemRSS, &p.Status); err != nil {
				return errMsg{err}
			}
			processes = append(processes, p)
		}

		return processesLoadedMsg{processes}
	}
}

func processesToRows(processes []models.Process) []table.Row {
	rows := make([]table.Row, len(processes))
	for i, p := range processes {
		rows[i] = table.Row{
			fmt.Sprintf("%d", p.PID),
			p.Name,
			fmt.Sprintf("%.2f", p.CPUPercent),
			fmt.Sprintf("%.1f", float64(p.MemRSS)/1024/1024),
			p.Status,
		}
	}
	return rows
}

func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	return s
}

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}
