package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/middleware"
	"webshop/internal/repository"
)

type AddressHandler struct {
	addressRepo *repository.AddressRepo
}

func NewAddressHandler(addressRepo *repository.AddressRepo) *AddressHandler {
	return &AddressHandler{addressRepo: addressRepo}
}

func (h *AddressHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		api.Error(w, http.StatusUnauthorized, "Inloggen vereist")
		return
	}

	addresses, err := h.addressRepo.ListByUser(r.Context(), user.ID)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen adressen")
		return
	}

	api.JSON(w, http.StatusOK, addresses)
}

func (h *AddressHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		api.Error(w, http.StatusUnauthorized, "Inloggen vereist")
		return
	}

	var a repository.Address
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige invoergegevens")
		return
	}

	a.UserID = user.ID
	if a.FullName == "" || a.Street == "" || a.HouseNumber == "" || a.City == "" || a.PostalCode == "" {
		api.Error(w, http.StatusBadRequest, "Alle verplichte adresvelden moeten ingevuld zijn")
		return
	}

	created, err := h.addressRepo.Create(r.Context(), &a)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon adres niet opslaan")
		return
	}

	api.Success(w, http.StatusCreated, "Adres succesvol toegevoegd", created)
}

func (h *AddressHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		api.Error(w, http.StatusUnauthorized, "Inloggen vereist")
		return
	}

	id := chi.URLParam(r, "id")
	var a repository.Address
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige invoergegevens")
		return
	}

	a.ID = id
	a.UserID = user.ID

	if err := h.addressRepo.Update(r.Context(), &a); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon adres niet bijwerken")
		return
	}

	api.Success(w, http.StatusOK, "Adres succesvol bijgewerkt", a)
}

func (h *AddressHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		api.Error(w, http.StatusUnauthorized, "Inloggen vereist")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.addressRepo.Delete(r.Context(), id, user.ID); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon adres niet verwijderen")
		return
	}

	api.Success(w, http.StatusOK, "Adres verwijderd", nil)
}