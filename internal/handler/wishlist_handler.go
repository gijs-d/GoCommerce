package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/middleware"
	"webshop/internal/repository"
)

type WishlistHandler struct {
	wishlistRepo *repository.WishlistRepo
}

func NewWishlistHandler(wishlistRepo *repository.WishlistRepo) *WishlistHandler {
	return &WishlistHandler{wishlistRepo: wishlistRepo}
}

func (h *WishlistHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		api.Error(w, http.StatusUnauthorized, "Inloggen vereist")
		return
	}

	items, err := h.wishlistRepo.ListByUser(r.Context(), user.ID)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen verlanglijst")
		return
	}

	api.JSON(w, http.StatusOK, items)
}

func (h *WishlistHandler) Add(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		api.Error(w, http.StatusUnauthorized, "Inloggen vereist")
		return
	}

	productID := chi.URLParam(r, "productId")
	if productID == "" {
		api.Error(w, http.StatusBadRequest, "Product ID ontbreekt")
		return
	}

	if err := h.wishlistRepo.Add(r.Context(), user.ID, productID); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon product niet toevoegen aan favorieten")
		return
	}

	api.Success(w, http.StatusOK, "Toegevoegd aan favorieten", nil)
}

func (h *WishlistHandler) Remove(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		api.Error(w, http.StatusUnauthorized, "Inloggen vereist")
		return
	}

	productID := chi.URLParam(r, "productId")
	if productID == "" {
		api.Error(w, http.StatusBadRequest, "Product ID ontbreekt")
		return
	}

	if err := h.wishlistRepo.Remove(r.Context(), user.ID, productID); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon product niet verwijderen uit favorieten")
		return
	}

	api.Success(w, http.StatusOK, "Verwijderd uit favorieten", nil)
}

func (h *WishlistHandler) Check(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		api.JSON(w, http.StatusOK, map[string]bool{"is_wishlisted": false})
		return
	}

	productID := chi.URLParam(r, "productId")
	exists, err := h.wishlistRepo.IsWishlisted(r.Context(), user.ID, productID)
	if err != nil {
		api.JSON(w, http.StatusOK, map[string]bool{"is_wishlisted": false})
		return
	}

	api.JSON(w, http.StatusOK, map[string]bool{"is_wishlisted": exists})
}