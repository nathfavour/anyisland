package registry

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Tool struct {
	ID           int
	Name         string
	Source       string
	SourceDir    string
	Version      string
	LastCommit   string
	BinaryHash   string
	InstallPath  string
	Type         string
	Dependencies []string
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
	// 1. Create tables if they don't exist
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tools (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			source TEXT NOT NULL,
			source_dir TEXT,
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
		`CREATE TABLE IF NOT EXISTS tool_dependencies (
			tool_id INTEGER,
			dependency_name TEXT,
			PRIMARY KEY(tool_id, dependency_name),
			FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE CASCADE
		);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}

	// 2. Migration: Add missing columns to existing tools table
	columns := map[string]string{
		"source_dir":   "TEXT",
		"last_commit":  "TEXT",
		"binary_hash":  "TEXT",
		"install_path": "TEXT",
	}

	for col, colType := range columns {
		// Check if column exists
		query := fmt.Sprintf("PRAGMA table_info(tools)")
		rows, err := db.Query(query)
		if err != nil {
			return err
		}
		
		found := false
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dflt_value interface{}
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
				continue
			}
			if name == col {
				found = true
				break
			}
		}
		rows.Close()

		if !found {
			alterQuery := fmt.Sprintf("ALTER TABLE tools ADD COLUMN %s %s", col, colType)
			if _, err := db.Exec(alterQuery); err != nil {
				return fmt.Errorf("failed to add column %s: %w", col, err)
			}
		}
	}

	return nil
}

func (r *Registry) RegisterTool(tool Tool) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT OR REPLACE INTO tools (name, source, source_dir, version, last_commit, binary_hash, install_path, type) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := tx.Exec(query, tool.Name, tool.Source, tool.SourceDir, tool.Version, tool.LastCommit, tool.BinaryHash, tool.InstallPath, tool.Type)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		// If it was a REPLACE, we might need to find the ID
		row := tx.QueryRow("SELECT id FROM tools WHERE name = ?", tool.Name)
		if err := row.Scan(&id); err != nil {
			return err
		}
	}

	// Remove old dependencies
	_, err = tx.Exec("DELETE FROM tool_dependencies WHERE tool_id = ?", id)
	if err != nil {
		return err
	}

	// Add new dependencies
	for _, dep := range tool.Dependencies {
		_, err = tx.Exec("INSERT INTO tool_dependencies (tool_id, dependency_name) VALUES (?, ?)", id, dep)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Registry) GetTool(name string) (*Tool, error) {
	query := `SELECT id, name, source, COALESCE(source_dir, ''), version, COALESCE(last_commit, ''), COALESCE(binary_hash, ''), COALESCE(install_path, ''), type FROM tools WHERE name = ?`
	row := r.db.QueryRow(query, name)
	var t Tool
	err := row.Scan(&t.ID, &t.Name, &t.Source, &t.SourceDir, &t.Version, &t.LastCommit, &t.BinaryHash, &t.InstallPath, &t.Type)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.Dependencies, _ = r.loadDependencies(t.ID)
	return &t, nil
}

func (r *Registry) GetToolByPath(path string) (*Tool, error) {
	query := `SELECT id, name, source, COALESCE(source_dir, ''), version, COALESCE(last_commit, ''), COALESCE(binary_hash, ''), COALESCE(install_path, ''), type FROM tools WHERE install_path = ?`
	row := r.db.QueryRow(query, path)
	var t Tool
	err := row.Scan(&t.ID, &t.Name, &t.Source, &t.SourceDir, &t.Version, &t.LastCommit, &t.BinaryHash, &t.InstallPath, &t.Type)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.Dependencies, _ = r.loadDependencies(t.ID)
	return &t, nil
}

func (r *Registry) ListTools() ([]Tool, error) {
	query := `SELECT id, name, source, COALESCE(source_dir, ''), version, COALESCE(last_commit, ''), COALESCE(binary_hash, ''), COALESCE(install_path, ''), type FROM tools`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		var t Tool
		if err := rows.Scan(&t.ID, &t.Name, &t.Source, &t.SourceDir, &t.Version, &t.LastCommit, &t.BinaryHash, &t.InstallPath, &t.Type); err != nil {
			return nil, err
		}
		t.Dependencies, _ = r.loadDependencies(t.ID)
		tools = append(tools, t)
	}
	return tools, nil
}

func (r *Registry) loadDependencies(toolID int) ([]string, error) {
	rows, err := r.db.Query("SELECT dependency_name FROM tool_dependencies WHERE tool_id = ?", toolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []string
	for rows.Next() {
		var dep string
		if err := rows.Scan(&dep); err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	return deps, nil
}

func (r *Registry) RemoveTool(name string) error {
	query := `DELETE FROM tools WHERE name = ?`
	_, err := r.db.Exec(query, name)
	return err
}

func (r *Registry) Close() error {
	return r.db.Close()
}
