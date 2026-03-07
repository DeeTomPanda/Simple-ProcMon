package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"simple-procmon/config"
	"simple-procmon/db"
	"simple-procmon/internal/agent"
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

	// Block until signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("procmon agent stopped")
}
