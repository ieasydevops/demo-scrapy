package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	if dbPath == "" {
		dbPath = os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "./monitor.db"
		}
	}
	
	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		os.MkdirAll(dbDir, 0755)
	}
	
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	if err = createTables(); err != nil {
		return err
	}

	return nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS web_pages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS keywords (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			keyword TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS push_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			push_time TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS monitor_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			web_page_id INTEGER NOT NULL,
			crawl_time TEXT NOT NULL,
			crawl_freq TEXT NOT NULL,
			keywords TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (web_page_id) REFERENCES web_pages(id)
		)`,
		`CREATE TABLE IF NOT EXISTS subscribe_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			push_time TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS announcements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			url TEXT NOT NULL UNIQUE,
			publish_date TEXT NOT NULL,
			content TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			web_page_id INTEGER,
			publisher TEXT,
			FOREIGN KEY (web_page_id) REFERENCES web_pages(id)
		)`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	if err := migrateTables(); err != nil {
		return err
	}

	return nil
}

func migrateTables() error {
	var exists int
	err := DB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='announcements'").Scan(&exists)
	if err != nil {
		return err
	}

	if exists > 0 {
		rows, err := DB.Query("PRAGMA table_info(announcements)")
		if err != nil {
			return err
		}
		defer rows.Close()

		hasWebPageID := false
		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull, pk int
			var defaultValue interface{}
			if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
				continue
			}
			if name == "web_page_id" {
				hasWebPageID = true
				break
			}
		}

		if !hasWebPageID {
			_, err = DB.Exec("ALTER TABLE announcements ADD COLUMN web_page_id INTEGER")
			if err != nil {
				return fmt.Errorf("添加 web_page_id 列失败: %v", err)
			}
		}

		hasPublisher := false
		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull, pk int
			var defaultValue interface{}
			if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
				continue
			}
			if name == "publisher" {
				hasPublisher = true
				break
			}
		}
		rows.Close()

		if !hasPublisher {
			_, err = DB.Exec("ALTER TABLE announcements ADD COLUMN publisher TEXT")
			if err != nil {
				return fmt.Errorf("添加 publisher 列失败: %v", err)
			}
		}
	}

	return nil
}
