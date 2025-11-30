package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// InitDBLite открывает SQLite и применяет схему, возвращая подключение
func InitDBLite(dataSourceName string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	schema := `
	CREATE TABLE IF NOT EXISTS users (
	  id               INTEGER PRIMARY KEY AUTOINCREMENT,
	  tg_id            INTEGER NOT NULL UNIQUE,
	  username         TEXT    NOT NULL,
	  gemini_api_key   TEXT,
	  is_admin         INTEGER NOT NULL DEFAULT 0,
	  is_active        INTEGER NOT NULL DEFAULT 1,
	  last_login       DATETIME,
	  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	`
	if _, err := sqlDB.Exec(schema); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return sqlDB, nil
}
