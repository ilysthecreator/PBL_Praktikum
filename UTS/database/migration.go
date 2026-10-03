package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrationsAndSeeds(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	migrationFiles := []string{
		"001_create_tables.sql",
		"002_seed_data.sql",
	}

	for _, file := range migrationFiles {
		filePath := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("gagal membaca file migrasi %s: %w", filePath, err)
		}

		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("gagal mengeksekusi migrasi %s: %w", file, err)
		}
	}

	return nil
}
