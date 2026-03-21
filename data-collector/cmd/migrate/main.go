package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var (
	infoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)
)

func logInfo(msg string, v ...any) {
	infoLogger.Println(fmt.Sprintf(msg, v...))
}

func logError(msg string, v ...any) {
	errorLogger.Println(fmt.Sprintf(msg, v...))
}

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func isMigrationApplied(db *sql.DB, version string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func markMigrationApplied(db *sql.DB, version string) error {
	_, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
	return err
}

func main() {
	logInfo("Starting migrations")

	if os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("could not load env file: %v", err)
		}
	}

	connectionString := os.Getenv("DB_CONNECTION_STRING")
	if connectionString == "" {
		logError("DB_CONNECTION_STRING is not set")
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		logError("failed to open db: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logError("failed to ping db: %v", err)
		os.Exit(1)
	}

	err = ensureMigrationTable(db)
	if err != nil {
		logError("failed to create migration table: %v", err)
		os.Exit(1)
	}

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		logError("failed to read migrations directory: %v", err)
		os.Exit(1)
	}

	logInfo("%v migration files found", len(files))

	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		applied, err := isMigrationApplied(db, version)
		if err != nil {
			logError("failed to check migration status: %v", err)
			os.Exit(1)
		}

		if applied {
			logInfo("skipping already applied migration: %s", version)
			continue
		}

		logInfo("running migration: %s", version)

		sqlContent, err := os.ReadFile(file)
		if err != nil {
			logError("failed to read file %s: %v", file, err)
			os.Exit(1)
		}

		_, err = db.Exec(string(sqlContent))
		if (err != nil) {
			logError("failed to execute migration %s: %v", version, err)
			os.Exit(1)
		}

		err = markMigrationApplied(db, version)
		if err != nil {
			logError("failed to mark migration as applied: %v", err)
			os.Exit(1)
		}

		logInfo("migration completed: %s", version)
	}

	logInfo("all migrations completed successfully")
}
