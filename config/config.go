package config

import (
	"fmt"
	"log"
	"os"
	"simple-procmon/internal/constants"
	"time"

	"github.com/joho/godotenv"
)

var GlobalConfig *Config

type Config struct {
	DBPath          string
	CollectInterval time.Duration
	CleanupInterval time.Duration
	MaxAge          time.Duration
	MigrationsDir   string
}

func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	collectInterval, err := parseDuration(constants.COLLECT_INTERVAL, 30*time.Second)
	if err != nil {
		return nil, err
	}

	cleanupInterval, err := parseDuration(constants.CLEANUP_INTERVAL, 1*time.Hour)
	if err != nil {
		return nil, err
	}

	maxAge, err := parseDuration(constants.MAX_AGE, 48*time.Hour)
	if err != nil {
		return nil, err
	}

	dbPath, err := getEnv(constants.DBPATH)
	if err != nil {
		return nil, err
	}

	migrationsDir, err := getEnv(constants.MIGRATIONS_DIR)
	if err != nil {
		return nil, err
	}

	GlobalConfig = &Config{
		DBPath:          dbPath,
		CollectInterval: collectInterval,
		CleanupInterval: cleanupInterval,
		MaxAge:          maxAge,
		MigrationsDir:   migrationsDir,
	}

	return GlobalConfig, nil
}

func getEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}
	return "", fmt.Errorf("error parsing path")
}

func parseDuration(key string, fallback time.Duration) (time.Duration, error) {
	val := os.Getenv(key)
	if val == "" {
		return fallback, nil
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}

	return d, nil
}
