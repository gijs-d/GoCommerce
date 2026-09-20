package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/middleware"
	"webshop/internal/repository"
)

import (
	"webshop/internal/pod"
	"webshop/internal/webhook"
)

type OrderHandler struct {
	orderRepo   *repository.OrderRepo
	gelato      *pod.GelatoClient
	dispatcher  *webhook.Dispatcher
}

func NewOrderHandler(orderRepo *repository.OrderRepo, gelato *pod.GelatoClient, dispatcher *webhook.Dispatcher) *OrderHandler {
	return &OrderHandler{
		orderRepo:  orderRepo,
		gelato:     gelato,
		dispatcher: dispatcher,
	}
}

type CreateOrderRequest struct {
	GuestEmail      *string                      `json:"guest_email,omitempty"`
	ShippingAddress json.RawMessage              `json:"shipping_address"`
	BillingAddress  json.RawMessage              `json:"billing_address"`
	Items           []repository.OrderItemInput `json:"items"`
	ShippingCost    float64                      `json:"shipping_cost"`
	PaymentProvider string                       `json:"payment_provider"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige aanvraaggegevens")
		return
	}

	if len(req.Items) == 0 {
		api.Error(w, http.StatusBadRequest, "Winkelmandje is leeg")
		return
	}

	var userID *string
	currentUser := middleware.GetUserFromContext(r.Context())
	if currentUser != nil {
		userID = &currentUser.ID
	} else {
		if req.GuestEmail == nil || strings.TrimSpace(*req.GuestEmail) == "" {
			api.Error(w, http.StatusBadRequest, "E-mailadres is verplicht voor gastbestellingen")
			return
		}
		cleanedEmail := strings.TrimSpace(strings.ToLower(*req.GuestEmail))
		req.GuestEmail = &cleanedEmail
	}

	input := repository.OrderCreateInput{
		UserID:          userID,
		GuestEmail:      req.GuestEmail,
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		Items:           req.Items,
		ShippingCost:    req.ShippingCost,
		PaymentProvider: req.PaymentProvider,
	}

	order, err := h.orderRepo.CreateOrder(r.Context(), input)
	if err != nil {
		api.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	api.Success(w, http.StatusCreated, "Bestelling succesvol aangemaakt", order)
}

// TrackOrder haalt een order op voor de live bestelstatus pagina
func (h *OrderHandler) TrackOrder(w http.ResponseWriter, r *http.Request) {
	orderNumber := chi.URLParam(r, "orderNumber")
	if orderNumber == "" {
		api.Error(w, http.StatusBadRequest, "Bestelnummer ontbreekt")
		return
	}

	order, err := h.orderRepo.GetByOrderNumber(r.Context(), orderNumber)
	if err != nil || order == nil {
		api.Error(w, http.StatusNotFound, "Bestelling niet gevonden")
		return
	}

	api.JSON(w, http.StatusOK, order)
}

// ListUserOrders haalt bestelgeschiedenis op van ingelogde klant
func (h *OrderHandler) ListUserOrders(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		api.Error(w, http.StatusUnauthorized, "Inloggen vereist")
		return
	}

	orders, err := h.orderRepo.ListByUser(r.Context(), user.ID)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen van bestelgeschiedenis")
		return
	}

	api.JSON(w, http.StatusOK, orders)
}

// MockPay simuleert een succesvolle betaling (voor testdoeleinden)
func (h *OrderHandler) MockPay(w http.ResponseWriter, r *http.Request) {
	orderNumber := chi.URLParam(r, "orderNumber")
	order, err := h.orderRepo.GetByOrderNumber(r.Context(), orderNumber)
	if err != nil || order == nil {
		api.Error(w, http.StatusNotFound, "Bestelling niet gevonden")
		return
	}

	if order.PaymentStatus == "paid" {
		api.Success(w, http.StatusOK, "Bestelling is al betaald", order)
		return
	}

	_ = h.orderRepo.UpdatePaymentStatus(r.Context(), order.ID, "paid")
	_ = h.orderRepo.UpdateStatus(r.Context(), order.ID, "processing", "Betaling succesvol ontvangen (Mock Checkout)", nil, nil)

	// POD Fulfillment activeren: automatische doorsturing naar Gelato indien nodig
	if h.gelato != nil {
		_ = h.gelato.CheckAndFulfillOrder(r.Context(), order)
	}

	// Webhook notificatie schieten naar Zapier/Make/Boekhouding
	if h.dispatcher != nil {
		h.dispatcher.Dispatch(r.Context(), "order.paid", order)
	}

	updatedOrder, _ := h.orderRepo.GetByOrderNumber(r.Context(), orderNumber)
	api.Success(w, http.StatusOK, "Betaling succesvol verwerkt", updatedOrder)
}