package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"webshop/internal/database"
)

type ShippingMethod struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	Cost          float64   `json:"cost"`
	FreeThreshold *float64  `json:"free_threshold,omitempty"`
	IsActive      bool      `json:"is_active"`
	SortOrder     int       `json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
}

type ShippingRepo struct {
	db *database.DB
}

func NewShippingRepo(db *database.DB) *ShippingRepo {
	return &ShippingRepo{db: db}
}

func (r *ShippingRepo) ListActive(ctx context.Context) ([]ShippingMethod, error) {
	query := `
		SELECT id, title, description, cost, free_threshold, is_active, sort_order, created_at
		FROM shipping_methods
		WHERE is_active = true
		ORDER BY sort_order ASC, cost ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []ShippingMethod
	for rows.Next() {
		var m ShippingMethod
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.Cost, &m.FreeThreshold, &m.IsActive, &m.SortOrder, &m.CreatedAt); err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}
	return methods, nil
}

func (r *ShippingRepo) ListAll(ctx context.Context) ([]ShippingMethod, error) {
	query := `
		SELECT id, title, description, cost, free_threshold, is_active, sort_order, created_at
		FROM shipping_methods
		ORDER BY sort_order ASC, cost ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []ShippingMethod
	for rows.Next() {
		var m ShippingMethod
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.Cost, &m.FreeThreshold, &m.IsActive, &m.SortOrder, &m.CreatedAt); err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}
	return methods, nil
}

func (r *ShippingRepo) GetByID(ctx context.Context, id string) (*ShippingMethod, error) {
	query := `
		SELECT id, title, description, cost, free_threshold, is_active, sort_order, created_at
		FROM shipping_methods
		WHERE id = $1
	`
	m := &ShippingMethod{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.Title, &m.Description, &m.Cost, &m.FreeThreshold, &m.IsActive, &m.SortOrder, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

func (r *ShippingRepo) Create(ctx context.Context, title string, desc *string, cost float64, freeThreshold *float64, sortOrder int, isActive bool) (*ShippingMethod, error) {
	query := `
		INSERT INTO shipping_methods (title, description, cost, free_threshold, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, cost, free_threshold, is_active, sort_order, created_at
	`
	m := &ShippingMethod{}
	err := r.db.Pool.QueryRow(ctx, query, title, desc, cost, freeThreshold, sortOrder, isActive).Scan(
		&m.ID, &m.Title, &m.Description, &m.Cost, &m.FreeThreshold, &m.IsActive, &m.SortOrder, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *ShippingRepo) Update(ctx context.Context, id, title string, desc *string, cost float64, freeThreshold *float64, sortOrder int, isActive bool) error {
	query := `
		UPDATE shipping_methods
		SET title = $2, description = $3, cost = $4, free_threshold = $5, sort_order = $6, is_active = $7
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id, title, desc, cost, freeThreshold, sortOrder, isActive)
	return err
}

func (r *ShippingRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM shipping_methods WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}