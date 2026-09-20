package pod

import (
	"context"
	"webshop/internal/repository"
)

type ShipmentQuote struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Cost        float64 `json:"cost"`
	Currency    string  `json:"currency"`
	MinDays     int     `json:"min_days"`
	MaxDays     int     `json:"max_days"`
	Provider    string  `json:"provider"`
}

type PODItem struct {
	VariantID    string `json:"variant_id"`
	PODVariantID string `json:"pod_variant_id"`
	Quantity     int    `json:"quantity"`
	PrintFileURL string `json:"print_file_url"`
}

type PODAddress struct {
	Country      string `json:"country"`
	Postcode     string `json:"postcode"`
	City         string `json:"city"`
	AddressLine1 string `json:"address_line_1"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
}

type PODProvider interface {
	Name() string
	GetShipmentQuotes(ctx context.Context, dest PODAddress, items []PODItem) ([]ShipmentQuote, error)
	CheckAndFulfillOrder(ctx context.Context, order *repository.Order) error
}