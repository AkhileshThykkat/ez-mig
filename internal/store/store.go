package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AkhileshThykkat/ez-mig/internal/session"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func StoreSetup() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to detect user home directory: %w", err)
	}

	dir := filepath.Join(home, ".ez-mig")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	dbPath := filepath.Join(dir, "ez-mig.db")

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open configuration database: %w", err)
	}

	db = conn

	return createTable()
}

func createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		name                 TEXT PRIMARY KEY,
		db_uri               TEXT NOT NULL,
		migration_files_path TEXT NOT NULL,
		created_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at           DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to bootstrap database schema: %w", err)
	}
	return nil
}

func Create(s session.Session) error {
	query := `
	INSERT INTO sessions (name, db_uri, migration_files_path)
	VALUES (?, ?, ?);`

	_, err := db.Exec(query, s.Name, s.DbURI, s.MigrationFilesPath)
	if err != nil {

		return fmt.Errorf("session %q already exists or database error: %w", s.Name, err)
	}
	return nil
}

func Update(s session.Session) error {
	query := `
	UPDATE sessions 
	SET db_uri = ?, migration_files_path = ?, updated_at = CURRENT_TIMESTAMP 
	WHERE name = ?;`

	res, err := db.Exec(query, s.DbURI, s.MigrationFilesPath, s.Name)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session %q not found", s.Name)
	}
	return nil
}

func Delete(name string) error {
	query := `DELETE FROM sessions WHERE name = ?;`

	res, err := db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session %q not found", name)
	}
	return nil
}

func GetByName(name string) (*session.Session, error) {
	query := `
	SELECT name, db_uri, migration_files_path, created_at, updated_at 
	FROM sessions 
	WHERE name = ?;`

	row := db.QueryRow(query, name)

	var s session.Session
	var createdAtStr, updatedAtStr string

	err := row.Scan(&s.Name, &s.DbURI, &s.MigrationFilesPath, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session %q not found", name)
		}
		return nil, fmt.Errorf("failed to fetch session: %w", err)
	}

	// Parse timestamps explicitly from text storage
	s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return &s, nil
}

func List() ([]session.Session, error) {
	query := `
	SELECT name, db_uri, migration_files_path, created_at, updated_at 
	FROM sessions 
	ORDER BY created_at ASC;`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []session.Session
	for rows.Next() {
		var s session.Session
		var createdAtStr, updatedAtStr string

		err := rows.Scan(&s.Name, &s.DbURI, &s.MigrationFilesPath, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}

		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}
