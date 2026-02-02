package registry

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Tool struct {
	ID     int
	Name   string
	Source string
	Version string
	Type    string
}

type Registry struct {
	db *sql.DB
}

func Open(islandDir string) (*Registry, error) {
	dbPath := filepath.Join(islandDir, "island.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return &Registry{db: db}, nil
}

func initSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS tools (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		source TEXT NOT NULL,
		version TEXT NOT NULL,
		type TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

func (r *Registry) RegisterTool(tool Tool) error {
	query := `INSERT OR REPLACE INTO tools (name, source, version, type) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, tool.Name, tool.Source, tool.Version, tool.Type)
	return err
}

func (r *Registry) ListTools() ([]Tool, error) {
	rows, err := r.db.Query("SELECT id, name, source, version, type FROM tools")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		var t Tool
		if err := rows.Scan(&t.ID, &t.Name, &t.Source, &t.Version, &t.Type); err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, nil
}

func (r *Registry) Close() error {
	return r.db.Close()
}
