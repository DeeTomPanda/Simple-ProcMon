package proc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	clockTicksPerSecond = 100
	maxReasonableRSSGB  = 64
)

type Process struct {
	PID       int
	Name      string
	CPUUsage  float64
	MemoryRSS uint64 // in bytes
	Status    string
}

type ProcCollector struct {
	mu               sync.Mutex
	prevProcessTimes map[int]float64
	lastCollect      time.Time
}

func NewProcCollector() *ProcCollector {
	return &ProcCollector{
		prevProcessTimes: make(map[int]float64),
	}
}

func (c *ProcCollector) Collect() ([]*Process, error) {
	pids, err := getProcessIDs()
	if err != nil {
		return nil, fmt.Errorf("could not list pids: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	deltaTime := 0.0
	if !c.lastCollect.IsZero() {
		deltaTime = now.Sub(c.lastCollect).Seconds()
	}
	c.lastCollect = now

	var processes []*Process
	for _, pid := range pids {
		p, err := c.readProcess(pid, deltaTime)
		if err != nil {
			continue
		}
		processes = append(processes, p)
	}

	return processes, nil
}

func (c *ProcCollector) readProcess(pid int, deltaTime float64) (*Process, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return nil, err
	}

	raw := string(data)
	nameStart := strings.Index(raw, "(")
	nameEnd := strings.LastIndex(raw, ")")
	if nameStart == -1 || nameEnd == -1 {
		return nil, fmt.Errorf("unexpected stat format for pid %d", pid)
	}

	name := raw[nameStart+1 : nameEnd]
	rest := strings.Fields(raw[nameEnd+2:])
	if len(rest) < 22 {
		return nil, fmt.Errorf("unexpected stat format for pid %d", pid)
	}

	status := statusName(rest[0])

	utime, _ := strconv.ParseFloat(rest[11], 64)
	stime, _ := strconv.ParseFloat(rest[14], 64)
	totalTime := utime + stime

	// memory from /proc/pid/status for accuracy
	memoryRSS := readMemory(pid)

	// CPU delta
	prevTotal := c.prevProcessTimes[pid]
	deltaProcess := (totalTime - prevTotal) / clockTicksPerSecond
	c.prevProcessTimes[pid] = totalTime

	cpuUsage := 0.0
	if deltaTime > 0 {
		cpuUsage = (deltaProcess / deltaTime) * 100
	}

	// skip kernel threads with no memory and no cpu
	if memoryRSS == 0 && cpuUsage == 0 {
		return nil, fmt.Errorf("skip kernel thread %d", pid)
	}

	return &Process{
		PID:       pid,
		Name:      name,
		CPUUsage:  cpuUsage,
		MemoryRSS: memoryRSS,
		Status:    status,
	}, nil
}

func readMemory(pid int) uint64 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return 0
			}
			kb, _ := strconv.ParseUint(fields[1], 10, 64)
			return kb * 1024 // KB to bytes
		}
	}
	return 0
}

func statusName(s string) string {
	switch s {
	case "R":
		return "running"
	case "S":
		return "sleep"
	case "I":
		return "idle"
	case "Z":
		return "zombie"
	case "D":
		return "disk-sleep"
	case "T":
		return "stopped"
	default:
		return s
	}
}

func getProcessIDs() ([]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	var pids []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}
