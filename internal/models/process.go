package models

import "time"

type Process struct {
	PID        int32
	Name       string
	CPUPercent float64
	MemRSS     uint64
	Status     string
	CapturedAt time.Time
}

type ProcessEvent struct {
	ID         int64
	PID        int32
	Name       string
	EventType  string // "started", "stopped", "zombie"
	OccurredAt time.Time
}

type ProcessSnapshot struct {
	Processes []Process
	CapturedAt time.Time
}

type TimeRange string

const (
	RangeHour  TimeRange = "1h"
	RangeDay   TimeRange = "24h"
	RangeWeek  TimeRange = "7d"
)