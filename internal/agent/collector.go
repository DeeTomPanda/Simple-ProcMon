package agent

import (
	"database/sql"
	"fmt"
	"log"
	"simple-procmon/internal/proc"
	"time"
)

type Collector struct {
	db            *sql.DB
	interval      time.Duration
	ProcCollector *proc.ProcCollector
}

func NewCollector(db *sql.DB, interval time.Duration) *Collector {
	return &Collector{
		db:            db,
		interval:      interval,
		ProcCollector: proc.NewProcCollector(),
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

	processes, err := c.ProcCollector.Collect()
	if err != nil {
		return fmt.Errorf("could not list processes: %w", err)
	}

	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	stmt, err := tx.Prepare(`
			INSERT INTO processes (pid, name, cpu_percent, mem_rss, status, captured_at)
			VALUES (?, ?, ?, ?, ?, ?)
			`)
	if err != nil {
		return fmt.Errorf("could not prepare statement: %w", err)
	}

	defer stmt.Close()

	now := time.Now()

	for _, process := range processes {
		_, err := stmt.Exec(process.PID, process.Name, process.CPUUsage, process.MemoryRSS, process.Status, now)
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
