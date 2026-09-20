package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/repository"
	"webshop/internal/tax"
)

type TaxHandler struct {
	taxService *tax.Service
	taxRepo    *repository.TaxRepo
}

func NewTaxHandler(taxService *tax.Service, taxRepo *repository.TaxRepo) *TaxHandler {
	return &TaxHandler{
		taxService: taxService,
		taxRepo:    taxRepo,
	}
}

type CalculateTaxRequest struct {
	CountryCode string  `json:"country_code"`
	Subtotal    float64 `json:"subtotal"`
	VatNumber   string  `json:"vat_number,omitempty"`
}

func (h *TaxHandler) CalculateTax(w http.ResponseWriter, r *http.Request) {
	var req CalculateTaxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige aanvraaggegevens")
		return
	}

	result, err := h.taxService.CalculateTax(r.Context(), req.Subtotal, req.CountryCode, req.VatNumber)
	if err != nil {
		log.Printf("[TAX ERROR] Fout bij btw berekening: %v", err)
		api.Error(w, http.StatusInternalServerError, "Fout bij btw-berekening: "+err.Error())
		return
	}

	api.JSON(w, http.StatusOK, result)
}

func (h *TaxHandler) ValidateVIES(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VatNumber string `json:"vat_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.VatNumber == "" {
		api.Error(w, http.StatusBadRequest, "Btw-nummer is verplicht")
		return
	}

	vies, err := h.taxService.ValidateVIES(r.Context(), req.VatNumber)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij raadplegen VIES database")
		return
	}

	api.JSON(w, http.StatusOK, vies)
}

func (h *TaxHandler) ListRates(w http.ResponseWriter, r *http.Request) {
	rates, err := h.taxRepo.ListAll(r.Context())
	if err != nil {
		log.Printf("[TAX ERROR] Fout bij ophalen btw-tarieven: %v", err)
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen btw-tarieven: "+err.Error())
		return
	}
	api.JSON(w, http.StatusOK, rates)
}

func (h *TaxHandler) UpsertRate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CountryCode string  `json:"country_code"`
		Name        string  `json:"name"`
		Rate        float64 `json:"rate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CountryCode == "" {
		api.Error(w, http.StatusBadRequest, "Ongeldige tariefgegevens")
		return
	}

	if err := h.taxRepo.UpsertRate(r.Context(), req.CountryCode, req.Name, req.Rate); err != nil {
		log.Printf("[TAX ERROR] Fout bij opslaan btw-tarief: %v", err)
		api.Error(w, http.StatusInternalServerError, "Kon btw-tarief niet opslaan: "+err.Error())
		return
	}

	api.Success(w, http.StatusOK, "Btw-tarief opgeslagen", nil)
}

func (h *TaxHandler) DeleteRate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.taxRepo.DeleteRate(r.Context(), id); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon tarief niet verwijderen")
		return
	}
	api.Success(w, http.StatusOK, "Btw-tarief verwijderd", nil)
}