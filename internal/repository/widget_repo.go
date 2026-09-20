package repository

import (
	"context"
	"encoding/json"
	"time"

	"webshop/internal/database"
)

type Widget struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"` // hero, featured_products, categories_grid, banner_promo, trust_badges, newsletter
	Title     *string         `json:"title,omitempty"`
	Subtitle  *string         `json:"subtitle,omitempty"`
	Config    json.RawMessage `json:"config"`
	SortOrder int             `json:"sort_order"`
	IsActive  bool            `json:"is_active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type WidgetRepo struct {
	db *database.DB
}

func NewWidgetRepo(db *database.DB) *WidgetRepo {
	return &WidgetRepo{db: db}
}

func (r *WidgetRepo) ListActive(ctx context.Context) ([]Widget, error) {
	query := `
		SELECT id, type, title, subtitle, config, sort_order, is_active, created_at, updated_at
		FROM widgets
		WHERE is_active = true
		ORDER BY sort_order ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var widgets []Widget
	for rows.Next() {
		var w Widget
		if err := rows.Scan(&w.ID, &w.Type, &w.Title, &w.Subtitle, &w.Config, &w.SortOrder, &w.IsActive, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		widgets = append(widgets, w)
	}
	return widgets, nil
}

func (r *WidgetRepo) ListAll(ctx context.Context) ([]Widget, error) {
	query := `
		SELECT id, type, title, subtitle, config, sort_order, is_active, created_at, updated_at
		FROM widgets
		ORDER BY sort_order ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var widgets []Widget
	for rows.Next() {
		var w Widget
		if err := rows.Scan(&w.ID, &w.Type, &w.Title, &w.Subtitle, &w.Config, &w.SortOrder, &w.IsActive, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		widgets = append(widgets, w)
	}
	return widgets, nil
}

func (r *WidgetRepo) Upsert(ctx context.Context, id *string, widgetType string, title, subtitle *string, config json.RawMessage, sortOrder int, isActive bool) error {
	if id == nil || *id == "" {
		query := `
			INSERT INTO widgets (type, title, subtitle, config, sort_order, is_active)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := r.db.Pool.Exec(ctx, query, widgetType, title, subtitle, config, sortOrder, isActive)
		return err
	}

	query := `
		UPDATE widgets
		SET type = $2, title = $3, subtitle = $4, config = $5, sort_order = $6, is_active = $7, updated_at = now()
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, *id, widgetType, title, subtitle, config, sortOrder, isActive)
	return err
}

func (r *WidgetRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM widgets WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}