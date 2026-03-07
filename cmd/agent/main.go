package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"simple-procmon/config"
	"simple-procmon/db"
	"simple-procmon/internal/agent"
	"simple-procmon/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// First load all config from env

	_, err := config.Load()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	database, err := db.Open()
	if err != nil {
		log.Fatalf("could not open db: %v", err)
	}
	defer database.Close()

	err = db.Migrate(database)
	if err != nil {
		log.Fatalf("failed runnning migrations %v", err)
	}

	// Start collector
	collector := agent.NewCollector(database, config.GlobalConfig.CollectInterval)
	go collector.Start()

	// Start cleaner
	cleaner := agent.NewCleaner(database, config.GlobalConfig.CleanupInterval, config.GlobalConfig.MaxAge)
	go cleaner.Start()

	log.Println("procmon agent started")

	model := tui.New(database)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // full screen TUI
		tea.WithMouseCellMotion(), // mouse support
	)

	if _, err := p.Run(); err != nil {
		log.Printf("error running TUI: %v", err)
		os.Exit(1)
	}

	// Block until signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("procmon agent stopped")
}
