package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"webshop/internal/repository"
)

type MollieClient struct {
	settingsRepo *repository.SettingsRepo
	httpClient   *http.Client
}

func NewMollieClient(settingsRepo *repository.SettingsRepo) *MollieClient {
	return &MollieClient{
		settingsRepo: settingsRepo,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (m *MollieClient) getApiKey(ctx context.Context) string {
	raw, err := m.settingsRepo.Get(ctx, "mollie_api_key")
	if err != nil || len(raw) == 0 {
		return ""
	}
	var key string
	_ = json.Unmarshal(raw, &key)
	return key
}

// CreatePayment start een officiële Mollie iDEAL / Bancontact / Kaart betaling
func (m *MollieClient) CreatePayment(ctx context.Context, order *repository.Order, method, baseURL string) (*PaymentInitResult, error) {
	apiKey := m.getApiKey(ctx)
	if apiKey == "" {
		return nil, fmt.Errorf("mollie API-sleutel ontbreekt in de instellingen")
	}

	mollieMethod := ""
	switch method {
	case "mollie_ideal":
		mollieMethod = "ideal"
	case "mollie_bancontact":
		mollieMethod = "bancontact"
	case "mollie_creditcard":
		mollieMethod = "creditcard"
	}

	redirectURL := fmt.Sprintf("%s/track/%s", baseURL, order.OrderNumber)
	webhookURL := fmt.Sprintf("%s/api/webhooks/mollie", baseURL)

	payload := map[string]interface{}{
		"amount": map[string]string{
			"currency": "EUR",
			"value":    fmt.Sprintf("%.2f", order.TotalAmount),
		},
		"description": fmt.Sprintf("Bestelling %s", order.OrderNumber),
		"redirectUrl": redirectURL,
		"webhookUrl":  webhookURL,
		"metadata": map[string]string{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
		},
	}
	if mollieMethod != "" {
		payload["method"] = mollieMethod
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mollie.com/v2/payments", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Detail string `json:"detail"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("mollie API fout: %s", errResp.Detail)
	}

	var mResp struct {
		ID    string `json:"id"`
		Links struct {
			Checkout struct {
				HRef string `json:"href"`
			} `json:"checkout"`
		} `json:"_links"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return nil, err
	}

	return &PaymentInitResult{
		Success:      true,
		RedirectURL:  mResp.Links.Checkout.HRef,
		PaymentID:    mResp.ID,
		RequiresWait: true,
		OrderNumber:  order.OrderNumber,
	}, nil
}

// VerifyPayment haalt de status op bij Mollie tijdens webhook afhandeling
func (m *MollieClient) VerifyPayment(ctx context.Context, paymentID string) (status string, orderNumber string, err error) {
	apiKey := m.getApiKey(ctx)
	if apiKey == "" {
		return "", "", fmt.Errorf("mollie API-sleutel ontbreekt")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.mollie.com/v2/payments/"+paymentID, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var mResp struct {
		Status   string `json:"status"` // 'paid', 'canceled', 'expired', 'failed'
		Metadata struct {
			OrderNumber string `json:"order_number"`
		} `json:"metadata"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return "", "", err
	}

	return mResp.Status, mResp.Metadata.OrderNumber, nil
}