package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"webshop/internal/repository"
)

type StripeClient struct {
	settingsRepo *repository.SettingsRepo
	httpClient   *http.Client
}

func NewStripeClient(settingsRepo *repository.SettingsRepo) *StripeClient {
	return &StripeClient{
		settingsRepo: settingsRepo,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *StripeClient) getSecretKey(ctx context.Context) string {
	raw, err := s.settingsRepo.Get(ctx, "stripe_secret_key")
	if err != nil || len(raw) == 0 {
		return ""
	}
	var key string
	_ = json.Unmarshal(raw, &key)
	return key
}

// CreateCheckoutSession start een officiële Stripe Checkout sessie (Creditcard / Apple Pay / Google Pay)
func (s *StripeClient) CreateCheckoutSession(ctx context.Context, order *repository.Order, baseURL string) (*PaymentInitResult, error) {
	secretKey := s.getSecretKey(ctx)
	if secretKey == "" {
		return nil, fmt.Errorf("stripe secret key ontbreekt in instellingen")
	}

	data := url.Values{}
	data.Set("payment_method_types[0]", "card")
	data.Set("mode", "payment")
	data.Set("client_reference_id", order.OrderNumber)
	data.Set("success_url", fmt.Sprintf("%s/track/%s?payment=success", baseURL, order.OrderNumber))
	data.Set("cancel_url", fmt.Sprintf("%s/checkout", baseURL))
	data.Set("metadata[order_number]", order.OrderNumber)

	amountInCents := int64(order.TotalAmount * 100)
	data.Set("line_items[0][price_data][currency]", "eur")
	data.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(amountInCents, 10))
	data.Set("line_items[0][price_data][product_data][name]", fmt.Sprintf("Bestelling %s", order.OrderNumber))
	data.Set("line_items[0][quantity]", "1")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.stripe.com/v1/checkout/sessions", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+secretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stripe fout: %s", string(bodyBytes))
	}

	var session struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(bodyBytes, &session); err != nil {
		return nil, err
	}

	return &PaymentInitResult{
		Success:      true,
		RedirectURL:  session.URL,
		PaymentID:    session.ID,
		RequiresWait: true,
		OrderNumber:  order.OrderNumber,
	}, nil
}