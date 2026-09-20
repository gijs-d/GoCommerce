package repository

import (
	"context"
	"time"

	"webshop/internal/database"
)

type WishlistItem struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	Product   Product   `json:"product"`
	CreatedAt time.Time `json:"created_at"`
}

type WishlistRepo struct {
	db *database.DB
}

func NewWishlistRepo(db *database.DB) *WishlistRepo {
	return &WishlistRepo{db: db}
}

func (r *WishlistRepo) Add(ctx context.Context, userID, productID string) error {
	query := `
		INSERT INTO wishlists (user_id, product_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, product_id) DO NOTHING
	`
	_, err := r.db.Pool.Exec(ctx, query, userID, productID)
	return err
}

func (r *WishlistRepo) Remove(ctx context.Context, userID, productID string) error {
	query := `DELETE FROM wishlists WHERE user_id = $1 AND product_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, userID, productID)
	return err
}

func (r *WishlistRepo) IsWishlisted(ctx context.Context, userID, productID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`
	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, userID, productID).Scan(&exists)
	return exists, err
}

func (r *WishlistRepo) ListByUser(ctx context.Context, userID string) ([]WishlistItem, error) {
	query := `
		SELECT 
			w.id, w.user_id, w.product_id, w.created_at,
			p.id, p.category_id, c.name, c.slug, p.name, p.slug, 
			p.short_description, p.description, p.base_price, p.featured, 
			p.is_active, p.seo_title, p.seo_description,
			(SELECT url FROM product_images pi WHERE pi.product_id = p.id ORDER BY pi.is_primary DESC, pi.sort_order ASC LIMIT 1) as primary_image,
			p.created_at, p.updated_at
		FROM wishlists w
		JOIN products p ON p.id = w.product_id
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WishlistItem
	for rows.Next() {
		var item WishlistItem
		var p Product
		err := rows.Scan(
			&item.ID, &item.UserID, &item.ProductID, &item.CreatedAt,
			&p.ID, &p.CategoryID, &p.CategoryName, &p.CategorySlug, &p.Name, &p.Slug,
			&p.ShortDescription, &p.Description, &p.BasePrice, &p.Featured,
			&p.IsActive, &p.SEOTitle, &p.SEODescription, &p.PrimaryImage,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		item.Product = p
		items = append(items, item)
	}

	return items, nil
}