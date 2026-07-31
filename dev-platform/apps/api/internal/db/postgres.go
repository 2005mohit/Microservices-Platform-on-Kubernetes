package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var pool *sql.DB

func Init(dsn string) error {
	var err error
	pool, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	pool.SetMaxOpenConns(25)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(5 * time.Minute)

	if err := pool.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	return runMigrations(pool)
}

func Close() {
	if pool != nil {
		pool.Close()
	}
}

func Pool() *sql.DB {
	return pool
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL UNIQUE,
			git_repo TEXT NOT NULL,
			git_branch VARCHAR(255) DEFAULT 'main',
			env_vars JSONB DEFAULT '{}',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS deployments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			status VARCHAR(50) DEFAULT 'queued',
			branch VARCHAR(255) DEFAULT 'main',
			commit_sha VARCHAR(255) DEFAULT '',
			logs TEXT DEFAULT '',
			domain VARCHAR(255) DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
	}
	return nil
}
