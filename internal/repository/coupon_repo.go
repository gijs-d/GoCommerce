package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"webshop/internal/database"
)

type Coupon struct {
	ID            string     `json:"id"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"` // 'percent' of 'fixed'
	DiscountValue float64    `json:"discount_value"`
	MinSpend      *float64   `json:"min_spend,omitempty"`
	MaxUses       *int       `json:"max_uses,omitempty"`
	UsesCount     int        `json:"uses_count"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CouponRepo struct {
	db *database.DB
}

func NewCouponRepo(db *database.DB) *CouponRepo {
	return &CouponRepo{db: db}
}

// GetValidCoupon controleert of de kortingscode bestaat, actief is, niet verlopen is en het minimumbedrag haalt
func (r *CouponRepo) GetValidCoupon(ctx context.Context, code string, subtotal float64) (*Coupon, float64, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	query := `
		SELECT id, code, discount_type, discount_value, min_spend, max_uses, uses_count, expires_at, is_active, created_at
		FROM coupons
		WHERE UPPER(code) = $1
	`
	c := &Coupon{}
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(
		&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinSpend,
		&c.MaxUses, &c.UsesCount, &c.ExpiresAt, &c.IsActive, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, errors.New("ongeldige kortingscode")
		}
		return nil, 0, err
	}

	if !c.IsActive {
		return nil, 0, errors.New("deze kortingscode is niet meer actief")
	}

	if c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now()) {
		return nil, 0, errors.New("deze kortingscode is verlopen")
	}

	if c.MaxUses != nil && c.UsesCount >= *c.MaxUses {
		return nil, 0, errors.New("het maximum aantal keren voor deze code is bereikt")
	}

	if c.MinSpend != nil && subtotal < *c.MinSpend {
		return nil, 0, fmt.Errorf("deze code is pas geldig vanaf een bestelbedrag van € %.2f", *c.MinSpend)
	}

	// Bereken de korting
	var discountAmount float64
	if c.DiscountType == "percent" {
		discountAmount = (subtotal * c.DiscountValue) / 100.0
	} else {
		discountAmount = c.DiscountValue
	}

	if discountAmount > subtotal {
		discountAmount = subtotal
	}

	return c, discountAmount, nil
}

func (r *CouponRepo) IncrementUsage(ctx context.Context, id string) error {
	query := `UPDATE coupons SET uses_count = uses_count + 1 WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *CouponRepo) ListAll(ctx context.Context) ([]Coupon, error) {
	query := `
		SELECT id, code, discount_type, discount_value, min_spend, max_uses, uses_count, expires_at, is_active, created_at
		FROM coupons
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coupons []Coupon
	for rows.Next() {
		var c Coupon
		if err := rows.Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinSpend, &c.MaxUses, &c.UsesCount, &c.ExpiresAt, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		coupons = append(coupons, c)
	}
	return coupons, nil
}

func (r *CouponRepo) Create(ctx context.Context, code, discountType string, discountValue float64, minSpend *float64, maxUses *int, expiresAt *time.Time, isActive bool) (*Coupon, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	query := `
		INSERT INTO coupons (code, discount_type, discount_value, min_spend, max_uses, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, code, discount_type, discount_value, min_spend, max_uses, uses_count, expires_at, is_active, created_at
	`
	c := &Coupon{}
	err := r.db.Pool.QueryRow(ctx, query, code, discountType, discountValue, minSpend, maxUses, expiresAt, isActive).Scan(
		&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinSpend, &c.MaxUses, &c.UsesCount, &c.ExpiresAt, &c.IsActive, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CouponRepo) Update(ctx context.Context, id, code, discountType string, discountValue float64, minSpend *float64, maxUses *int, expiresAt *time.Time, isActive bool) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	query := `
		UPDATE coupons
		SET code = $2, discount_type = $3, discount_value = $4, min_spend = $5, max_uses = $6, expires_at = $7, is_active = $8
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id, code, discountType, discountValue, minSpend, maxUses, expiresAt, isActive)
	return err
}

func (r *CouponRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM coupons WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}