package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/repository"
)

type CategoryHandler struct {
	categoryRepo *repository.CategoryRepo
}

func NewCategoryHandler(categoryRepo *repository.CategoryRepo) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryRepo.ListActive(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen categorieën")
		return
	}

	api.JSON(w, http.StatusOK, categories)
}

func (h *CategoryHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		api.Error(w, http.StatusBadRequest, "Slug ontbreekt")
		return
	}

	cat, err := h.categoryRepo.GetBySlug(r.Context(), slug)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen categorie")
		return
	}
	if cat == nil || !cat.IsActive {
		api.Error(w, http.StatusNotFound, "Categorie niet gevonden")
		return
	}

	api.JSON(w, http.StatusOK, cat)
}