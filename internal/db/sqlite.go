package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// InitDB opens (or creates) the SQLite database at the given path with WAL mode enabled.
func InitDB(path string) (*sql.DB, error) {
	// Ensure the parent directory exists.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", dir, err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys.
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if err := CreateTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// CreateTables creates all required tables if they do not already exist.
func CreateTables(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			agent_token TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			public_bio TEXT NOT NULL DEFAULT '',
			ip_address TEXT NOT NULL,
			mac_address TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			report_count INTEGER NOT NULL DEFAULT 0,
			last_heartbeat TEXT,
			created_at TEXT NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			task_id TEXT NOT NULL,
			mode TEXT NOT NULL CHECK(mode IN ('beacon', 'radar')),
			type TEXT NOT NULL,
			title TEXT NOT NULL,
			keywords TEXT NOT NULL DEFAULT '[]',
			status TEXT NOT NULL DEFAULT 'active',
			created_at TEXT NOT NULL,
			FOREIGN KEY (agent_id) REFERENCES agents(id)
		)`,

		`CREATE TABLE IF NOT EXISTS task_embeddings (
			task_id TEXT PRIMARY KEY,
			embedding BLOB NOT NULL,
			FOREIGN KEY (task_id) REFERENCES tasks(id)
		)`,

		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			initiator_agent TEXT NOT NULL,
			target_agent TEXT NOT NULL,
			initiator_task TEXT NOT NULL,
			target_task TEXT NOT NULL,
			state TEXT NOT NULL DEFAULT 'pending_acceptance',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (initiator_agent) REFERENCES agents(id),
			FOREIGN KEY (target_agent) REFERENCES agents(id),
			FOREIGN KEY (initiator_task) REFERENCES tasks(id),
			FOREIGN KEY (target_task) REFERENCES tasks(id)
		)`,

		`CREATE TABLE IF NOT EXISTS message_queue (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			from_agent_id TEXT NOT NULL,
			to_agent_id TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id),
			FOREIGN KEY (from_agent_id) REFERENCES agents(id),
			FOREIGN KEY (to_agent_id) REFERENCES agents(id)
		)`,

		`CREATE TABLE IF NOT EXISTS reports (
			id TEXT PRIMARY KEY,
			reporter_id TEXT NOT NULL,
			target_id TEXT NOT NULL,
			reason TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (reporter_id) REFERENCES agents(id),
			FOREIGN KEY (target_id) REFERENCES agents(id)
		)`,

		`CREATE TABLE IF NOT EXISTS registration_limits (
			ip_mac_hash TEXT PRIMARY KEY,
			daily_count INTEGER NOT NULL DEFAULT 0,
			last_reset_date TEXT NOT NULL
		)`,

		// Indexes for common queries.
		`CREATE INDEX IF NOT EXISTS idx_tasks_agent_id ON tasks(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_token ON agents(agent_token)`,
		`CREATE INDEX IF NOT EXISTS idx_message_queue_to_agent ON message_queue(to_agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_initiator ON conversations(initiator_agent)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_target ON conversations(target_agent)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_agent_task ON tasks(agent_id, task_id)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %s: %w", stmt[:60], err)
		}
	}

	return nil
}

// RunMigrations applies schema changes that cannot be expressed with CREATE TABLE IF NOT EXISTS.
// Each migration is idempotent (safe to run multiple times).
func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`ALTER TABLE tasks ADD COLUMN updated_at TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE conversations ADD COLUMN last_message_at TEXT`,
	}

	for _, m := range migrations {
		_, err := db.Exec(m)
		if err != nil {
			// Ignore "duplicate column" errors — means migration already applied.
			if isAlreadyExistsError(err) {
				continue
			}
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// isAlreadyExistsError checks if an error is a "duplicate column" SQLite error.
func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "duplicate column")
}
