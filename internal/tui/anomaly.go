package tui

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AnomalyModel struct {
	db    *sql.DB
	table table.Model
	err   error
}

type anomaly struct {
	name   string
	avgCPU sql.NullFloat64
	maxCPU sql.NullFloat64
	avgMem sql.NullFloat64
	maxMem sql.NullFloat64
}

type anomaliesLoadedMsg struct {
	anomalies []anomaly
}

func NewAnomalyModel(db *sql.DB) AnomalyModel {
	cols := []table.Column{
		{Title: "NAME", Width: 24},
		{Title: "AVG CPU%", Width: 10},
		{Title: "MAX CPU%", Width: 10},
		{Title: "AVG MEM(MB)", Width: 12},
		{Title: "MAX MEM(MB)", Width: 12},
		{Title: "SPIKE", Width: 10},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(20),
	)
	t.SetStyles(tableStyles())

	return AnomalyModel{
		db:    db,
		table: t,
	}
}

func (m AnomalyModel) Init() tea.Cmd {
	return tea.Batch(m.loadAnomalies(), tickCmd())
}

func (m AnomalyModel) Update(msg tea.Msg) (AnomalyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case anomaliesLoadedMsg:
		m.table.SetRows(anomaliesToRows(msg.anomalies))
	case errMsg:
		m.err = msg.err
	case tickMsg:
		return m, tea.Batch(m.loadAnomalies(), tickCmd())
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "down":
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m AnomalyModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v", m.err)
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("208")).
		Render("  Anomalies — processes spiking above 2x their hourly average")

	return fmt.Sprintf("%s\n%s", header, m.table.View())
}

func (m AnomalyModel) loadAnomalies() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.db.Query(`
			SELECT 
				name,
				AVG(cpu_percent) as avg_cpu,
				MAX(cpu_percent) as max_cpu,
				AVG(mem_rss)     as avg_mem,
				MAX(mem_rss)     as max_mem
			FROM processes
			WHERE captured_at > datetime('now', '-1 hour')
			GROUP BY name
			HAVING max_cpu > avg_cpu * 2
			ORDER BY max_cpu DESC
		`)
		if err != nil {
			return errMsg{err}
		}
		defer rows.Close()

		var anomalies []anomaly
		for rows.Next() {
			var a anomaly
			if err := rows.Scan(
				&a.name, &a.avgCPU, &a.maxCPU, &a.avgMem, &a.maxMem,
			); err != nil {
				return errMsg{err}
			}
			anomalies = append(anomalies, a)
		}

		return anomaliesLoadedMsg{anomalies}
	}
}

func anomaliesToRows(anomalies []anomaly) []table.Row {
	rows := make([]table.Row, len(anomalies))
	for i, a := range anomalies {
		avgCPU := 0.0
		maxCPU := 0.0
		avgMem := 0.0
		maxMem := 0.0

		if a.avgCPU.Valid {
			avgCPU = a.avgCPU.Float64
		}
		if a.maxCPU.Valid {
			maxCPU = a.maxCPU.Float64
		}
		if a.avgMem.Valid {
			avgMem = a.avgMem.Float64
		}
		if a.maxMem.Valid {
			maxMem = a.maxMem.Float64
		}

		spike := fmt.Sprintf("%.1fx", maxCPU/max(avgCPU, 0.01))
		rows[i] = table.Row{
			a.name,
			fmt.Sprintf("%.2f", avgCPU),
			fmt.Sprintf("%.2f", maxCPU),
			fmt.Sprintf("%.1f", avgMem/1024/1024),
			fmt.Sprintf("%.1f", maxMem/1024/1024),
			spike,
		}
	}
	return rows
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
