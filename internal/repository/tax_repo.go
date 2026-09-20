package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"webshop/internal/database"
)

type TaxRate struct {
	ID          string    `json:"id"`
	CountryCode string    `json:"country_code"`
	StateCode   string    `json:"state_code,omitempty"`
	Rate        float64   `json:"rate"`
	Name        string    `json:"name"`
	IsCompound  bool      `json:"is_compound"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
}

type TaxRepo struct {
	db           *database.DB
	settingsRepo *SettingsRepo
}

func NewTaxRepo(db *database.DB, settingsRepo *SettingsRepo) *TaxRepo {
	return &TaxRepo{
		db:           db,
		settingsRepo: settingsRepo,
	}
}

// ensureTable zorgt dat de tabel en de unieke index gegarandeerd bestaan, inclusief ontbrekende kolommen
func (r *TaxRepo) ensureTable(ctx context.Context) {
	_, _ = r.db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tax_rates (
			id UUID PRIMARY KEY DEFAULT uuidv7(),
			country_code VARCHAR(2) NOT NULL,
			rate NUMERIC(5, 2) NOT NULL,
			name TEXT NOT NULL,
			is_compound BOOLEAN NOT NULL DEFAULT false,
			priority INT NOT NULL DEFAULT 1,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		ALTER TABLE tax_rates ADD COLUMN IF NOT EXISTS state_code VARCHAR(10);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_tax_rates_country_unique ON tax_rates (country_code);
	`)
}

// GetRateForCountry zoekt het geldende Btw-tarief voor een ISO landcode
func (r *TaxRepo) GetRateForCountry(ctx context.Context, countryCode string) (*TaxRate, error) {
	r.ensureTable(ctx)
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	query := `
		SELECT id, country_code, rate, name, is_compound, priority, created_at
		FROM tax_rates
		WHERE country_code = $1
		ORDER BY priority ASC
		LIMIT 1
	`
	t := &TaxRate{}
	err := r.db.Pool.QueryRow(ctx, query, countryCode).Scan(
		&t.ID, &t.CountryCode, &t.Rate, &t.Name, &t.IsCompound, &t.Priority, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if !isEUCountry(countryCode) {
				return &TaxRate{
					CountryCode: countryCode,
					Rate:        0.0,
					Name:        "Export buiten EU (0% Btw)",
				}, nil
			}
			return &TaxRate{
				CountryCode: countryCode,
				Rate:        21.0,
				Name:        "Standaard Btw (21%)",
			}, nil
		}
		return nil, err
	}
	return t, nil
}

func isEUCountry(code string) bool {
	euCodes := map[string]bool{
		"AT": true, "BE": true, "BG": true, "HR": true, "CY": true, "CZ": true,
		"DK": true, "EE": true, "FI": true, "FR": true, "DE": true, "GR": true,
		"HU": true, "IE": true, "IT": true, "LV": true, "LT": true, "LU": true,
		"MT": true, "NL": true, "PL": true, "PT": true, "RO": true, "SK": true,
		"SI": true, "ES": true, "SE": true,
	}
	return euCodes[code]
}

func (r *TaxRepo) ListAll(ctx context.Context) ([]TaxRate, error) {
	r.ensureTable(ctx)
	query := `
		SELECT id, country_code, rate, name, is_compound, priority, created_at
		FROM tax_rates
		ORDER BY country_code ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return []TaxRate{}, err
	}
	defer rows.Close()

	rates := []TaxRate{}
	for rows.Next() {
		var t TaxRate
		if err := rows.Scan(&t.ID, &t.CountryCode, &t.Rate, &t.Name, &t.IsCompound, &t.Priority, &t.CreatedAt); err == nil {
			rates = append(rates, t)
		}
	}
	return rates, nil
}

func (r *TaxRepo) UpsertRate(ctx context.Context, countryCode, name string, rate float64) error {
	r.ensureTable(ctx)
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	query := `
		INSERT INTO tax_rates (country_code, name, rate)
		VALUES ($1, $2, $3)
		ON CONFLICT (country_code) DO UPDATE
		SET name = EXCLUDED.name, rate = EXCLUDED.rate
	`
	_, err := r.db.Pool.Exec(ctx, query, countryCode, name, rate)
	return err
}

func (r *TaxRepo) DeleteRate(ctx context.Context, id string) error {
	query := `DELETE FROM tax_rates WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}