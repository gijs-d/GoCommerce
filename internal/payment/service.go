package payment

import (
	"context"
	"encoding/json"
	"fmt"

	"webshop/internal/config"
	"webshop/internal/repository"
)

type Service struct {
	cfg          *config.Config
	settingsRepo *repository.SettingsRepo
	orderRepo    *repository.OrderRepo
	mollie       *MollieClient
	stripe       *StripeClient
}

func NewService(cfg *config.Config, settingsRepo *repository.SettingsRepo, orderRepo *repository.OrderRepo) *Service {
	return &Service{
		cfg:          cfg,
		settingsRepo: settingsRepo,
		orderRepo:    orderRepo,
		mollie:       NewMollieClient(settingsRepo),
		stripe:       NewStripeClient(settingsRepo),
	}
}

// GetAvailableMethods geeft alle actieve betaalmethoden terug
func (s *Service) GetAvailableMethods(ctx context.Context) []PaymentMethodInfo {
	var methods []PaymentMethodInfo

	// 1. Directe Testbetaler (Mock)
	mockRaw, _ := s.settingsRepo.Get(ctx, "mock_payments_enabled")
	mockEnabled := true
	if len(mockRaw) > 0 {
		_ = json.Unmarshal(mockRaw, &mockEnabled)
	}

	if mockEnabled || s.cfg.AppEnv != "production" {
		methods = append(methods, PaymentMethodInfo{
			ID:          "mock",
			Title:       "Directe Testbetaling (Mock Gateway)",
			Description: "Simuleer direct een succesvolle betaling zonder bankpas",
			Provider:    "mock",
			Icon:        "shield",
		})
	}

	// 2. Mollie (iDEAL, Bancontact)
	mollieKeyRaw, _ := s.settingsRepo.Get(ctx, "mollie_api_key")
	var mollieKey string
	if len(mollieKeyRaw) > 0 {
		_ = json.Unmarshal(mollieKeyRaw, &mollieKey)
	}

	if mollieKey != "" {
		methods = append(methods, 
			PaymentMethodInfo{
				ID:          "mollie_ideal",
				Title:       "iDEAL",
				Description: "Betaal veilig via je eigen Nederlandse bank",
				Provider:    "mollie",
				Icon:        "ideal",
			},
			PaymentMethodInfo{
				ID:          "mollie_bancontact",
				Title:       "Bancontact",
				Description: "Betaal snel via de Bancontact-app of Belgische kaart",
				Provider:    "mollie",
				Icon:        "bancontact",
			},
		)
	}

	// 3. Stripe (Creditcard, Apple Pay)
	stripeKeyRaw, _ := s.settingsRepo.Get(ctx, "stripe_secret_key")
	var stripeKey string
	if len(stripeKeyRaw) > 0 {
		_ = json.Unmarshal(stripeKeyRaw, &stripeKey)
	}

	if stripeKey != "" {
		methods = append(methods, PaymentMethodInfo{
			ID:          "stripe_card",
			Title:       "Creditcard / Apple Pay (Stripe)",
			Description: "Visa, Mastercard, American Express & Apple Pay",
			Provider:    "stripe",
			Icon:        "credit-card",
		})
	}

	return methods
}

// InitiatePayment start de betaling bij de juiste provider
func (s *Service) InitiatePayment(ctx context.Context, orderNumber, method string) (*PaymentInitResult, error) {
	order, err := s.orderRepo.GetByOrderNumber(ctx, orderNumber)
	if err != nil || order == nil {
		return nil, fmt.Errorf("bestelling %s niet gevonden", orderNumber)
	}

	if order.PaymentStatus == "paid" {
		return &PaymentInitResult{
			Success:      true,
			RedirectURL:  fmt.Sprintf("%s/track/%s", s.cfg.BaseURL, order.OrderNumber),
			RequiresWait: false,
			OrderNumber:  order.OrderNumber,
		}, nil
	}

	// Provider dispatch
	switch {
	case method == "mock":
		_ = s.orderRepo.UpdatePaymentStatus(ctx, order.ID, "paid")
		_ = s.orderRepo.UpdateStatus(ctx, order.ID, "processing", "Testbetaling succesvol afgerond", nil, nil)
		return &PaymentInitResult{
			Success:      true,
			RedirectURL:  fmt.Sprintf("%s/track/%s?payment=success", s.cfg.BaseURL, order.OrderNumber),
			RequiresWait: false,
			OrderNumber:  order.OrderNumber,
		}, nil

	case method == "mollie_ideal" || method == "mollie_bancontact" || method == "mollie_creditcard":
		return s.mollie.CreatePayment(ctx, order, method, s.cfg.BaseURL)

	case method == "stripe" || method == "stripe_card":
		return s.stripe.CreateCheckoutSession(ctx, order, s.cfg.BaseURL)

	default:
		return nil, fmt.Errorf("onbekende betaalmethode: %s", method)
	}
}