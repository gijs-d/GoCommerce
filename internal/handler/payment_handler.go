package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/payment"
	"webshop/internal/pod"
	"webshop/internal/repository"
	"webshop/internal/webhook"
)

type PaymentHandler struct {
	paymentService *payment.Service
	orderRepo      *repository.OrderRepo
	mollieClient   *payment.MollieClient
	gelatoClient   *pod.GelatoClient
	dispatcher     *webhook.Dispatcher
}

func NewPaymentHandler(
	paymentService *payment.Service,
	orderRepo *repository.OrderRepo,
	settingsRepo *repository.SettingsRepo,
	gelatoClient *pod.GelatoClient,
	dispatcher *webhook.Dispatcher,
) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		orderRepo:      orderRepo,
		mollieClient:   payment.NewMollieClient(settingsRepo),
		gelatoClient:   gelatoClient,
		dispatcher:     dispatcher,
	}
}

// ListMethods geeft alle actieve betaalopties terug voor de checkout
func (h *PaymentHandler) ListMethods(w http.ResponseWriter, r *http.Request) {
	methods := h.paymentService.GetAvailableMethods(r.Context())
	api.JSON(w, http.StatusOK, methods)
}

type InitiatePaymentRequest struct {
	OrderNumber string `json:"order_number"`
	Method      string `json:"method"` // 'mock', 'mollie_ideal', 'mollie_bancontact', 'stripe_card'
}

// InitiatePayment start de betaalsessie
func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var req InitiatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OrderNumber == "" {
		api.Error(w, http.StatusBadRequest, "Ongeldige betaalaanvraag")
		return
	}

	result, err := h.paymentService.InitiatePayment(r.Context(), req.OrderNumber, req.Method)
	if err != nil {
		api.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	api.JSON(w, http.StatusOK, result)
}

// MollieWebhook verwerkt live betalingsstatussen van Mollie (iDEAL, Bancontact)
func (h *PaymentHandler) MollieWebhook(w http.ResponseWriter, r *http.Request) {
	paymentID := r.FormValue("id")
	if paymentID == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	status, orderNumber, err := h.mollieClient.VerifyPayment(r.Context(), paymentID)
	if err != nil || orderNumber == "" {
		log.Printf("[MOLLIE WEBHOOK] Fout bij verifiëren betaling %s: %v", paymentID, err)
		w.WriteHeader(http.StatusOK)
		return
	}

	if status == "paid" {
		order, err := h.orderRepo.GetByOrderNumber(r.Context(), orderNumber)
		if err == nil && order != nil && order.PaymentStatus != "paid" {
			_ = h.orderRepo.UpdatePaymentStatus(r.Context(), order.ID, "paid")
			_ = h.orderRepo.UpdateStatus(r.Context(), order.ID, "processing", "Betaling succesvol ontvangen via Mollie ("+paymentID+")", nil, nil)
			
			// Trigger Gelato POD doorsturing
			if h.gelatoClient != nil {
				_ = h.gelatoClient.CheckAndFulfillOrder(r.Context(), order)
			}
			// Trigger uitgaande webhooks
			if h.dispatcher != nil {
				h.dispatcher.Dispatch(r.Context(), "order.paid", order)
			}
			log.Printf("[MOLLIE SUCCESS] Bestelling %s succesvol gemarkeerd als betaald", orderNumber)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// StripeWebhook verwerkt live betalingsstatussen van Stripe Checkout
func (h *PaymentHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	var event struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ClientReferenceID string `json:"client_reference_id"`
				PaymentStatus     string `json:"payment_status"` // 'paid'
			} `json:"object"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if event.Type == "checkout.session.completed" && event.Data.Object.PaymentStatus == "paid" {
		orderNum := event.Data.Object.ClientReferenceID
		if orderNum != "" {
			order, err := h.orderRepo.GetByOrderNumber(r.Context(), orderNum)
			if err == nil && order != nil && order.PaymentStatus != "paid" {
				_ = h.orderRepo.UpdatePaymentStatus(r.Context(), order.ID, "paid")
				_ = h.orderRepo.UpdateStatus(r.Context(), order.ID, "processing", "Betaling succesvol voldaan via Stripe Checkout", nil, nil)

				if h.gelatoClient != nil {
					_ = h.gelatoClient.CheckAndFulfillOrder(r.Context(), order)
				}
				if h.dispatcher != nil {
					h.dispatcher.Dispatch(r.Context(), "order.paid", order)
				}
				log.Printf("[STRIPE SUCCESS] Bestelling %s gemarkeerd als betaald", orderNum)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// RefundOrder voert een terugbetaling uit
func (h *PaymentHandler) RefundOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Amount float64 `json:"amount"`
		Reason string  `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	msg := "Terugbetaling geregistreerd: € " + fmt.Sprintf("%.2f", req.Amount)
	if req.Reason != "" {
		msg += " (Reden: " + req.Reason + ")"
	}

	_ = h.orderRepo.UpdateStatus(r.Context(), id, "refunded", msg, nil, nil)
	_ = h.orderRepo.UpdatePaymentStatus(r.Context(), id, "refunded")

	api.Success(w, http.StatusOK, "Bestelling succesvol terugbetaald", nil)
}