package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"webshop/internal/api"
	"webshop/internal/database"
	"webshop/internal/repository"
)

type AdminHandler struct {
	db           *database.DB
	productRepo  *repository.ProductRepo
	categoryRepo *repository.CategoryRepo
	orderRepo    *repository.OrderRepo
	widgetRepo   *repository.WidgetRepo
	userRepo     *repository.UserRepo
}

func NewAdminHandler(
	db *database.DB,
	productRepo *repository.ProductRepo,
	categoryRepo *repository.CategoryRepo,
	orderRepo *repository.OrderRepo,
	widgetRepo *repository.WidgetRepo,
	userRepo *repository.UserRepo,
) *AdminHandler {
	return &AdminHandler{
		db:           db,
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		orderRepo:    orderRepo,
		widgetRepo:   widgetRepo,
		userRepo:     userRepo,
	}
}

// GetStats haalt WooCommerce KPI cijfers op
func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var totalRevenue float64
	var totalOrders int
	var totalCustomers int
	var lowStockCount int

	_ = h.db.Pool.QueryRow(ctx, `SELECT COALESCE(SUM(total_amount), 0), COUNT(*) FROM orders WHERE payment_status = 'paid'`).Scan(&totalRevenue, &totalOrders)
	_ = h.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'customer'`).Scan(&totalCustomers)
	_ = h.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM product_variants WHERE stock_quantity <= 5 AND is_active = true`).Scan(&lowStockCount)

	api.JSON(w, http.StatusOK, map[string]interface{}{
		"total_revenue":   totalRevenue,
		"total_orders":    totalOrders,
		"total_customers": totalCustomers,
		"low_stock_count": lowStockCount,
	})
}

// ListCustomers haalt klanten op met Lifetime Value en bestelaantallen
func (h *AdminHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := h.userRepo.ListCustomersWithStats(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen klanten: "+err.Error())
		return
	}
	api.JSON(w, http.StatusOK, customers)
}

// Product Management
type ProductPayload struct {
	CategoryID       *string `json:"category_id"`
	Name             string  `json:"name"`
	Slug             string  `json:"slug"`
	ShortDescription *string `json:"short_description"`
	Description      string  `json:"description"`
	BasePrice        float64 `json:"base_price"`
	Featured         bool    `json:"featured"`
	IsActive         bool    `json:"is_active"`
	SEOTitle         *string `json:"seo_title"`
	SEODescription   *string `json:"seo_description"`
}

func (h *AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req ProductPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige invoer")
		return
	}

	prod, err := h.productRepo.Create(
		r.Context(), req.CategoryID, req.Name, req.Slug, req.ShortDescription,
		req.Description, req.BasePrice, req.Featured, req.IsActive, req.SEOTitle, req.SEODescription,
	)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon product niet aanmaken: "+err.Error())
		return
	}

	api.Success(w, http.StatusCreated, "Product aangemaakt", prod)
}

func (h *AdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req ProductPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige invoer")
		return
	}

	err := h.productRepo.Update(
		r.Context(), id, req.CategoryID, req.Name, req.Slug, req.ShortDescription,
		req.Description, req.BasePrice, req.Featured, req.IsActive, req.SEOTitle, req.SEODescription,
	)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon product niet bijwerken: "+err.Error())
		return
	}

	api.Success(w, http.StatusOK, "Product bijgewerkt", nil)
}

func (h *AdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.productRepo.Delete(r.Context(), id); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon product niet verwijderen")
		return
	}
	api.Success(w, http.StatusOK, "Product verwijderd", nil)
}

func (h *AdminHandler) UpsertVariant(w http.ResponseWriter, r *http.Request) {
	var v repository.ProductVariant
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige variant")
		return
	}

	if err := h.productRepo.UpsertVariant(r.Context(), v); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon variant niet opslaan: "+err.Error())
		return
	}

	api.Success(w, http.StatusOK, "Variant opgeslagen", nil)
}

func (h *AdminHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.productRepo.DeleteVariant(r.Context(), id); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon variant niet verwijderen")
		return
	}
	api.Success(w, http.StatusOK, "Variant verwijderd", nil)
}

func (h *AdminHandler) AddProductImage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID string  `json:"product_id"`
		VariantID *string `json:"variant_id"`
		URL       string  `json:"url"`
		AltText   *string `json:"alt_text"`
		SortOrder int     `json:"sort_order"`
		IsPrimary bool    `json:"is_primary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID == "" || req.URL == "" {
		api.Error(w, http.StatusBadRequest, "ProductID en URL zijn verplicht")
		return
	}

	err := h.productRepo.AddImage(r.Context(), req.ProductID, req.VariantID, req.URL, req.AltText, req.SortOrder, req.IsPrimary)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon afbeelding niet koppelen")
		return
	}
	api.Success(w, http.StatusCreated, "Afbeelding gekoppeld", nil)
}

func (h *AdminHandler) DeleteProductImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.productRepo.DeleteImage(r.Context(), id); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon afbeelding niet verwijderen")
		return
	}
	api.Success(w, http.StatusOK, "Afbeelding ontkoppeld", nil)
}

// Categories
func (h *AdminHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.categoryRepo.ListAll(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.JSON(w, http.StatusOK, cats)
}

func (h *AdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		Slug        string  `json:"slug"`
		Description *string `json:"description"`
		ImageURL    *string `json:"image_url"`
		SortOrder   int     `json:"sort_order"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	cat, err := h.categoryRepo.Create(r.Context(), req.Name, req.Slug, req.Description, req.ImageURL, req.SortOrder)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.Success(w, http.StatusCreated, "Categorie aangemaakt", cat)
}

func (h *AdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name        string  `json:"name"`
		Slug        string  `json:"slug"`
		Description *string `json:"description"`
		ImageURL    *string `json:"image_url"`
		SortOrder   int     `json:"sort_order"`
		IsActive    bool    `json:"is_active"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	err := h.categoryRepo.Update(r.Context(), id, req.Name, req.Slug, req.Description, req.ImageURL, req.SortOrder, req.IsActive)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.Success(w, http.StatusOK, "Categorie bijgewerkt", nil)
}

func (h *AdminHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = h.categoryRepo.Delete(r.Context(), id)
	api.Success(w, http.StatusOK, "Categorie verwijderd", nil)
}

// Widgets
func (h *AdminHandler) ListWidgets(w http.ResponseWriter, r *http.Request) {
	widgets, err := h.widgetRepo.ListAll(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.JSON(w, http.StatusOK, widgets)
}

func (h *AdminHandler) UpsertWidget(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID        *string         `json:"id"`
		Type      string          `json:"type"`
		Title     *string         `json:"title"`
		Subtitle  *string         `json:"subtitle"`
		Config    json.RawMessage `json:"config"`
		SortOrder int             `json:"sort_order"`
		IsActive  bool            `json:"is_active"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	err := h.widgetRepo.Upsert(r.Context(), req.ID, req.Type, req.Title, req.Subtitle, req.Config, req.SortOrder, req.IsActive)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.Success(w, http.StatusOK, "Widget opgeslagen", nil)
}

func (h *AdminHandler) DeleteWidget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = h.widgetRepo.Delete(r.Context(), id)
	api.Success(w, http.StatusOK, "Widget verwijderd", nil)
}

// Orders
func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := q.Get("status")
	limit := 20
	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 { limit = l }
	page := 1
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 { page = p }
	offset := (page - 1) * limit

	orders, total, err := h.orderRepo.ListAll(r.Context(), status, limit, offset)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.JSON(w, http.StatusOK, map[string]interface{}{
		"orders": orders, "total": total, "page": page, "limit": limit,
	})
}

func (h *AdminHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Status       string  `json:"status"`
		Message      string  `json:"message"`
		TrackingCode *string `json:"tracking_code"`
		Carrier      *string `json:"carrier"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Message == "" { req.Message = "Status bijgewerkt naar " + req.Status }
	_ = h.orderRepo.UpdateStatus(r.Context(), id, req.Status, req.Message, req.TrackingCode, req.Carrier)
	api.Success(w, http.StatusOK, "Status bijgewerkt", nil)
}