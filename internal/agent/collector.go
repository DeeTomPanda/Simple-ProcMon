package agent

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type Collector struct {
	db       *sql.DB
	interval time.Duration
}

func NewCollector(db *sql.DB, interval time.Duration) *Collector {
	return &Collector{
		db:       db,
		interval: interval,
	}
}

func (c *Collector) Start() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.collect(); err != nil {
			log.Printf("collection error: %v", err)
		}
	}
}

func (c *Collector) collect() error {
	processes, err := process.Processes()
	if err != nil {
		return fmt.Errorf("could not list processes: %w", err)
	}

	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	now := time.Now()

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		cpu, err := p.CPUPercent()
		if err != nil {
			continue
		}

		mem, err := p.MemoryInfo()
		if err != nil {
			continue
		}

		status, err := p.Status()
		if err != nil {
			continue
		}

		_, err = tx.Exec(`
			INSERT INTO processes (pid, name, cpu_percent, mem_rss, status, captured_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			p.Pid, name, cpu, mem.RSS, status[0], now,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("could not insert process: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit: %w", err)
	}

	return nil
}