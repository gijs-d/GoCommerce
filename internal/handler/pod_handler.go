package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"webshop/internal/api"
	"webshop/internal/repository"
)

type PODHandler struct {
	orderRepo *repository.OrderRepo
}

func NewPODHandler(orderRepo *repository.OrderRepo) *PODHandler {
	return &PODHandler{orderRepo: orderRepo}
}

type GelatoWebhookPayload struct {
	Event string `json:"event"` // bijv. "order_shipped" of "package_shipped"
	Order struct {
		OrderReferenceID string `json:"orderReferenceId"`
	} `json:"order"`
	Tracking struct {
		Code    string `json:"code"`
		Carrier string `json:"carrier"`
		URL     string `json:"url"`
	} `json:"tracking"`
}

// GelatoWebhook verwerkt binnenkomende updates van Gelato (bijv. tracking code zodra shirt geprint en verzonden is)
func (h *PODHandler) GelatoWebhook(w http.ResponseWriter, r *http.Request) {
	var payload GelatoWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige webhook data")
		return
	}

	orderNum := payload.Order.OrderReferenceID
	if orderNum == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	order, err := h.orderRepo.GetByOrderNumber(r.Context(), orderNum)
	if err != nil || order == nil {
		log.Printf("[GELATO WEBHOOK] Bestelling %s niet gevonden", orderNum)
		w.WriteHeader(http.StatusOK)
		return
	}

	if payload.Event == "order_shipped" || payload.Event == "package_shipped" || payload.Tracking.Code != "" {
		tracking := payload.Tracking.Code
		carrier := payload.Tracking.Carrier
		if carrier == "" {
			carrier = "PostNL / Gelato Express"
		}

		msg := "Bestelling geproduceerd en verzonden door Gelato drukkerij"
		_ = h.orderRepo.UpdateStatus(r.Context(), order.ID, "shipped", msg, &tracking, &carrier)
		log.Printf("[GELATO WEBHOOK] Order %s bijgewerkt naar 'shipped' met tracking: %s", orderNum, tracking)
	}

	w.WriteHeader(http.StatusOK)
}