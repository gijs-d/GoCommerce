package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/repository"
)

type ProductHandler struct {
	productRepo *repository.ProductRepo
}

func NewProductHandler(productRepo *repository.ProductRepo) *ProductHandler {
	return &ProductHandler{productRepo: productRepo}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := 24
	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	page := 1
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		page = p
	}
	offset := (page - 1) * limit

	var minPrice *float64
	if minStr := q.Get("min_price"); minStr != "" {
		if val, err := strconv.ParseFloat(minStr, 64); err == nil {
			minPrice = &val
		}
	}

	var maxPrice *float64
	if maxStr := q.Get("max_price"); maxStr != "" {
		if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
			maxPrice = &val
		}
	}

	filter := repository.ProductFilter{
		CategorySlug: q.Get("category"),
		Search:       q.Get("search"),
		MinPrice:     minPrice,
		MaxPrice:     maxPrice,
		Size:         q.Get("size"),
		Color:        q.Get("color"),
		FeaturedOnly: q.Get("featured") == "true",
		ActiveOnly:   true,
		SortBy:       q.Get("sort"),
		Limit:        limit,
		Offset:       offset,
	}

	products, total, err := h.productRepo.ListProducts(r.Context(), filter)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen van producten")
		return
	}

	api.JSON(w, http.StatusOK, map[string]interface{}{
		"products":    products,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + limit - 1) / limit,
	})
}

func (h *ProductHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		api.Error(w, http.StatusBadRequest, "Slug ontbreekt")
		return
	}

	product, err := h.productRepo.GetBySlug(r.Context(), slug)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen product")
		return
	}
	if product == nil || !product.IsActive {
		api.Error(w, http.StatusNotFound, "Product niet gevonden")
		return
	}

	api.JSON(w, http.StatusOK, product)
}