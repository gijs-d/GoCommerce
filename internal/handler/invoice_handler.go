package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/invoice"
	"webshop/internal/repository"
)

type InvoiceHandler struct {
	generator *invoice.InvoiceGenerator
	orderRepo *repository.OrderRepo
}

func NewInvoiceHandler(generator *invoice.InvoiceGenerator, orderRepo *repository.OrderRepo) *InvoiceHandler {
	return &InvoiceHandler{
		generator: generator,
		orderRepo: orderRepo,
	}
}

// GetPublicInvoice toont de printklare factuur aan de klant
func (h *InvoiceHandler) GetPublicInvoice(w http.ResponseWriter, r *http.Request) {
	orderNumber := chi.URLParam(r, "orderNumber")
	order, err := h.orderRepo.GetByOrderNumber(r.Context(), orderNumber)
	if err != nil || order == nil {
		api.Error(w, http.StatusNotFound, "Factuur niet gevonden")
		return
	}

	htmlContent, err := h.generator.GenerateHTML(r.Context(), order)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij genereren factuur")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(htmlContent))
}

// GetAdminInvoice toont de factuur aan de beheerder via order ID
func (h *InvoiceHandler) GetAdminInvoice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orders, _, err := h.orderRepo.ListAll(r.Context(), "", 1, 0)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen")
		return
	}

	var targetOrder *repository.Order
	for _, o := range orders {
		if o.ID == id {
			targetOrder, _ = h.orderRepo.GetByOrderNumber(r.Context(), o.OrderNumber)
			break
		}
	}

	if targetOrder == nil {
		api.Error(w, http.StatusNotFound, "Bestelling niet gevonden")
		return
	}

	htmlContent, err := h.generator.GenerateHTML(r.Context(), targetOrder)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij genereren factuur")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(htmlContent))
}