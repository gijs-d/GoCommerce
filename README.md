# ⚡ Modern Go + SolidJS WooCommerce Engine (Headless & No-Code)

An ultra-fast, scalable, and modern e-commerce webshop engine written in **Go (Golang 1.23+)**, **SolidJS SPA**, and **PostgreSQL 17**. 

Designed as a fully-fledged, lightning-fast alternative to WooCommerce and Shopify. You compile the backend once, after which you manage the entire store via the No-Code Admin UI.

---

## ✨ Enterprise Features

* 🌍 **International & B2B:** Built-in i18n (NL/EN), Multi-Currency (EUR, USD, GBP), and a dynamic EU OSS Tax matrix with **live VIES validation** for automatic 0% reverse charge on B2B orders.
* 🖨️ **Print-on-Demand (Gelato):** Full Gelato API v2 integration. Sync your catalog with 1 click, show **live shipping rates** during checkout, and automatically route orders to the printing facility.
* 🔌 **WooCommerce REST API v3:** Generate Consumer Keys (`ck_...`) in the admin. External tools like Sendcloud, Printful, Klaviyo, or Exact Online can connect directly as if it were a WordPress site!
* 📊 **Analytics & CRM:** Built-in SVG revenue charts, CSV export for accounting, and a Customer CRM with Lifetime Value (LTV) tracking.
* 📧 **Invoicing & Email:** Automatically generates print-ready A4 HTML/PDF invoices with tax breakdown. Built-in SMTP & Resend mailer for order confirmations.
* 🎨 **Theme System:** Manage multiple frontends via the `themes/` folder. Switch themes with 1 click in the admin. Go automatically injects server-side OpenGraph and Schema.org JSON-LD for perfect SEO.

---

## 🚀 Quick Start

### 1. Requirements
* **Go 1.23+** installed
* **Node.js 20+** & npm
* **Docker** & Docker Compose

### 2. Start Database (Postgres 17 with UUIDv7)
```bash
docker compose up -d postgres
```
*Runs on host port `5433` (internal `5432`) to avoid conflicts with any local PostgreSQL installations.*

### 3. Fetch Dependencies & Start Backend
```bash
# Fetch Go dependencies
go mod tidy

# Start server (automatically runs migrations and demo-seeding)
go run cmd/server/main.go
```
The Go server will now start at: **`http://localhost:8080`**.

### 4. Start Frontend (During Development with Hot-Reload)
Open a second terminal:
```bash
cd frontend
npm install
npm run dev
```
Vite starts at `http://localhost:3000` and automatically proxies all API calls to the Go backend at `:8080`.

For a production build:
```bash
cd frontend && npm run build && cd ..
```

---

## 🔑 Test & Admin Accounts

As soon as the backend starts for the first time, two accounts are automatically seeded in PostgreSQL:

| Account Type | Email Address | Password | Rights & Access |
| :--- | :--- | :--- | :--- |
| **WooCommerce Admin** | `admin@aesthetic.be` | `admin123!` | Full management via the shield icon (`/admin`) |
| **Customer Account** | `gijs@test.be` | `test123!` | Order history, invoices, wishlist, address management (`/account`) |

---

## 🎨 The Theme System (`themes/`)

Each theme is a standalone frontend project located in the `themes/` folder:
```text
themes/
├── aesthetic/          # The active theme
│   ├── theme.json      # Metadata (name, author, version, preview)
│   ├── package.json
│   ├── src/            # Unbuilt TypeScript / SolidJS source code
│   └── dist/           # Compiled assets served by Go
```

### Creating a new theme:
1. Copy an existing folder: `cp -r themes/aesthetic themes/my-new-style`
2. Change the name in `themes/my-new-style/theme.json`.
3. Tweak the styling/layout and run `npm run build`.
4. Go to the admin panel (`/admin > Themes`) and click **"Activate"**. Go switches over instantly!

---

## 🛠️ No-Code Admin Dashboard Features (`/admin`)

1. **Products & Variants:**
   * **Simple vs. Variable:** Fixed price/stock or attribute matrix (sizes, colors, custom stock).
   * **Print-on-Demand (POD):** Link items directly to Gelato or another printer using `pod_variant_id` and a print file URL.
   * **Gallery:** Add multiple images and link them to specific color variants.
2. **Orders & Live Timeline:**
   * Manage statuses (`pending`, `processing`, `shipped`, `delivered`, `cancelled`).
   * Add Track & Trace codes or let couriers/printers link them automatically.
3. **Discount Codes (Coupons):**
   * Fixed discounts or percentages with minimum spend and expiration dates.
4. **Store Settings:**
   * **Payments:** Easily switch between **Test/Mock Payer** (for local testing) and live payments via **Mollie** (iDEAL, Bancontact) or **Stripe** (Credit Card, Apple Pay).
   * **Shipping:** Set standard rates and free shipping thresholds.
   * **Taxes & Invoicing:** Tax percentages per country and company/VAT number for invoices.
   * **Email:** Sender name and SMTP/Resend settings.
5. **No-Code Homepage Widgets:**
   * Live edit texts, photos, CTA buttons, and links for the Hero Banner, Trust Badges, Featured Products, and Newsletter.

---

## 📦 Standalone Docker Production Deployment

Want to run the entire store (Go server + compiled frontend) as **a single container of ~30MB**?
```bash
docker compose up --build -d
```
Both containers (PostgreSQL 17 and the Go webshop engine) will start automatically and restart upon crashes.