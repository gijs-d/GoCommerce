# ⚡ Modern Go + SolidJS WooCommerce Engine (Headless & No-Code)

Een ultra-snelle, schaalbare en moderne e-commerce webshop engine geschreven in **Go (Golang 1.23+)**, **SolidJS SPA** en **PostgreSQL 17**. 

Ontworpen als een volwaardig, bliksemsnel alternatief voor WooCommerce en Shopify. Je compileert de backend één keer, waarna je de volledige winkel beheert via de No-Code Admin UI.

---

## ✨ Enterprise Functionaliteiten

* 🌍 **Internationaal & B2B:** Ingebouwde i18n (NL/EN), Multi-Valuta (EUR, USD, GBP) en een dynamische EU OSS Btw-matrix met **live VIES-validatie** voor automatische 0% Btw-verlegging bij zakelijke klanten.
* 🖨️ **Print-on-Demand (Gelato):** Volledige Gelato API v2 integratie. Synchroniseer je catalogus met 1 klik, toon **live verzendtarieven** tijdens de checkout en stuur orders automatisch door naar de drukkerij.
* 🔌 **WooCommerce REST API v3:** Genereer Consumer Keys (`ck_...`) in de admin. Externe tools zoals Sendcloud, Printful, Klaviyo of Exact Online kunnen direct koppelen alsof het een WordPress-site is!
* 📊 **Analytics & CRM:** Ingebouwde SVG-omzetgrafieken, CSV-export voor de boekhouder, en een Klanten-CRM met Lifetime Value (LTV) tracking.
* 📧 **Facturatie & E-mail:** Genereert automatisch printklare A4 HTML/PDF facturen met btw-uitsplitsing. Ingebouwde SMTP & Resend mailer voor orderbevestigingen.
* 🎨 **Thema-Systeem:** Beheer meerdere frontends via de `themes/` map. Wissel met 1 klik in de admin van thema. Go injecteert server-side automatisch OpenGraph en Schema.org JSON-LD voor perfecte SEO.

---

## 🚀 Snelstartgids (Quick Start)

### 1. Vereisten
* **Go 1.23+** geïnstalleerd
* **Node.js 20+** & npm
* **Docker** & Docker Compose

### 2. Database Opstarten (Postgres 17 met UUIDv7)
```bash
docker compose up -d postgres
```
*Draait op hostpoort `5433` (intern `5432`) om conflicten met eventuele lokale PostgreSQL installaties te voorkomen.*

### 3. Dependencies & Backend Starten
```bash
# Go dependencies ophalen
go mod tidy

# Server starten (voert automatisch migraties en demo-seeding uit)
go run cmd/server/main.go
```
De Go server start nu op: **`http://localhost:8080`**.

### 4. Frontend Starten (Tijdens Ontwikkeling met Hot-Reload)
Open een tweede terminal:
```bash
cd frontend
npm install
npm run dev
```
Vite start op `http://localhost:3000` en proxyt alle API calls automatisch door naar de Go backend op `:8080`.

Voor productie-build:
```bash
cd frontend && npm run build && cd ..
```

---

## 🔑 Test- & Beheeraccounts

Zodra de backend voor de eerste keer start, worden automatisch twee accounts klaargezet:

| Type | E-mailadres | Wachtwoord | Rechten & Toegang |
| :--- | :--- | :--- | :--- |
| **WooCommerce Admin** | `admin@aesthetic.be` | `admin123!` | Volledig beheer via het schild-icoon (`/admin`) |
| **Klantaccount** | `gijs@test.be` | `test123!` | Bestelhistoriek, facturen, favorieten, adresbeheer (`/account`) |

---

## 🎨 Het Thema-Systeem (`themes/`)

Elk thema is een standalone frontend-project in de map `themes/`:
```text
themes/
├── aesthetic/          # Het actieve thema
│   ├── theme.json      # Metagegevens (naam, auteur, versie, preview)
│   ├── package.json
│   ├── src/            # Ongebuilde TypeScript / SolidJS broncode
│   └── dist/           # Gecompileerde assets die Go serveert
```

### Nieuw thema maken:
1. Kopieer een bestaande map: `cp -r themes/aesthetic themes/mijn-nieuwe-stijl`
2. Pas de naam aan in `themes/mijn-nieuwe-stijl/theme.json`.
3. Pas de styling/layout aan en run `npm run build`.
4. Ga naar het adminpaneel (`/admin > Thema's`) en klik op **"Activeren"**. Go schakelt direct over!

---

## 🛠️ No-Code Admin Dashboard Functionaliteiten (`/admin`)

1. **Producten & Varianten:** Simpel vs. Variabel, POD-koppelingen, en een centrale Media Bibliotheek.
2. **Bestellingen & Facturen:** Statussen beheren, trackingcodes toevoegen en PDF-facturen inzien.
3. **Kortingscodes (Coupons):** Vaste kortingen of percentages met minimum besteding.
4. **Winkelinstellingen:**
   * **Betalingen:** Schakel eenvoudig tussen **Test/Mock Betaler** en live betalingen via **Mollie** (iDEAL, Bancontact) of **Stripe** (Creditcard, Apple Pay).
   * **Verzending & Btw:** Tarieven, drempels voor gratis verzending en Btw-percentages per land.
   * **E-mail:** Afzendernaam en SMTP/Resend instellingen.
5. **No-Code Homepage Widgets:** Pas teksten, foto's, CTA-knoppen en links live aan voor de Hero Banner, Trust Badges, Uitgelichte Producten en Nieuwsbrief.

---

## 📦 Standalone Docker Productie-Deployment

Wil je de complete winkel (Go server + gecompileerde frontend) als **één enkele container van ~30MB** draaien?
```bash
docker compose up --build -d
```
Beide containers (PostgreSQL 17 en de Go webshop engine) starten automatisch op en herstarten bij crashes.