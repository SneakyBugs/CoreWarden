package database

import (
	"embed"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*
var f embed.FS

// Usage:
// migrate.Exec(db, "postgres", database.GetMigrations(), migrate.Up)
func GetMigrations() *migrate.AssetMigrationSource {
	// For more info see https://github.com/rubenv/sql-migrate#as-a-library
	return &migrate.AssetMigrationSource{
		Asset: func(path string) ([]byte, error) {
			data, err := f.ReadFile(path)
			if err != nil {
				return nil, err
			}
			return data, nil
		},
		AssetDir: func(path string) ([]string, error) {
			entries, err := f.ReadDir(path)
			if err != nil {
				return nil, err
			}
			paths := make([]string, 0)
			for _, entry := range entries {
				paths = append(paths, entry.Name())
			}
			return paths, nil
		},
		Dir: "migrations",
	}
}
