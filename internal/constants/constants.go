package constants

import "time"

const DBPATH = "DB_PATH"
const COLLECT_INTERVAL = "COLLECT_INTERVAL"
const CLEANUP_INTERVAL = "CLEANUP_INTERVAL"
const MAX_AGE = "MAX_AGE"
const MIGRATIONS_DIR = "GOOSE_MIGRATIONS"

const (
	// Default DB driver for goose + sqlite
	DBDriver = "sqlite"
	// Environment variable name used to override the DB file path
	DBEnvKey = DBPATH
	// Default directory to store the sqlite DB file
	DefaultDBDir = "data"
	// Default sqlite DB file name
	DefaultDBFile = "procmon.db"
	// Directory containing goose migrations (relative to project root)
	MigrationsDir = "migrations"
	// Default acquisition interval for process metrics
	DefaultCollectInterval = 30 * time.Second
	// Default cleanup interval for old process metrics
	DefaultCleanupInterval = 1 * time.Hour
	// Default maximum age for process metrics before cleanup
	DefaultMaxAge = 48 * time.Hour
)
