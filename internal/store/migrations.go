package store

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate runs all migration SQL files in order
func Migrate(ctx context.Context, db *sqlx.DB) error {
	migrationDir := "migrations"

	// Read migration directory
	entries, err := fs.ReadDir(migrations, migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort SQL files
	var sqlFiles []fs.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			sqlFiles = append(sqlFiles, entry)
		}
	}

	// Execute each migration
	for _, file := range sqlFiles {
		migrationName := file.Name()
		migrationPath := filepath.Join(migrationDir, migrationName)

		// Read migration file
		content, err := fs.ReadFile(migrations, migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", migrationName, err)
		}

		// Execute migration
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationName, err)
		}

		fmt.Printf("✓ Applied migration: %s\n", migrationName)
	}

	return nil
}
