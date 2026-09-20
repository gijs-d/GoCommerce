package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("ongeldige database url: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("fout bij openen database pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("kan geen verbinding maken met postgres: %w", err)
	}

	log.Println("Database verbonden met PostgreSQL 17")
	return &DB{Pool: pool}, nil
}

func (db *DB) RunMigrations(ctx context.Context, migrationPath string) error {
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("migratiebestand %s niet gevonden: %w", migrationPath, err)
	}

	_, err = db.Pool.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("fout tijdens uitvoeren van migratie: %w", err)
	}

	log.Println("Database migraties succesvol uitgevoerd")
	return nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}