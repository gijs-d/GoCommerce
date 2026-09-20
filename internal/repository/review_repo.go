package repository

import (
	"context"
	"time"

	"webshop/internal/database"
)

type ProductReview struct {
	ID         string    `json:"id"`
	ProductID  string    `json:"product_id"`
	UserID     *string   `json:"user_id,omitempty"`
	AuthorName string    `json:"author_name"`
	Rating     int       `json:"rating"`
	Comment    string    `json:"comment"`
	IsApproved bool      `json:"is_approved"`
	CreatedAt  time.Time `json:"created_at"`
}

type ReviewStats struct {
	AverageRating float64 `json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`
}

type ReviewRepo struct {
	db *database.DB
}

func NewReviewRepo(db *database.DB) *ReviewRepo {
	return &ReviewRepo{db: db}
}

func (r *ReviewRepo) ListApproved(ctx context.Context, productID string) ([]ProductReview, error) {
	query := `
		SELECT id, product_id, user_id, author_name, rating, comment, is_approved, created_at
		FROM product_reviews
		WHERE product_id = $1 AND is_approved = true
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []ProductReview
	for rows.Next() {
		var rev ProductReview
		if err := rows.Scan(&rev.ID, &rev.ProductID, &rev.UserID, &rev.AuthorName, &rev.Rating, &rev.Comment, &rev.IsApproved, &rev.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, nil
}

func (r *ReviewRepo) GetStats(ctx context.Context, productID string) (*ReviewStats, error) {
	query := `
		SELECT COALESCE(AVG(rating), 0), COUNT(*)
		FROM product_reviews
		WHERE product_id = $1 AND is_approved = true
	`
	stats := &ReviewStats{}
	err := r.db.Pool.QueryRow(ctx, query, productID).Scan(&stats.AverageRating, &stats.TotalReviews)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *ReviewRepo) Create(ctx context.Context, productID string, userID *string, authorName string, rating int, comment string) (*ProductReview, error) {
	query := `
		INSERT INTO product_reviews (product_id, user_id, author_name, rating, comment, is_approved)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id, product_id, user_id, author_name, rating, comment, is_approved, created_at
	`
	rev := &ProductReview{}
	err := r.db.Pool.QueryRow(ctx, query, productID, userID, authorName, rating, comment).Scan(
		&rev.ID, &rev.ProductID, &rev.UserID, &rev.AuthorName, &rev.Rating, &rev.Comment, &rev.IsApproved, &rev.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return rev, nil
}

func (r *ReviewRepo) ListAll(ctx context.Context) ([]ProductReview, error) {
	query := `
		SELECT id, product_id, user_id, author_name, rating, comment, is_approved, created_at
		FROM product_reviews
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []ProductReview
	for rows.Next() {
		var rev ProductReview
		if err := rows.Scan(&rev.ID, &rev.ProductID, &rev.UserID, &rev.AuthorName, &rev.Rating, &rev.Comment, &rev.IsApproved, &rev.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, nil
}

func (r *ReviewRepo) SetApproval(ctx context.Context, id string, isApproved bool) error {
	query := `UPDATE product_reviews SET is_approved = $2 WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, isApproved)
	return err
}

func (r *ReviewRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM product_reviews WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}