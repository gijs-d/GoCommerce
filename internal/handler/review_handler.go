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

type ReviewHandler struct {
	reviewRepo *repository.ReviewRepo
}

func NewReviewHandler(reviewRepo *repository.ReviewRepo) *ReviewHandler {
	return &ReviewHandler{reviewRepo: reviewRepo}
}

func (h *ReviewHandler) ListByProduct(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	reviews, err := h.reviewRepo.ListApproved(r.Context(), productID)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen reviews")
		return
	}

	stats, _ := h.reviewRepo.GetStats(r.Context(), productID)

	api.JSON(w, http.StatusOK, map[string]interface{}{
		"reviews": reviews,
		"stats":   stats,
	})
}

type CreateReviewRequest struct {
	AuthorName string `json:"author_name"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
}

func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	var req CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige review data")
		return
	}

	req.AuthorName = strings.TrimSpace(req.AuthorName)
	req.Comment = strings.TrimSpace(req.Comment)

	if req.Rating < 1 || req.Rating > 5 || req.Comment == "" {
		api.Error(w, http.StatusBadRequest, "Een score tussen 1 en 5 sterren en een toelichting zijn verplicht")
		return
	}

	var userID *string
	user := middleware.GetUserFromContext(r.Context())
	if user != nil {
		userID = &user.ID
		if req.AuthorName == "" {
			req.AuthorName = user.FullName
		}
	}

	if req.AuthorName == "" {
		req.AuthorName = "Klant"
	}

	review, err := h.reviewRepo.Create(r.Context(), productID, userID, req.AuthorName, req.Rating, req.Comment)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon review niet opslaan")
		return
	}

	api.Success(w, http.StatusCreated, "Bedankt voor je review!", review)
}

func (h *ReviewHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	reviews, err := h.reviewRepo.ListAll(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen reviews")
		return
	}
	api.JSON(w, http.StatusOK, reviews)
}

func (h *ReviewHandler) ApproveReview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		IsApproved bool `json:"is_approved"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.reviewRepo.SetApproval(r.Context(), id, req.IsApproved); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon status niet bijwerken")
		return
	}
	api.Success(w, http.StatusOK, "Review status bijgewerkt", nil)
}

func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.reviewRepo.Delete(r.Context(), id); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon review niet verwijderen")
		return
	}
	api.Success(w, http.StatusOK, "Review verwijderd", nil)
}