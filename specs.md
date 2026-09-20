# 📐 Volledige Systeemspecificaties & Architectuur (SPECS.MD)

Dit document is het **definitieve referentiekader** van het e-commerce platform. Het beschrijft alle geïmplementeerde datamodellen, tabellen, backend-logica, API-endpoints, frontend-functionaliteiten en no-code instellingen. 

---

## 1. Technologie-Stack & Systeemarchitectuur

* **Backend:** Go (Golang 1.23+)
  * **HTTP Router & Middleware:** `github.com/go-chi/chi/v5` met gzip-compressie, real IP, CORS en recoverer.
  * **Database Driver & Pooling:** `github.com/jackc/pgx/v5/pgxpool` (25 max connections, pool lifetime management).
  * **Wachtwoord-Hashing:** `golang.org/x/crypto/bcrypt` (standaard cost 10).
  * **Opslag (Media Storage):** Cloudflare R2 (S3-compatibel via AWS SDK Go v2) met automatische lokale fallback naar `./uploads`.
  * **Fulfillment & Print-on-Demand:** Gelato API v2 integratie met live shipment quotes (`/v2/shipment-quotes`), catalogus synchronisatie en status-webhooks.
  * **E-mail Engine:** Volledig werkende SMTP mailer met TLS (`net/smtp`) en Resend API ondersteuning.
  * **Facturatie Engine:** Printklare A4 HTML/PDF generator met 21% Btw-uitsplitsing en bedrijfsgegevens.
* **Database:** PostgreSQL 17
  * **Primaire Sleutels:** 100% time-ordered **RFC 9562 UUIDv7** voor alle tabellen via de ingebouwde pure SQL functie `uuidv7()`.
* **Frontend:** SolidJS SPA (Single Page Application)
  * **Build Tool:** Vite 5 + TypeScript.
  * **Routing:** `@solidjs/router`.
  * **Styling:** Semantische SCSS met CSS custom properties (`data-theme="dark"` / `"light"`).
  * **Meertaligheid (i18n):** Reactieve vertalingen voor Nederlands (NL) en Engels (EN) met taalkiezer popover.
  * **Multi-Valuta:** Real-time conversie en formattering voor EUR (€), USD ($) en GBP (£).
* **Server-Side SEO & Meta Injectie:**
  * Go serveert de frontend en vervangt `<!-- SEO_TAGS_INJECTION -->` in `index.html` dynamisch door pagina-specifieke `<title>`, `<meta name="description">`, OpenGraph tags en Schema.org Product JSON-LD (Rich Snippets voor Google).
* **WooCommerce REST API v3 Compatibiliteitslaag:**
  * Implementeert de officiële `/wp-json/wc/v3/*` endpoints met Consumer Key (`ck_...`) en Consumer Secret (`cs_...`) authenticatie.

---

## 2. Inloggegevens & Testaccounts

Bij de eerste serverstart zaait Go automatisch twee accounts in PostgreSQL:

| Account Type | E-mailadres | Wachtwoord | Doel / Rechten |
| :--- | :--- | :--- | :--- |
| **WooCommerce Beheerder** | `admin@aesthetic.be` | `admin123!` | Volledige toegang tot het Admin Dashboard via `/admin` (schild-icoon). |
| **Klantaccount** | `gijs@test.be` | `test123!` | Toegang tot bestelgeschiedenis, favorieten, adresboek en facturen via `/account`. |

---

## 3. Volledig PostgreSQL Databaseschema (14 Tabellen)

Alle tabellen maken gebruik van `id UUID PRIMARY KEY DEFAULT uuidv7()`.

1. **`users`**: Klanten en beheerders.
   * `id`, `email` (UNIQUE), `password_hash`, `full_name`, `role` ('customer' / 'admin'), `created_at`, `updated_at`.
2. **`sessions`**: Server-side sessiebeheer.
   * `id`, `user_id` (FK `users.id` ON DELETE CASCADE), `token` (UNIQUE hex string van 64 tekens), `expires_at`, `created_at`.
3. **`categories`**: Productcategorieën.
   * `id`, `name`, `slug` (UNIQUE), `description`, `image_url`, `sort_order`, `is_active`, `created_at`.
4. **`products`**: Hoofdproducten.
   * `id`, `category_id` (FK `categories.id` ON DELETE SET NULL), `name`, `slug` (UNIQUE), `short_description`, `description`, `base_price` (NUMERIC 10,2), `featured` (BOOLEAN), `is_active` (BOOLEAN), `product_type` ('variable' / 'simple' / 'digital'), `seo_title`, `seo_description`, `created_at`, `updated_at`.
5. **`product_variants`**: Maten, kleuren en POD-koppelingen per product.
   * `id`, `product_id` (FK `products.id` ON DELETE CASCADE), `sku` (UNIQUE), `title`, `size`, `color`, `color_hex`, `price_override` (NUMERIC 10,2), `stock_quantity` (INT), `image_url`, `pod_provider` ('gelato' / NULL), `pod_variant_id`, `print_file_url`, `is_active`, `created_at`, `updated_at`.
6. **`product_images`**: Meerdere galerijfoto's per product.
   * `id`, `product_id` (FK `products.id` ON DELETE CASCADE), `variant_id` (FK `product_variants.id` ON DELETE SET NULL), `url`, `alt_text`, `sort_order`, `is_primary` (BOOLEAN), `created_at`.
7. **`product_reviews`**: Klantervaringen en sterrenbeoordelingen.
   * `id`, `product_id` (FK `products.id` ON DELETE CASCADE), `user_id` (FK `users.id` ON DELETE SET NULL), `author_name`, `rating` (1 t/m 5), `comment`, `is_approved` (BOOLEAN), `created_at`.
8. **`product_downloads`**: Digitale bestanden voor downloadbare producten.
   * `id`, `product_id` (FK `products.id` ON DELETE CASCADE), `name`, `file_url`, `download_limit`, `created_at`.
9. **`orders`**: Klantbestellingen.
   * `id`, `order_number` (UNIQUE, bijv. `ORD-20260920-XXXX`), `user_id` (FK `users.id` ON DELETE SET NULL), `guest_email`, `status` ('pending', 'processing', 'shipped', 'delivered', 'cancelled'), `subtotal`, `shipping_cost`, `total_amount`, `shipping_address` (JSONB), `billing_address` (JSONB), `tracking_code`, `carrier`, `payment_status` ('unpaid', 'paid', 'failed', 'refunded'), `payment_provider` ('mock', 'mollie', 'stripe'), `created_at`, `updated_at`.
10. **`order_items`**: Bestelregels binnen een order.
    * `id`, `order_id` (FK `orders.id` ON DELETE CASCADE), `product_id` (FK `products.id`), `variant_id` (FK `product_variants.id`), `product_name`, `variant_title`, `sku`, `unit_price`, `quantity`, `total_price`.
11. **`order_timeline`**: Live status-historiek per bestelling.
    * `id`, `order_id` (FK `orders.id` ON DELETE CASCADE), `status`, `message`, `created_at`.
12. **`order_refunds`**: Geregistreerde terugbetalingen.
    * `id`, `order_id` (FK `orders.id` ON DELETE CASCADE), `amount`, `reason`, `created_at`.
13. **`coupons`**: Kortingscodes.
    * `id`, `code` (UNIQUE), `discount_type` ('percent' / 'fixed'), `discount_value`, `min_spend`, `max_uses`, `uses_count`, `expires_at`, `is_active`, `created_at`.
14. **`shipping_methods`**: Verzendopties en tarieven.
    * `id`, `title`, `description`, `cost`, `free_threshold`, `is_active`, `sort_order`, `created_at`.
15. **`tax_rates`**: Btw-tarievenmatrix per land (EU OSS).
    * `id`, `country_code` (VARCHAR 2), `state_code`, `rate` (NUMERIC 5,2), `name`, `is_compound`, `priority`, `created_at`.
16. **`widgets`**: Homepage widgets.
    * `id`, `type` ('hero', 'featured_products', 'categories_grid', 'banner_promo', 'trust_badges', 'newsletter', 'category_spotlight', 'best_sellers', 'recent_products'), `title`, `subtitle`, `config` (JSONB), `sort_order`, `is_active`, `created_at`, `updated_at`.
17. **`wishlists`**: Klantenverlanglijst.
    * `id`, `user_id` (FK `users.id` ON DELETE CASCADE), `product_id` (FK `products.id` ON DELETE CASCADE), `created_at`. UNIQUE(`user_id`, `product_id`).
18. **`addresses`**: Opgeslagen klantadressen.
    * `id`, `user_id` (FK `users.id` ON DELETE CASCADE), `label`, `full_name`, `street`, `house_number`, `bus`, `city`, `postal_code`, `country`, `is_default`, `created_at`, `updated_at`.
19. **`store_settings`**: Sleutel-waarde configuratietabel (No-Code).
    * `key` (TEXT PRIMARY KEY), `value` (JSONB), `updated_at`.
20. **`api_keys`**: WooCommerce v3 REST API sleutels.
    * `id`, `description`, `consumer_key` (UNIQUE), `consumer_secret_hash`, `permissions` ('read', 'write', 'read_write'), `created_at`.
21. **`webhooks`**: Uitgaande webhooks voor Zapier/Make/Boekhouding.
    * `id`, `name`, `target_url`, `secret`, `events` (TEXT[]), `is_active`, `created_at`.

---

## 4. Complete REST API Specificatie (Alle Endpoints)

### A. Publieke Storefront & Catalogus
* `GET /api/products` - Productoverzicht met filtering (`category`, `search`, `min_price`, `max_price`, `size`, `color`, `sort`, `page`, `limit`).
* `GET /api/products/:slug` - Productdetails inclusief varianten, POD gegevens en fotogalerij.
* `GET /api/categories` - Overzicht van actieve categorieën.
* `GET /api/categories/:slug` - Specifieke categorie data.
* `GET /api/widgets` - Actieve homepage widgets.
* `GET /api/products/:id/reviews` - Reviews en gemiddelde score van een product.
* `POST /api/products/:id/reviews` - Review plaatsen.

### B. Winkelmand, Verzendkosten & Betalingen
* `POST /api/shipping-quotes` - Berekent beschikbare bezorgopties + realtime Gelato POD tarieven op basis van adres en artikelen.
* `POST /api/coupons/validate` - Valideert een kortingscode en berekent de korting.
* `POST /api/taxes/calculate` - Berekent Btw volgens EU OSS en past automatische B2B Btw-verlegging (0%) toe.
* `POST /api/taxes/validate-vies` - Live validatie van Europees Btw-nummer via officiële EU VIES API.
* `GET /api/payment-methods` - Geeft actieve betaalmethoden terug (Mock, Mollie iDEAL/Bancontact, Stripe).
* `POST /api/orders` - Plaatst de bestelling atomisch (inclusief voorraadafboeking en ordernummer generatie).
* `POST /api/payments/initiate` - Start betaling bij de gekozen provider en geeft bank-redirect URL terug.
* `POST /api/orders/:orderNumber/pay-mock` - Simuleert direct een testbetaling.

### C. Live Bestelling Volgen & Facturen
* `GET /api/orders/:orderNumber` - Live tracking en tijdlijnstatus.
* `GET /api/orders/:orderNumber/invoice` - Officiële A4 HTML/PDF factuur.

### D. Klantenportaal (Sessie vereist)
* `POST /api/auth/register` - Registreren.
* `POST /api/auth/login` - Inloggen (start 30-dagen `HttpOnly` sessiecookie).
* `POST /api/auth/logout` - Uitloggen (wist sessie direct uit database en browser).
* `GET /api/auth/me` - Ingelogde gebruiker ophalen.
* `GET /api/user/orders` - Bestelhistoriek van klant.
* `GET /api/user/wishlist` - Favorietenlijst van klant.
* `POST /api/user/wishlist/:productId` - Toevoegen aan favorieten.
* `DELETE /api/user/wishlist/:productId` - Verwijderen uit favorieten.
* `GET /api/user/addresses` - Opgeslagen adressen.
* `POST /api/user/addresses` - Nieuw adres toevoegen.
* `PUT /api/user/addresses/:id` - Adres bijwerken.
* `DELETE /api/user/addresses/:id` - Adres verwijderen.

### E. Inkomende Webhooks
* `POST /api/webhooks/mollie` - Verwerkt live iDEAL / Bancontact statussen.
* `POST /api/webhooks/stripe` - Verwerkt live Stripe Checkout statussen.
* `POST /api/webhooks/gelato` - Ontvangt trackingcodes zodra shirts geprint en verzonden zijn door Gelato.

### F. WooCommerce REST API v3 (Compatibiliteitslaag)
*Beveiligd via Consumer Key (`ck_...`) en Consumer Secret (`cs_...`)*
* `GET /wp-json/wc/v3/system_status` - Systeemstatus controle voor externe tools.
* `GET /wp-json/wc/v3/products` - Producten exporteren in WooCommerce v3 JSON-formaat.
* `GET /wp-json/wc/v3/products/:id` - Enkel product in WooCommerce formaat.
* `GET /wp-json/wc/v3/orders` - Bestellingen exporteren in WooCommerce v3 formaat.
* `PUT /wp-json/wc/v3/orders/:id` - Orderstatus en trackingnummer bijwerken.

### G. Admin Beheerpanel (`/api/admin/*`, Admin Sessie vereist)
* `GET /api/admin/stats` - WooCommerce KPI's (Omzet, Bestellingen, Klanten, Lage voorraad).
* `GET /api/admin/analytics/overview` - Tijdreeksdata omzet, AOV, top sellers en categorie-omzet.
* `GET /api/admin/analytics/export-csv` - Downloadbaar CSV verkooprapport voor boekhouding.
* `GET /api/admin/customers` - Klanten CRM overzicht met Lifetime Value (LTV).
* `GET /api/admin/media` - Alle bestanden in de media bibliotheek.
* `POST /api/admin/upload` - Afbeelding uploaden (max 10MB).
* `DELETE /api/admin/media/:filename` - Afbeelding verwijderen.
* `POST /api/admin/products` - Product aanmaken.
* `PUT /api/admin/products/:id` - Product bewerken.
* `DELETE /api/admin/products/:id` - Product verwijderen.
* `POST /api/admin/products/variants` - Variant opslaan (inclusief Gelato POD ID en drukbestand URL).
* `DELETE /api/admin/products/variants/:id` - Variant verwijderen.
* `POST /api/admin/products/images` - Afbeelding koppelen aan product/variant.
* `DELETE /api/admin/products/images/:id` - Afbeelding ontkoppelen.
* `GET /api/admin/orders` - Bestellingenoverzicht met statusfilters.
* `PUT /api/admin/orders/:id/status` - Bestelstatus en trackingcode wijzigen.
* `POST /api/admin/orders/:id/refund` - Bestelling terugbetalen.
* `GET /api/admin/orders/:id/invoice` - Beheerder factuurweergave.
* `GET /api/admin/widgets` - Homepage widgets ophalen.
* `POST /api/admin/widgets` - Widget aanmaken of bewerken (titel, foto, knoptekst, link).
* `DELETE /api/admin/widgets/:id` - Widget verwijderen.
* `GET /api/admin/themes` - Geïnstalleerde thema's in `themes/*` scannen.
* `POST /api/admin/themes/activate` - Thema met 1 klik activeren.
* `GET /api/admin/settings` - Alle winkelinstellingen ophalen.
* `PUT /api/admin/settings` - Winkelinstellingen opslaan.
* `POST /api/admin/pod/sync` - Direct de Gelato catalogus synchroniseren.
* `POST /api/admin/email/test` - Test e-mail versturen via geconfigureerde SMTP/Resend.
* `GET /api/admin/coupons` - Kortingscodes beheren.
* `POST /api/admin/coupons` - Kortingscode aanmaken.
* `DELETE /api/admin/coupons/:id` - Kortingscode verwijderen.
* `GET /api/admin/shipping-methods` - Verzendmethoden beheren.
* `POST /api/admin/shipping-methods` - Verzendmethode toevoegen.
* `DELETE /api/admin/shipping-methods/:id` - Verzendmethode verwijderen.
* `GET /api/admin/taxes` - Btw-tarieven beheren.
* `POST /api/admin/taxes` - Btw-tarief toevoegen.
* `DELETE /api/admin/taxes/:id` - Btw-tarief verwijderen.
* `GET /api/admin/api-keys` - WooCommerce API-sleutels beheren.
* `POST /api/admin/api-keys` - Nieuwe Consumer Key & Secret genereren.
* `DELETE /api/admin/api-keys/:id` - API-sleutel intrekken.

---

## 5. UI & Thema Functionaliteiten

1. **Header (56px hoog):**
   * Transparant op homepagina wanneer bovenaan (`scrollY <= 30`), vloeiend overgaand in frosted glass bij scrollen.
   * Geen harde kadertjes om knoppen voor een schone, minimalistische uitstraling.
   * **Taal- & Valuta Dropdown:** Schakelen tussen NL 🇳🇱 / EN 🇬🇧 en EUR (€), USD ($), GBP (£).
   * Live winkelmandje met badge en realtime prijs in geselecteerde valuta.
   * Responsive: menu-items schuiven automatisch naar het zijmenu onder `1150px`.
2. **Zijmenu (Aside Drawer):**
   * Schuift in vanaf **rechts** met `border-radius: 0` (strak tegen schermrand).
   * Account-tegel over volle breedte direct onder het logo.
   * Knoppen over 100% breedte met slide-in hover-effect en accentlijn.
   * Uitlogknop vast gepind aan de onderkant.
3. **Productkaarten:**
   * Volledige kaart (afbeelding, titel, prijs) is klikbaar en leidt naar de PDP.
   * Liken via het hartje stopt bubbling (`e.stopPropagation()`).
4. **Product Detail Pagina (PDP):**
   * Bovenbalk met *← Terug naar overzicht* knop en broodkruimelpad.
   * Variantenkiezer met dynamische prijs- en voorraadindicatie.
   * Grote gecentreerde favorietknop met rode vulling bij opslaan.
5. **Checkout & Betaling:**
   * Live verzendopties selector (Winkelmethoden + Gelato POD quotes).
   * Kortingscode invoerveld met directe prijsverrekening.
   * Zakelijke B2B Btw-verlegging met live EU VIES controle.
6. **Order Tracking:**
   * Live tijdlijn met status-events en directe link naar de A4 PDF-factuur.