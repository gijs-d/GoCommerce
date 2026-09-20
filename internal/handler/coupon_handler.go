package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/repository"
)

type CouponHandler struct {
	couponRepo *repository.CouponRepo
}

func NewCouponHandler(couponRepo *repository.CouponRepo) *CouponHandler {
	return &CouponHandler{couponRepo: couponRepo}
}

type ValidateCouponRequest struct {
	Code     string  `json:"code"`
	Subtotal float64 `json:"subtotal"`
}

func (h *CouponHandler) ValidateCoupon(w http.ResponseWriter, r *http.Request) {
	var req ValidateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		api.Error(w, http.StatusBadRequest, "Ongeldige coupon aanvraag")
		return
	}

	coupon, discountAmount, err := h.couponRepo.GetValidCoupon(r.Context(), req.Code, req.Subtotal)
	if err != nil {
		api.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	api.Success(w, http.StatusOK, "Kortingscode succesvol toegepast", map[string]interface{}{
		"valid":           true,
		"coupon_id":       coupon.ID,
		"code":            coupon.Code,
		"discount_type":   coupon.DiscountType,
		"discount_value":  coupon.DiscountValue,
		"discount_amount": discountAmount,
	})
}

func (h *CouponHandler) ListCoupons(w http.ResponseWriter, r *http.Request) {
	coupons, err := h.couponRepo.ListAll(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen kortingscodes")
		return
	}
	api.JSON(w, http.StatusOK, coupons)
}

type CouponPayload struct {
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	MinSpend      *float64   `json:"min_spend"`
	MaxUses       *int       `json:"max_uses"`
	ExpiresAt     *time.Time `json:"expires_at"`
	IsActive      bool       `json:"is_active"`
}

func (h *CouponHandler) CreateCoupon(w http.ResponseWriter, r *http.Request) {
	var req CouponPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" || req.DiscountValue <= 0 {
		api.Error(w, http.StatusBadRequest, "Code en geldige kortingswaarde zijn verplicht")
		return
	}

	coupon, err := h.couponRepo.Create(r.Context(), req.Code, req.DiscountType, req.DiscountValue, req.MinSpend, req.MaxUses, req.ExpiresAt, req.IsActive)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon kortingscode niet aanmaken: "+err.Error())
		return
	}

	api.Success(w, http.StatusCreated, "Kortingscode aangemaakt", coupon)
}

func (h *CouponHandler) UpdateCoupon(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req CouponPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige invoer")
		return
	}

	err := h.couponRepo.Update(r.Context(), id, req.Code, req.DiscountType, req.DiscountValue, req.MinSpend, req.MaxUses, req.ExpiresAt, req.IsActive)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon kortingscode niet bijwerken: "+err.Error())
		return
	}

	api.Success(w, http.StatusOK, "Kortingscode bijgewerkt", nil)
}

func (h *CouponHandler) DeleteCoupon(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.couponRepo.Delete(r.Context(), id); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon kortingscode niet verwijderen")
		return
	}
	api.Success(w, http.StatusOK, "Kortingscode verwijderd", nil)
}