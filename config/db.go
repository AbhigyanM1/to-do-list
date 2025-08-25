package config

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func ConnectDatabase() {
	var err error
	DB, err = sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatalf("❌ Failed to connect to SQLite database: %v", err)
	}

	// Base table (safe if it already exists)
	createTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		scheduled_time TEXT NOT NULL,
		done INTEGER DEFAULT 0
	);
	`
	if _, err = DB.Exec(createTable); err != nil {
		log.Fatalf("❌ Failed to create tasks table: %v", err)
	}

	// Ensure required columns exist (idempotent)
	ensureColumn("tasks", "arrival_time", "TEXT")
	ensureColumn("tasks", "start_time", "TEXT")
	ensureColumn("tasks", "end_time", "TEXT")
	ensureColumn("tasks", "duration_sec", "INTEGER DEFAULT 2")

	log.Println("✅ Connected to SQLite and ensured tasks table/columns.")
}

// ensureColumn adds a column only if it doesn't exist yet.
func ensureColumn(table, col, typ string) {
	exists, err := columnExists(table, col)
	if err != nil {
		log.Fatalf("❌ Failed to inspect table %s: %v", table, err)
	}
	if exists {
		return
	}

	stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table, col, typ)
	if _, err := DB.Exec(stmt); err != nil {
		// Be tolerant of different duplicate error strings across SQLite versions.
		msg := strings.ToLower(err.Error())
		if !strings.Contains(msg, "duplicate column name") {
			log.Fatalf("❌ Failed to add column %s to %s: %v", col, table, err)
		}
	} else {
		log.Printf("🧱 Added column %s to %s.\n", col, table)
	}
}

// columnExists checks PRAGMA table_info for a column name.
func columnExists(table, col string) (bool, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s);", table)
	rows, err := DB.Query(query)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var (
		cid       int
		name      string
		colType   string
		notNull   int
		dfltValue any
		pk        int
	)

	for rows.Next() {
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if strings.EqualFold(name, col) {
			return true, nil
		}
	}
	return false, rows.Err()
}