package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/repository"
)

type ShippingHandler struct {
	shippingRepo *repository.ShippingRepo
}

func NewShippingHandler(shippingRepo *repository.ShippingRepo) *ShippingHandler {
	return &ShippingHandler{shippingRepo: shippingRepo}
}

func (h *ShippingHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	methods, err := h.shippingRepo.ListActive(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen verzendmethoden")
		return
	}
	api.JSON(w, http.StatusOK, methods)
}

func (h *ShippingHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	methods, err := h.shippingRepo.ListAll(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen verzendmethoden")
		return
	}
	api.JSON(w, http.StatusOK, methods)
}

type ShippingPayload struct {
	Title         string   `json:"title"`
	Description   *string  `json:"description"`
	Cost          float64  `json:"cost"`
	FreeThreshold *float64 `json:"free_threshold"`
	SortOrder     int      `json:"sort_order"`
	IsActive      bool     `json:"is_active"`
}

func (h *ShippingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req ShippingPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		api.Error(w, http.StatusBadRequest, "Titel is verplicht")
		return
	}

	method, err := h.shippingRepo.Create(r.Context(), req.Title, req.Description, req.Cost, req.FreeThreshold, req.SortOrder, req.IsActive)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon verzendmethode niet aanmaken")
		return
	}

	api.Success(w, http.StatusCreated, "Verzendmethode aangemaakt", method)
}

func (h *ShippingHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req ShippingPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige invoer")
		return
	}

	err := h.shippingRepo.Update(r.Context(), id, req.Title, req.Description, req.Cost, req.FreeThreshold, req.SortOrder, req.IsActive)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon verzendmethode niet bijwerken")
		return
	}

	api.Success(w, http.StatusOK, "Verzendmethode bijgewerkt", nil)
}

func (h *ShippingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.shippingRepo.Delete(r.Context(), id); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon verzendmethode niet verwijderen")
		return
	}
	api.Success(w, http.StatusOK, "Verzendmethode verwijderd", nil)
}