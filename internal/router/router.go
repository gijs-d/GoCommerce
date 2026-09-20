package router

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"webshop/internal/config"
	"webshop/internal/handler"
	"webshop/internal/middleware"
	"webshop/internal/repository"
	"webshop/internal/seo"
	"webshop/internal/theme"
)

type RouterParams struct {
	Cfg              *config.Config
	UserRepo         *repository.UserRepo
	ProductRepo      *repository.ProductRepo
	CategoryRepo     *repository.CategoryRepo
	SettingsRepo     *repository.SettingsRepo
	ApiKeyRepo       *repository.ApiKeyRepo
	ThemeManager     *theme.ThemeManager
	AuthHandler      *handler.AuthHandler
	ProductHandler   *handler.ProductHandler
	CategoryHandler  *handler.CategoryHandler
	WidgetHandler    *handler.WidgetHandler
	OrderHandler     *handler.OrderHandler
	WishlistHandler  *handler.WishlistHandler
	AddressHandler   *handler.AddressHandler
	MediaHandler     *handler.MediaHandler
	AdminHandler     *handler.AdminHandler
	SettingsHandler  *handler.SettingsHandler
	CouponHandler    *handler.CouponHandler
	ShippingHandler  *handler.ShippingHandler
	ReviewHandler    *handler.ReviewHandler
	PODHandler       *handler.PODHandler
	ApiKeyHandler    *handler.ApiKeyHandler
	WcV3Handler      *handler.WcV3Handler
	PODSyncHandler   *handler.PODSyncHandler
	PaymentHandler   *handler.PaymentHandler
	InvoiceHandler   *handler.InvoiceHandler
	EmailHandler     *handler.EmailHandler
	TaxHandler       *handler.TaxHandler
	AnalyticsHandler *handler.AnalyticsHandler
}

func Setup(p RouterParams) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Compress(5))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", p.Cfg.BaseURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); err == nil {
		fs := http.FileServer(http.Dir(uploadDir))
		r.Handle("/uploads/*", http.StripPrefix("/uploads/", fs))
	}

	// ================= API ROUTES =================
	r.Route("/api", func(api chi.Router) {
		// Publieke Auth
		api.Post("/auth/register", p.AuthHandler.Register)
		api.Post("/auth/login", p.AuthHandler.Login)
		api.Post("/auth/logout", p.AuthHandler.Logout)
		api.With(middleware.SessionMiddleware(p.UserRepo, "")).Get("/auth/me", p.AuthHandler.Me)

		// Publieke Catalogus
		api.Get("/products", p.ProductHandler.List)
		api.Get("/products/{slug}", p.ProductHandler.GetBySlug)
		api.Get("/categories", p.CategoryHandler.List)
		api.Get("/categories/{slug}", p.CategoryHandler.GetBySlug)
		api.Get("/widgets", p.WidgetHandler.ListActive)

		// Bestellingen & Betalingen
		api.With(middleware.OptionalSession(p.UserRepo)).Post("/orders", p.OrderHandler.CreateOrder)
		api.Get("/orders/{orderNumber}", p.OrderHandler.TrackOrder)
		api.Post("/orders/{orderNumber}/pay-mock", p.OrderHandler.MockPay)
		api.Get("/orders/{orderNumber}/invoice", p.InvoiceHandler.GetPublicInvoice)

		// Echte Betalingen & Betaalmethoden
		api.Get("/payment-methods", p.PaymentHandler.ListMethods)
		api.Post("/payments/initiate", p.PaymentHandler.InitiatePayment)

		// Btw-berekening & VIES B2B validatie
		api.Post("/taxes/calculate", p.TaxHandler.CalculateTax)
		api.Post("/taxes/validate-vies", p.TaxHandler.ValidateVIES)

		// Publieke Webhooks
		api.Post("/webhooks/mollie", p.PaymentHandler.MollieWebhook)
		api.Post("/webhooks/stripe", p.PaymentHandler.StripeWebhook)
		api.Post("/webhooks/gelato", p.PODHandler.GelatoWebhook)

		// Live Verzendkosten & Sync
		api.Post("/shipping-quotes", p.PODSyncHandler.GetLiveShipmentQuotes)

		// Coupons & Verzendopties
		api.Post("/coupons/validate", p.CouponHandler.ValidateCoupon)
		api.Get("/shipping-methods", p.ShippingHandler.ListActive)

		// Reviews
		api.Get("/products/{id}/reviews", p.ReviewHandler.ListByProduct)
		api.With(middleware.OptionalSession(p.UserRepo)).Post("/products/{id}/reviews", p.ReviewHandler.CreateReview)

		// Klantenportaal routes
		api.Group(func(cust chi.Router) {
			cust.Use(middleware.SessionMiddleware(p.UserRepo, "customer"))
			cust.Get("/user/orders", p.OrderHandler.ListUserOrders)
			cust.Get("/user/wishlist", p.WishlistHandler.List)
			cust.Post("/user/wishlist/{productId}", p.WishlistHandler.Add)
			cust.Delete("/user/wishlist/{productId}", p.WishlistHandler.Remove)
			cust.Get("/user/wishlist/{productId}/check", p.WishlistHandler.Check)
			cust.Get("/user/addresses", p.AddressHandler.List)
			cust.Post("/user/addresses", p.AddressHandler.Create)
			cust.Put("/user/addresses/{id}", p.AddressHandler.Update)
			cust.Delete("/user/addresses/{id}", p.AddressHandler.Delete)
		})

		// WooCommerce Admin routes
		api.Group(func(adm chi.Router) {
			adm.Use(middleware.SessionMiddleware(p.UserRepo, "admin"))

			adm.Get("/admin/stats", p.AdminHandler.GetStats)

			// Verkoop Analytics & CSV Export
			adm.Get("/admin/analytics/overview", p.AnalyticsHandler.GetOverview)
			adm.Get("/admin/analytics/export-csv", p.AnalyticsHandler.ExportCSV)

			// Klanten CRM
			adm.Get("/admin/customers", p.AdminHandler.ListCustomers)

			// Producten & Categorieën
			adm.Post("/admin/products", p.AdminHandler.CreateProduct)
			adm.Put("/admin/products/{id}", p.AdminHandler.UpdateProduct)
			adm.Delete("/admin/products/{id}", p.AdminHandler.DeleteProduct)
			adm.Post("/admin/products/variants", p.AdminHandler.UpsertVariant)
			adm.Delete("/admin/products/variants/{id}", p.AdminHandler.DeleteVariant)
			adm.Post("/admin/products/images", p.AdminHandler.AddProductImage)
			adm.Delete("/admin/products/images/{id}", p.AdminHandler.DeleteProductImage)

			adm.Get("/admin/categories", p.AdminHandler.ListCategories)
			adm.Post("/admin/categories", p.AdminHandler.CreateCategory)
			adm.Put("/admin/categories/{id}", p.AdminHandler.UpdateCategory)
			adm.Delete("/admin/categories/{id}", p.AdminHandler.DeleteCategory)

			// Widgets
			adm.Get("/admin/widgets", p.AdminHandler.ListWidgets)
			adm.Post("/admin/widgets", p.AdminHandler.UpsertWidget)
			adm.Delete("/admin/widgets/{id}", p.AdminHandler.DeleteWidget)

			// Bestellingen, Facturen & Terugbetalingen
			adm.Get("/admin/orders", p.AdminHandler.ListOrders)
			adm.Put("/admin/orders/{id}/status", p.AdminHandler.UpdateOrderStatus)
			adm.Post("/admin/orders/{id}/refund", p.PaymentHandler.RefundOrder)
			adm.Get("/admin/orders/{id}/invoice", p.InvoiceHandler.GetAdminInvoice)

			// Media Bibliotheek
			adm.Get("/admin/media", p.MediaHandler.List)
			adm.Post("/admin/upload", p.MediaHandler.Upload)
			adm.Delete("/admin/media/{filename}", p.MediaHandler.Delete)

			// Thema's & Store Settings
			adm.Get("/admin/themes", p.SettingsHandler.ListThemes)
			adm.Post("/admin/themes/activate", p.SettingsHandler.ActivateTheme)
			adm.Get("/admin/settings", p.SettingsHandler.GetSettings)
			adm.Put("/admin/settings", p.SettingsHandler.UpdateSettings)

			// Btw-Tarieven Matrix
			adm.Get("/admin/taxes", p.TaxHandler.ListRates)
			adm.Post("/admin/taxes", p.TaxHandler.UpsertRate)
			adm.Delete("/admin/taxes/{id}", p.TaxHandler.DeleteRate)

			// E-mail Test
			adm.Post("/admin/email/test", p.EmailHandler.SendTestEmail)

			// Gelato POD Sync
			adm.Post("/admin/pod/sync", p.PODSyncHandler.SyncCatalog)

			// Coupons & Verzendmethoden
			adm.Get("/admin/coupons", p.CouponHandler.ListCoupons)
			adm.Post("/admin/coupons", p.CouponHandler.CreateCoupon)
			adm.Put("/admin/coupons/{id}", p.CouponHandler.UpdateCoupon)
			adm.Delete("/admin/coupons/{id}", p.CouponHandler.DeleteCoupon)

			adm.Get("/admin/shipping-methods", p.ShippingHandler.ListAll)
			adm.Post("/admin/shipping-methods", p.ShippingHandler.Create)
			adm.Put("/admin/shipping-methods/{id}", p.ShippingHandler.Update)
			adm.Delete("/admin/shipping-methods/{id}", p.ShippingHandler.Delete)

			// Reviews
			adm.Get("/admin/reviews", p.ReviewHandler.ListAll)
			adm.Put("/admin/reviews/{id}/approve", p.ReviewHandler.ApproveReview)
			adm.Delete("/admin/reviews/{id}", p.ReviewHandler.DeleteReview)

			// WooCommerce REST API Sleutels
			adm.Get("/admin/api-keys", p.ApiKeyHandler.List)
			adm.Post("/admin/api-keys", p.ApiKeyHandler.Create)
			adm.Delete("/admin/api-keys/{id}", p.ApiKeyHandler.Delete)
		})
	})

	// ================= WOOCOMMERCE REST API v3 COMPATIBILITY LAYER =================
	r.Route("/wp-json/wc/v3", func(wc chi.Router) {
		wc.Use(middleware.WcAuthMiddleware(p.ApiKeyRepo, false))
		wc.Get("/system_status", p.WcV3Handler.SystemStatus)
		wc.Get("/products", p.WcV3Handler.ListProducts)
		wc.Get("/products/{id}", p.WcV3Handler.GetProduct)
		wc.Get("/orders", p.WcV3Handler.ListOrders)
		wc.Put("/orders/{id}", p.WcV3Handler.UpdateOrder)
	})

	// ================= SPA & DYNAMISCHE THEMA SEO INJECTIE =================
	r.Get("/assets/*", func(w http.ResponseWriter, req *http.Request) {
		distDir := p.ThemeManager.ResolveDistDir(req.Context())
		assetsDir := filepath.Join(distDir, "assets")
		http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))).ServeHTTP(w, req)
	})

	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		distDir := p.ThemeManager.ResolveDistDir(req.Context())
		indexPath := filepath.Join(distDir, "index.html")
		var rawHTML string

		if content, err := os.ReadFile(indexPath); err == nil {
			rawHTML = string(content)
		} else {
			rawHTML = `<!DOCTYPE html>
<html lang="nl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <!-- SEO_TAGS_INJECTION -->
</head>
<body>
    <div id="root"></div>
</body>
</html>`
		}

		path := strings.TrimPrefix(req.URL.Path, "/")
		meta := p.buildSEOMeta(req.Context(), path)
		injectedHTML := seo.InjectMeta(rawHTML, meta)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(injectedHTML))
	})

	return r
}

func (p *RouterParams) buildSEOMeta(ctx context.Context, path string) seo.PageMeta {
	defaultMeta := seo.DefaultMeta(p.Cfg.BaseURL)

	parts := strings.Split(path, "/")
	if len(parts) >= 2 && parts[0] == "product" {
		slug := parts[1]
		if prod, err := p.ProductRepo.GetBySlug(ctx, slug); err == nil && prod != nil {
			title := prod.Name
			if prod.SEOTitle != nil && *prod.SEOTitle != "" {
				title = *prod.SEOTitle
			}
			desc := prod.Description
			if prod.SEODescription != nil && *prod.SEODescription != "" {
				desc = *prod.SEODescription
			} else if prod.ShortDescription != nil && *prod.ShortDescription != "" {
				desc = *prod.ShortDescription
			}

			img := defaultMeta.ImageURL
			if prod.PrimaryImage != nil && *prod.PrimaryImage != "" {
				img = *prod.PrimaryImage
			}

			return seo.PageMeta{
				Title:        title,
				Description:  desc,
				CanonicalURL: fmt.Sprintf("%s/product/%s", p.Cfg.BaseURL, slug),
				ImageURL:     img,
				Type:         "product",
				Product: &seo.ProductMeta{
					ID:           prod.ID,
					Name:         prod.Name,
					Description:  desc,
					ImageURL:     img,
					Price:        fmt.Sprintf("%.2f", prod.BasePrice),
					Currency:     "EUR",
					Availability: "https://schema.org/InStock",
					Brand:        "Aesthetic Studios",
					URL:          fmt.Sprintf("%s/product/%s", p.Cfg.BaseURL, slug),
				},
			}
		}
	}

	if len(parts) >= 2 && parts[0] == "category" {
		slug := parts[1]
		if cat, err := p.CategoryRepo.GetBySlug(ctx, slug); err == nil && cat != nil {
			desc := fmt.Sprintf("Bekijk al onze %s producten in de webshop.", cat.Name)
			if cat.Description != nil && *cat.Description != "" {
				desc = *cat.Description
			}
			return seo.PageMeta{
				Title:        fmt.Sprintf("%s Collectie | Shop Online", cat.Name),
				Description:  desc,
				CanonicalURL: fmt.Sprintf("%s/category/%s", p.Cfg.BaseURL, slug),
				ImageURL:     defaultMeta.ImageURL,
				Type:         "website",
			}
		}
	}

	return defaultMeta
}