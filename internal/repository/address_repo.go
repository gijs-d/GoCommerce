package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"webshop/internal/database"
)

type Address struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Label       string    `json:"label"` // bijv. "Thuis", "Werk"
	FullName    string    `json:"full_name"`
	Street      string    `json:"street"`
	HouseNumber string    `json:"house_number"`
	Bus         *string   `json:"bus,omitempty"`
	City        string    `json:"city"`
	PostalCode  string    `json:"postal_code"`
	Country     string    `json:"country"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AddressRepo struct {
	db *database.DB
}

func NewAddressRepo(db *database.DB) *AddressRepo {
	return &AddressRepo{db: db}
}

func (r *AddressRepo) ListByUser(ctx context.Context, userID string) ([]Address, error) {
	query := `
		SELECT id, user_id, label, full_name, street, house_number, bus, city, postal_code, country, is_default, created_at, updated_at
		FROM addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []Address
	for rows.Next() {
		var a Address
		if err := rows.Scan(&a.ID, &a.UserID, &a.Label, &a.FullName, &a.Street, &a.HouseNumber, &a.Bus, &a.City, &a.PostalCode, &a.Country, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		addresses = append(addresses, a)
	}
	return addresses, nil
}

func (r *AddressRepo) Create(ctx context.Context, a *Address) (*Address, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if a.IsDefault {
		_, _ = tx.Exec(ctx, `UPDATE addresses SET is_default = false WHERE user_id = $1`, a.UserID)
	}

	query := `
		INSERT INTO addresses (user_id, label, full_name, street, house_number, bus, city, postal_code, country, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, query, a.UserID, a.Label, a.FullName, a.Street, a.HouseNumber, a.Bus, a.City, a.PostalCode, a.Country, a.IsDefault).
		Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AddressRepo) Update(ctx context.Context, a *Address) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if a.IsDefault {
		_, _ = tx.Exec(ctx, `UPDATE addresses SET is_default = false WHERE user_id = $1 AND id != $2`, a.UserID, a.ID)
	}

	query := `
		UPDATE addresses
		SET label = $3, full_name = $4, street = $5, house_number = $6, bus = $7, city = $8, postal_code = $9, country = $10, is_default = $11, updated_at = now()
		WHERE id = $1 AND user_id = $2
	`
	_, err = tx.Exec(ctx, query, a.ID, a.UserID, a.Label, a.FullName, a.Street, a.HouseNumber, a.Bus, a.City, a.PostalCode, a.Country, a.IsDefault)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *AddressRepo) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM addresses WHERE id = $1 AND user_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, id, userID)
	return err
}

func (r *AddressRepo) GetDefault(ctx context.Context, userID string) (*Address, error) {
	query := `
		SELECT id, user_id, label, full_name, street, house_number, bus, city, postal_code, country, is_default, created_at, updated_at
		FROM addresses
		WHERE user_id = $1 AND is_default = true
		LIMIT 1
	`
	var a Address
	err := r.db.Pool.QueryRow(ctx, query, userID).
		Scan(&a.ID, &a.UserID, &a.Label, &a.FullName, &a.Street, &a.HouseNumber, &a.Bus, &a.City, &a.PostalCode, &a.Country, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}