package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"webshop/internal/database"
)

type SettingItem struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type SettingsRepo struct {
	db *database.DB
}

func NewSettingsRepo(db *database.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) Get(ctx context.Context, key string) (json.RawMessage, error) {
	query := `SELECT value FROM store_settings WHERE key = $1`
	var val json.RawMessage
	err := r.db.Pool.QueryRow(ctx, query, key).Scan(&val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

func (r *SettingsRepo) Set(ctx context.Context, key string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO store_settings (key, value, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = now()
	`
	_, err = r.db.Pool.Exec(ctx, query, key, bytes)
	return err
}

func (r *SettingsRepo) GetAll(ctx context.Context) (map[string]json.RawMessage, error) {
	query := `SELECT key, value FROM store_settings`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]json.RawMessage)
	for rows.Next() {
		var key string
		var val json.RawMessage
		if err := rows.Scan(&key, &val); err == nil {
			settings[key] = val
		}
	}
	return settings, nil
}