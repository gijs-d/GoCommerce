package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/repository"
)

type ApiKeyHandler struct {
	apiKeyRepo *repository.ApiKeyRepo
}

func NewApiKeyHandler(apiKeyRepo *repository.ApiKeyRepo) *ApiKeyHandler {
	return &ApiKeyHandler{apiKeyRepo: apiKeyRepo}
}

func (h *ApiKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	keys, err := h.apiKeyRepo.ListAll(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen API-sleutels")
		return
	}
	api.JSON(w, http.StatusOK, keys)
}

type CreateKeyRequest struct {
	Description string `json:"description"`
	Permissions string `json:"permissions"`
}

func (h *ApiKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Description == "" {
		api.Error(w, http.StatusBadRequest, "Omschrijving is verplicht")
		return
	}

	consumerKey, consumerSecret, key, err := h.apiKeyRepo.GenerateKey(r.Context(), req.Description, req.Permissions)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon API-sleutel niet aanmaken: "+err.Error())
		return
	}

	api.Success(w, http.StatusCreated, "WooCommerce API-sleutel aangemaakt", map[string]interface{}{
		"key":             key,
		"consumer_key":    consumerKey,
		"consumer_secret": consumerSecret,
	})
}

func (h *ApiKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.apiKeyRepo.Delete(r.Context(), id); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon API-sleutel niet intrekken")
		return
	}
	api.Success(w, http.StatusOK, "API-sleutel ingetrokken", nil)
}