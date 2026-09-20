package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webshop/internal/auth"
	"webshop/internal/config"
	"webshop/internal/database"
	"webshop/internal/email"
	"webshop/internal/handler"
	"webshop/internal/invoice"
	"webshop/internal/payment"
	"webshop/internal/pod"
	"webshop/internal/repository"
	"webshop/internal/router"
	"webshop/internal/tax"
	"webshop/internal/theme"
	"webshop/internal/webhook"
	"webshop/internal/storage"
)

func main() {
	log.Println("Shop server wordt geïnitialiseerd...")
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Database verbinding opzetten
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database initialisatie mislukt: %v", err)
	}
	defer db.Close()

	// 2. Migraties uitvoeren (Init + WooCommerce Core + Enterprise)
	for _, migrationPath := range []string{
		"migrations/000001_init.up.sql",
		"migrations/000002_woocommerce_core.up.sql",
		"migrations/000003_enterprise_woocommerce.up.sql",
	} {
		if _, err := os.Stat(migrationPath); err == nil {
			if err := db.RunMigrations(ctx, migrationPath); err != nil {
				log.Printf("Waarschuwing bij migratie (%s): %v", migrationPath, err)
			}
		}
	}

	// 3. Storage initialiseren
	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Storage initialisatie mislukt: %v", err)
	}

	// 4. Repositories
	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	widgetRepo := repository.NewWidgetRepo(db)
	wishlistRepo := repository.NewWishlistRepo(db)
	addressRepo := repository.NewAddressRepo(db)
	settingsRepo := repository.NewSettingsRepo(db)
	couponRepo := repository.NewCouponRepo(db)
	shippingRepo := repository.NewShippingRepo(db)
	reviewRepo := repository.NewReviewRepo(db)
	apiKeyRepo := repository.NewApiKeyRepo(db)
	taxRepo := repository.NewTaxRepo(db, settingsRepo)
	analyticsRepo := repository.NewAnalyticsRepo(db)

	// 5. Thema manager, POD, Tax & Payment Services
	themeManager := theme.NewThemeManager("./themes", settingsRepo)
	gelatoClient := pod.NewGelatoClient(settingsRepo, productRepo, categoryRepo, db)
	webhookDispatcher := webhook.NewDispatcher(db)
	paymentService := payment.NewService(cfg, settingsRepo, orderRepo)
	mailer := email.NewMailer(settingsRepo)
	invoiceGen := invoice.NewInvoiceGenerator(settingsRepo)
	taxService := tax.NewService(taxRepo, settingsRepo)

	// Demo data zaaien
	seedDemoData(ctx, db, userRepo, categoryRepo, productRepo, widgetRepo, taxRepo)

	// 6. Handlers
	authHandler := handler.NewAuthHandler(userRepo, cfg)
	productHandler := handler.NewProductHandler(productRepo)
	categoryHandler := handler.NewCategoryHandler(categoryRepo)
	widgetHandler := handler.NewWidgetHandler(widgetRepo)
	orderHandler := handler.NewOrderHandler(orderRepo, gelatoClient, webhookDispatcher)
	wishlistHandler := handler.NewWishlistHandler(wishlistRepo)
	addressHandler := handler.NewAddressHandler(addressRepo)
	mediaHandler := handler.NewMediaHandler(store, cfg)
	adminHandler := handler.NewAdminHandler(db, productRepo, categoryRepo, orderRepo, widgetRepo, userRepo)
	settingsHandler := handler.NewSettingsHandler(settingsRepo, themeManager)
	couponHandler := handler.NewCouponHandler(couponRepo)
	shippingHandler := handler.NewShippingHandler(shippingRepo)
	reviewHandler := handler.NewReviewHandler(reviewRepo)
	podHandler := handler.NewPODHandler(orderRepo)
	apiKeyHandler := handler.NewApiKeyHandler(apiKeyRepo)
	wcV3Handler := handler.NewWcV3Handler(cfg, productRepo, orderRepo)
	podSyncHandler := handler.NewPODSyncHandler(gelatoClient, shippingRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService, orderRepo, settingsRepo, gelatoClient, webhookDispatcher)
	invoiceHandler := handler.NewInvoiceHandler(invoiceGen, orderRepo)
	emailHandler := handler.NewEmailHandler(mailer)
	taxHandler := handler.NewTaxHandler(taxService, taxRepo)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsRepo)

	// 7. Router samenstellen
	r := router.Setup(router.RouterParams{
		Cfg:              cfg,
		UserRepo:         userRepo,
		ProductRepo:      productRepo,
		CategoryRepo:     categoryRepo,
		SettingsRepo:     settingsRepo,
		ApiKeyRepo:       apiKeyRepo,
		ThemeManager:     themeManager,
		AuthHandler:      authHandler,
		ProductHandler:   productHandler,
		CategoryHandler:  categoryHandler,
		WidgetHandler:    widgetHandler,
		OrderHandler:     orderHandler,
		WishlistHandler:  wishlistHandler,
		AddressHandler:   addressHandler,
		MediaHandler:     mediaHandler,
		AdminHandler:     adminHandler,
		SettingsHandler:  settingsHandler,
		CouponHandler:    couponHandler,
		ShippingHandler:  shippingHandler,
		ReviewHandler:    reviewHandler,
		PODHandler:       podHandler,
		ApiKeyHandler:    apiKeyHandler,
		WcV3Handler:      wcV3Handler,
		PODSyncHandler:   podSyncHandler,
		PaymentHandler:   paymentHandler,
		InvoiceHandler:   invoiceHandler,
		EmailHandler:     emailHandler,
		TaxHandler:       taxHandler,
		AnalyticsHandler: analyticsHandler,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server draait live op: %s (Port %s)", cfg.BaseURL, cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server fout: %v", err)
		}
	}()

	<-stopChan
	log.Println("Server wordt netjes afgesloten...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server geforceerd gestopt: %v", err)
	}

	log.Println("Server succesvol gestopt.")
}

func seedDemoData(ctx context.Context, db *database.DB, userRepo *repository.UserRepo, catRepo *repository.CategoryRepo, prodRepo *repository.ProductRepo, widgetRepo *repository.WidgetRepo, taxRepo *repository.TaxRepo) {
	adminEmail := "admin@aesthetic.be"
	existingAdmin, _ := userRepo.GetByEmail(ctx, adminEmail)
	if existingAdmin == nil {
		hash, _ := auth.HashPassword("admin123!")
		_, _ = userRepo.CreateUser(ctx, adminEmail, hash, "Aesthetic Admin", "admin")
		log.Println("Demo admin aangemaakt: admin@aesthetic.be (wachtwoord: admin123!)")
	}

	customerEmail := "gijs@test.be"
	existingCust, _ := userRepo.GetByEmail(ctx, customerEmail)
	if existingCust == nil {
		custHash, _ := auth.HashPassword("test123!")
		_, _ = userRepo.CreateUser(ctx, customerEmail, custHash, "Gijs Klant", "customer")
		log.Println("Demo klant aangemaakt: gijs@test.be (wachtwoord: test123!)")
	}

	cats, _ := catRepo.ListAll(ctx)
	if len(cats) == 0 {
		descOvr := "Ruimvallende heavy-cotton t-shirts met minimalistisch silhouet"
		c1, err := catRepo.Create(ctx, "Oversized", "oversized", &descOvr, nil, 1)
		if err != nil || c1 == nil {
			log.Printf("Kon demo categorie 1 niet aanmaken: %v", err)
			return
		}

		descHoo := "Premium zware hoodies met subtiele details"
		c2, err := catRepo.Create(ctx, "Hoodies & Basics", "hoodies", &descHoo, nil, 2)
		if err != nil || c2 == nil {
			log.Printf("Kon demo categorie 2 niet aanmaken: %v", err)
			return
		}

		p1Short := "240 GSM gekamd biologisch katoen"
		p1Desc := "Het ultieme oversized t-shirt. Gemaakt van 240 GSM zwaar gekamd biokatoen met een stevige boord en perfecte boxy fit."
		p1Img := "https://images.unsplash.com/photo-1521572267360-ee0c2909d518?w=800&auto=format&fit=crop&q=80"
		p1, err := prodRepo.Create(ctx, &c1.ID, "Pitch Black Oversized Tee", "pitch-black-oversized-tee", &p1Short, p1Desc, 39.95, true, true, nil, nil)
		if err == nil {
			_ = prodRepo.AddImage(ctx, p1.ID, nil, p1Img, nil, 1, true)
			for _, size := range []string{"S", "M", "L", "XL"} {
				_ = prodRepo.UpsertVariant(ctx, repository.ProductVariant{
					ProductID:     p1.ID,
					SKU:           "TSH-BLK-" + size,
					Title:         size + " / Pitch Black",
					Size:          &size,
					StockQuantity: 25,
					IsActive:      true,
				})
			}
		}

		p2Short := "Vintage acid wash behandeling"
		p2Desc := "Unieke vintage uitstraling door onze zorgvuldige acid-wash wassing. Zacht gevoel met een vintage look."
		p2Img := "https://images.unsplash.com/photo-1583743814966-8936f5b7be1a?w=800&auto=format&fit=crop&q=80"
		p2, err := prodRepo.Create(ctx, &c1.ID, "Vintage Washed Acid Grey Tee", "vintage-washed-acid-grey-tee", &p2Short, p2Desc, 44.95, true, true, nil, nil)
		if err == nil {
			_ = prodRepo.AddImage(ctx, p2.ID, nil, p2Img, nil, 1, true)
			for _, size := range []string{"M", "L", "XL"} {
				_ = prodRepo.UpsertVariant(ctx, repository.ProductVariant{
					ProductID:     p2.ID,
					SKU:           "TSH-GRY-" + size,
					Title:         size + " / Acid Grey",
					Size:          &size,
					StockQuantity: 20,
					IsActive:      true,
				})
			}
		}

		p3Short := "450 GSM French Terry katoen"
		p3Desc := "Zware kwaliteit hoodie zonder trekkoorden voor een modern en strak minimalistisch uiterlijk."
		p3Img := "https://images.unsplash.com/photo-1556905055-8f358a7a47b2?w=800&auto=format&fit=crop&q=80"
		p3, err := prodRepo.Create(ctx, &c2.ID, "Heavy Boxy Slate Hoodie", "heavy-boxy-slate-hoodie", &p3Short, p3Desc, 79.95, true, true, nil, nil)
		if err == nil {
			_ = prodRepo.AddImage(ctx, p3.ID, nil, p3Img, nil, 1, true)
			for _, size := range []string{"S", "M", "L"} {
				_ = prodRepo.UpsertVariant(ctx, repository.ProductVariant{
					ProductID:     p3.ID,
					SKU:           "HOD-SLT-" + size,
					Title:         size + " / Slate",
					Size:          &size,
					StockQuantity: 15,
					IsActive:      true,
				})
			}
		}

		log.Println("Demo categorieën, t-shirts en varianten succesvol geseed!")
	}

	ws, _ := widgetRepo.ListAll(ctx)
	if len(ws) == 0 {
		tHero := "Hero Banner"
		_ = widgetRepo.Upsert(ctx, nil, "hero", &tHero, nil, []byte("{}"), 1, true)
		tTrust := "Trust Badges"
		_ = widgetRepo.Upsert(ctx, nil, "trust_badges", &tTrust, nil, []byte("{}"), 2, true)
		tFeatured := "Populaire Items"
		_ = widgetRepo.Upsert(ctx, nil, "featured_products", &tFeatured, nil, []byte("{}"), 3, true)
		tNL := "Nieuwsbrief"
		_ = widgetRepo.Upsert(ctx, nil, "newsletter", &tNL, nil, []byte("{}"), 4, true)
		log.Println("Demo homepage widgets succesvol geseed!")
	}

	// Btw-tarieven zaaien indien leeg
	if taxRepo != nil {
		rates, _ := taxRepo.ListAll(ctx)
		if len(rates) == 0 {
			_ = taxRepo.UpsertRate(ctx, "BE", "Belgische Btw (21%)", 21.0)
			_ = taxRepo.UpsertRate(ctx, "NL", "Nederlandse Btw (21%)", 21.0)
			_ = taxRepo.UpsertRate(ctx, "DE", "Duitse MwSt (19%)", 19.0)
			_ = taxRepo.UpsertRate(ctx, "FR", "Franse TVA (20%)", 20.0)
			log.Println("Standaard EU Btw-tarieven succesvol geseed!")
		}
	}
}