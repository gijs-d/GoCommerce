# ⚡ Modern Go + SolidJS WooCommerce Engine (Headless & No-Code)

Een ultra-snelle, schaalbare en moderne e-commerce webshop engine geschreven in **Go (Golang 1.23+)**, **SolidJS SPA** en **PostgreSQL 17**. 

Ontworpen als een volwaardig alternatief voor WooCommerce/Shopify:
* **No-Code Beheer:** Beheer alle producten, betalingen, drukkerijen (Gelato/POD), thema's en widgets 100% via het ingebouwde Admin Dashboard zonder ooit backend code aan te hoeven raken.
* **Thema-Systeem (WooCommerce Stijl):** Meerdere frontends beheren via de `themes/` map. Wissel met 1 klik in de admin van thema.
* **WooCommerce REST API v3 Compatibel:** Ondersteunt de officiële `/wp-json/wc/v3/*` endpoints zodat externe software (Sendcloud, Gelato, Printful, Klaviyo, Exact Online) naadloos kan koppelen.
* **Dynamische SEO & SSR Injectie:** Go injecteert server-side automatisch OpenGraph meta-tags en Schema.org Product JSON-LD in `index.html` voor crawlers.

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
| **Klantaccount** | `gijs@test.be` | `test123!` | Bestelhistoriek, favorieten, adresbeheer (`/account`) |

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

1. **Producten & Varianten:**
   * **Simpel vs. Variabel:** Vaste prijs/voorraad of attributenmatrix (maten, kleuren, eigen voorraad).
   * **Print-on-Demand (POD):** Koppel artikelen direct aan Gelato of een andere drukkerij met `pod_variant_id` en een drukbestand URL.
   * **Galerij:** Meerdere afbeeldingen toevoegen en koppelen aan specifieke kleurvarianten.
2. **Bestellingen & Live Tijdlijn:**
   * Statussen beheren (`pending`, `processing`, `shipped`, `delivered`, `cancelled`).
   * Track & Trace code invoeren of automatisch laten koppelen door koeriers/drukkerijen.
3. **Kortingscodes (Coupons):**
   * Vaste kortingen of percentages met minimum besteding en vervaldatum.
4. **Winkelinstellingen:**
   * **Betalingen:** Schakel eenvoudig tussen **Test/Mock Betaler** (voor lokaal testen) en live betalingen via **Mollie** (iDEAL, Bancontact) of **Stripe** (Creditcard, Apple Pay).
   * **Verzending:** Standaardtarief en drempelbedrag voor gratis verzending instellen.
   * **Btw & Facturatie:** Btw-percentage en bedrijfs-/btw-nummer voor op facturen.
   * **E-mail:** Afzendernaam en e-mail notificatie instellingen.
5. **No-Code Homepage Widgets:**
   * Pas teksten, foto's, CTA-knoppen en links live aan voor de Hero Banner, Trust Badges, Uitgelichte Producten en Nieuwsbrief.

---

## 📦 Standalone Docker Productie-Deployment

Wil je de complete winkel (Go server + gecompileerde frontend) als **één enkele container van ~30MB** draaien?
```bash
docker compose up --build -d
```
Beide containers (PostgreSQL 17 en de Go webshop engine) starten automatisch op en herstarten bij crashes.