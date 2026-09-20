package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"webshop/internal/database"
)

type Dispatcher struct {
	db         *database.DB
	httpClient *http.Client
}

func NewDispatcher(db *database.DB) *Dispatcher {
	return &Dispatcher{
		db:         db,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type WebhookEvent struct {
	Event     string      `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Dispatch stuurt op de achtergrond een webhook naar alle actieve ontvangers
func (d *Dispatcher) Dispatch(ctx context.Context, eventName string, payload interface{}) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		query := `
			SELECT target_url, secret
			FROM webhooks
			WHERE is_active = true AND $1 = ANY(events)
		`
		rows, err := d.db.Pool.Query(bgCtx, query, eventName)
		if err != nil {
			return
		}
		defer rows.Close()

		eventData := WebhookEvent{
			Event:     eventName,
			Timestamp: time.Now(),
			Data:      payload,
		}
		body, err := json.Marshal(eventData)
		if err != nil {
			return
		}

		for rows.Next() {
			var targetURL string
			var secret *string
			if err := rows.Scan(&targetURL, &secret); err != nil {
				continue
			}

			req, err := http.NewRequestWithContext(bgCtx, "POST", targetURL, bytes.NewReader(body))
			if err != nil {
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Shop-Event", eventName)

			if secret != nil && *secret != "" {
				h := hmac.New(sha256.New, []byte(*secret))
				h.Write(body)
				req.Header.Set("X-Shop-Signature", hex.EncodeToString(h.Sum(nil)))
			}

			resp, err := d.httpClient.Do(req)
			if err != nil {
				log.Printf("[WEBHOOK FOUT] Kon niet sturen naar %s: %v", targetURL, err)
				continue
			}
			_ = resp.Body.Close()
		}
	}()
}