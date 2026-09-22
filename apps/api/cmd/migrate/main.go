package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "../../../database/migrations"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(128) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		log.Fatal(err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(files)

	for _, path := range files {
		name := filepath.Base(path)
		version := strings.TrimSuffix(name, ".up.sql")

		var exists bool
		if err := db.QueryRowContext(ctx,
			"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)",
			version,
		).Scan(&exists); err != nil {
			log.Fatal(err)
		}
		if exists {
			continue
		}

		body, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			log.Fatalf("migration %s failed: %v", name, err)
		}
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations(version) VALUES($1)",
			version,
		); err != nil {
			_ = tx.Rollback()
			log.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("applied %s\n", name)
	}
}
