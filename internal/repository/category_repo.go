package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"webshop/internal/database"
)

type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	ImageURL    *string   `json:"image_url,omitempty"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type CategoryRepo struct {
	db *database.DB
}

func NewCategoryRepo(db *database.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) ListActive(ctx context.Context) ([]Category, error) {
	query := `
		SELECT id, name, slug, description, image_url, sort_order, is_active, created_at
		FROM categories
		WHERE is_active = true
		ORDER BY sort_order ASC, name ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.ImageURL, &c.SortOrder, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *CategoryRepo) ListAll(ctx context.Context) ([]Category, error) {
	query := `
		SELECT id, name, slug, description, image_url, sort_order, is_active, created_at
		FROM categories
		ORDER BY sort_order ASC, name ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.ImageURL, &c.SortOrder, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *CategoryRepo) GetBySlug(ctx context.Context, slug string) (*Category, error) {
	query := `
		SELECT id, name, slug, description, image_url, sort_order, is_active, created_at
		FROM categories
		WHERE slug = $1
	`
	c := &Category{}
	err := r.db.Pool.QueryRow(ctx, query, slug).
		Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.ImageURL, &c.SortOrder, &c.IsActive, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}

func (r *CategoryRepo) Create(ctx context.Context, name, slug string, description, imageURL *string, sortOrder int) (*Category, error) {
	query := `
		INSERT INTO categories (name, slug, description, image_url, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, slug, description, image_url, sort_order, is_active, created_at
	`
	c := &Category{}
	err := r.db.Pool.QueryRow(ctx, query, name, slug, description, imageURL, sortOrder).
		Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.ImageURL, &c.SortOrder, &c.IsActive, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CategoryRepo) Update(ctx context.Context, id, name, slug string, description, imageURL *string, sortOrder int, isActive bool) error {
	query := `
		UPDATE categories
		SET name = $2, slug = $3, description = $4, image_url = $5, sort_order = $6, is_active = $7
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id, name, slug, description, imageURL, sortOrder, isActive)
	return err
}

func (r *CategoryRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM categories WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}