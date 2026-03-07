package agent

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Cleaner struct {
	db       *sql.DB
	interval time.Duration
	maxAge   time.Duration
}

func NewCleaner(db *sql.DB, interval time.Duration, maxAge time.Duration) *Cleaner {
	return &Cleaner{
		db:       db,
		interval: interval,
		maxAge:   maxAge,
	}
}

func (c *Cleaner) Start() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.cleanup(); err != nil {
			log.Printf("cleanup error: %v", err)
		}
	}
}

func (c *Cleaner) cleanup() error {
	cutoff := time.Now().Add(-c.maxAge)

	result, err := c.db.Exec(`
		DELETE FROM processes 
		WHERE captured_at < ?`, cutoff,
	)
	if err != nil {
		return fmt.Errorf("could not cleanup processes: %w", err)
	}

	rows, _ := result.RowsAffected()

	result, err = c.db.Exec(`
		DELETE FROM process_events 
		WHERE occurred_at < ?`, cutoff,
	)
	if err != nil {
		return fmt.Errorf("could not cleanup process events: %w", err)
	}

	evtRows, _ := result.RowsAffected()

	log.Printf("cleanup: removed %d process snapshots, %d events older than %s",
		rows, evtRows, c.maxAge,
	)

	return nil
}