package registry

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Tool struct {
	ID          int
	Name        string
	Source      string
	Version     string
	LastCommit  string
	BinaryHash  string
	InstallPath string
	Type        string
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
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tools (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			source TEXT NOT NULL,
			version TEXT NOT NULL,
			last_commit TEXT,
			binary_hash TEXT,
			install_path TEXT,
			type TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS tool_embeddings (
			tool_id INTEGER PRIMARY KEY,
			description TEXT,
			vector BLOB,
			FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE CASCADE
		);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) RegisterTool(tool Tool) error {
	query := `INSERT OR REPLACE INTO tools (name, source, version, last_commit, binary_hash, install_path, type) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, tool.Name, tool.Source, tool.Version, tool.LastCommit, tool.BinaryHash, tool.InstallPath, tool.Type)
	return err
}

func (r *Registry) GetTool(name string) (*Tool, error) {
	query := `SELECT id, name, source, version, last_commit, binary_hash, install_path, type FROM tools WHERE name = ?`
	row := r.db.QueryRow(query, name)
	var t Tool
	err := row.Scan(&t.ID, &t.Name, &t.Source, &t.Version, &t.LastCommit, &t.BinaryHash, &t.InstallPath, &t.Type)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Registry) ListTools() ([]Tool, error) {
	rows, err := r.db.Query("SELECT id, name, source, version, last_commit, binary_hash, install_path, type FROM tools")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		var t Tool
		if err := rows.Scan(&t.ID, &t.Name, &t.Source, &t.Version, &t.LastCommit, &t.BinaryHash, &t.InstallPath, &t.Type); err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, nil
}

func (r *Registry) Close() error {
	return r.db.Close()
}
