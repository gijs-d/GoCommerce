package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/config"
	"webshop/internal/repository"
)

type WcV3Handler struct {
	cfg         *config.Config
	productRepo *repository.ProductRepo
	orderRepo   *repository.OrderRepo
}

func NewWcV3Handler(cfg *config.Config, productRepo *repository.ProductRepo, orderRepo *repository.OrderRepo) *WcV3Handler {
	return &WcV3Handler{
		cfg:         cfg,
		productRepo: productRepo,
		orderRepo:   orderRepo,
	}
}

// SystemStatus wordt gebruikt door externe tools zoals Gelato/Printful om de verbinding te verifiëren
func (h *WcV3Handler) SystemStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"environment": map[string]interface{}{
			"home_url":         h.cfg.BaseURL,
			"site_url":         h.cfg.BaseURL,
			"version":          "8.9.0", // Nabootsen van recente stabiele WooCommerce versie
			"log_directory":    "/tmp",
			"database_version": "8.9.0",
		},
		"database": map[string]interface{}{
			"database_engine": "PostgreSQL 17 (High Performance)",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

// ListProducts exporteert producten in de officiële WooCommerce REST API v3 JSON-structuur
func (h *WcV3Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 50
	if l, err := strconv.Atoi(q.Get("per_page")); err == nil && l > 0 {
		limit = l
	}

	filter := repository.ProductFilter{
		Limit:      limit,
		ActiveOnly: false,
	}

	prods, _, err := h.productRepo.ListProducts(r.Context(), filter)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	var wcProducts []map[string]interface{}
	for _, p := range prods {
		fullProd, _ := h.productRepo.GetBySlug(r.Context(), p.Slug)
		if fullProd == nil {
			fullProd = &p
		}

		images := []map[string]interface{}{}
		for _, img := range fullProd.Images {
			images = append(images, map[string]interface{}{
				"id":  img.ID,
				"src": img.URL,
				"alt": img.AltText,
			})
		}
		if len(images) == 0 && fullProd.PrimaryImage != nil {
			images = append(images, map[string]interface{}{
				"id":  "primary",
				"src": *fullProd.PrimaryImage,
			})
		}

		variants := []map[string]interface{}{}
		for _, v := range fullProd.Variants {
			priceStr := fmt.Sprintf("%.2f", fullProd.BasePrice)
			if v.PriceOverride != nil {
				priceStr = fmt.Sprintf("%.2f", *v.PriceOverride)
			}
			variants = append(variants, map[string]interface{}{
				"id":             v.ID,
				"sku":            v.SKU,
				"price":          priceStr,
				"regular_price":  priceStr,
				"stock_quantity": v.StockQuantity,
				"manage_stock":   true,
				"attributes": []map[string]string{
					{"name": "Size", "option": *v.Size},
					{"name": "Color", "option": *v.Color},
				},
				"image": map[string]interface{}{
					"src": v.ImageURL,
				},
			})
		}

		priceStr := fmt.Sprintf("%.2f", fullProd.BasePrice)
		wcProd := map[string]interface{}{
			"id":                fullProd.ID,
			"name":              fullProd.Name,
			"slug":              fullProd.Slug,
			"permalink":         fmt.Sprintf("%s/product/%s", h.cfg.BaseURL, fullProd.Slug),
			"type":              "variable",
			"status":            "publish",
			"featured":          fullProd.Featured,
			"description":       fullProd.Description,
			"short_description": fullProd.ShortDescription,
			"sku":               fullProd.Slug,
			"price":             priceStr,
			"regular_price":     priceStr,
			"manage_stock":      false,
			"images":            images,
			"variations":        variants,
			"categories": []map[string]interface{}{
				{"id": fullProd.CategoryID, "name": fullProd.CategoryName, "slug": fullProd.CategorySlug},
			},
		}
		wcProducts = append(wcProducts, wcProd)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(wcProducts)
}

// GetProduct haalt één product op in WooCommerce v3 formaat
func (h *WcV3Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	slugOrID := chi.URLParam(r, "id")
	p, err := h.productRepo.GetBySlug(r.Context(), slugOrID)
	if err != nil || p == nil {
		api.Error(w, http.StatusNotFound, "Product niet gevonden")
		return
	}

	priceStr := fmt.Sprintf("%.2f", p.BasePrice)
	images := []map[string]interface{}{}
	for _, img := range p.Images {
		images = append(images, map[string]interface{}{
			"id":  img.ID,
			"src": img.URL,
		})
	}

	variations := []map[string]interface{}{}
	for _, v := range p.Variants {
		vPrice := priceStr
		if v.PriceOverride != nil {
			vPrice = fmt.Sprintf("%.2f", *v.PriceOverride)
		}
		variations = append(variations, map[string]interface{}{
			"id":             v.ID,
			"sku":            v.SKU,
			"price":          vPrice,
			"regular_price":  vPrice,
			"stock_quantity": v.StockQuantity,
		})
	}

	wcProd := map[string]interface{}{
		"id":            p.ID,
		"name":          p.Name,
		"slug":          p.Slug,
		"type":          "variable",
		"status":        "publish",
		"price":         priceStr,
		"regular_price": priceStr,
		"description":   p.Description,
		"images":        images,
		"variations":    variations,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(wcProd)
}

// ListOrders geeft bestellingen terug in de officiële WooCommerce v3 structuur
func (h *WcV3Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders, _, err := h.orderRepo.ListAll(r.Context(), "", 100, 0)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	var wcOrders []map[string]interface{}
	for _, o := range orders {
		fullOrder, _ := h.orderRepo.GetByOrderNumber(r.Context(), o.OrderNumber)
		if fullOrder == nil {
			fullOrder = &o
		}

		var lineItems []map[string]interface{}
		for _, it := range fullOrder.Items {
			lineItems = append(lineItems, map[string]interface{}{
				"id":           it.ID,
				"name":         it.ProductName,
				"product_id":   it.ProductID,
				"variation_id": it.VariantID,
				"quantity":     it.Quantity,
				"sku":          it.SKU,
				"price":        it.UnitPrice,
				"total":        fmt.Sprintf("%.2f", it.TotalPrice),
			})
		}

		var shippingAddr map[string]interface{}
		_ = json.Unmarshal(fullOrder.ShippingAddress, &shippingAddr)

		wcOrders = append(wcOrders, map[string]interface{}{
			"id":               fullOrder.ID,
			"number":           fullOrder.OrderNumber,
			"status":           fullOrder.Status,
			"currency":         "EUR",
			"date_created":     fullOrder.CreatedAt.Format(time.RFC3339),
			"total":            fmt.Sprintf("%.2f", fullOrder.TotalAmount),
			"shipping_total":   fmt.Sprintf("%.2f", fullOrder.ShippingCost),
			"payment_method":   fullOrder.PaymentProvider,
			"billing":          shippingAddr,
			"shipping":         shippingAddr,
			"line_items":       lineItems,
			"tracking_number":  fullOrder.TrackingCode,
			"tracking_carrier": fullOrder.Carrier,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(wcOrders)
}

// UpdateOrder staat externe partijen toe om trackingcode en status bij te werken
func (h *WcV3Handler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var payload struct {
		Status   string `json:"status"`
		Tracking struct {
			Number  string `json:"number"`
			Carrier string `json:"carrier"`
		} `json:"tracking"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige WooCommerce order data")
		return
	}

	var trackNum, carrier *string
	if payload.Tracking.Number != "" {
		trackNum = &payload.Tracking.Number
	}
	if payload.Tracking.Carrier != "" {
		carrier = &payload.Tracking.Carrier
	}

	msg := "Bijgewerkt via WooCommerce REST API v3"
	if payload.Status != "" {
		_ = h.orderRepo.UpdateStatus(r.Context(), id, payload.Status, msg, trackNum, carrier)
	}

	api.Success(w, http.StatusOK, "Order bijgewerkt", nil)
}