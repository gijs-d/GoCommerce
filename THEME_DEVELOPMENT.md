# 🎨 Gids voor Thema-Ontwikkeling (THEME_DEVELOPMENT.MD)

Deze handleiding legt exact uit hoe het thema-systeem werkt en hoe je vanaf nul (of via een kopie) een compleet nieuwe frontend kunt bouwen in ieder gewenst framework (SolidJS, Svelte, Vue, React of vanilla HTML/CSS).

---

## 1. Hoe het Thema-Systeem Werkt

De Go-backend fungeert als een **Headless E-Commerce Engine** gecombineerd met een **Dynamic Static Asset Server**:
1. De Go-server beheert alle logica: PostgreSQL database, sessies, producten, betalingen (Stripe/Mollie/Mock), Gelato POD en SEO-injectie.
2. Elk thema bevindt zich in de map `themes/<thema-naam>/`.
3. Wanneer een thema geactiveerd wordt in het Admin Dashboard (`/admin > Thema's`), leest de Go-server de gecompileerde bestanden uit `themes/<thema-naam>/dist/` en injecteert server-side OpenGraph en Schema.org JSON-LD tags in de `index.html`.

---

## 2. Vereiste Bestandsstructuur van een Thema

Elk thema in de map `themes/` heeft minimaal de volgende opbouw:

```text
themes/mijn-thema/
├── theme.json            # [VERPLICHT] Metagegevens voor het Adminpaneel
├── index.html            # [VERPLICHT] HTML template met de SEO injectie-tag
├── package.json          # Dependencies & build scripts (Vite, Vite plugins, etc.)
├── vite.config.ts        # Bundler configuratie gericht op dist/
├── src/                  # De ongebuilde broncode (componenten, styling, stores)
└── dist/                 # [VERPLICHT VOOR PRODUCTIE] Gecompileerde HTML, CSS en JS
    ├── index.html
    └── assets/
        ├── index-xxx.js
        └── index-xxx.css
```

---

## 3. Het `theme.json` Specificatiebestand

Elke themamap **moet** een `theme.json` bevatten in de hoofdmap. Dit bestand wordt door de Go-server gescand om het thema in de WooCommerce Admin te tonen:

```json
{
  "id": "vintage-streetwear",
  "name": "Vintage Streetwear 90s",
  "version": "1.0.0",
  "author": "Jouw Naam of Studio",
  "description": "Retro donker thema met grote typografie en rauwe streetwear vibe.",
  "screenshot": "/themes/vintage-streetwear/screenshot.jpg"
}
```

---

## 4. De `index.html` en Server-Side SEO Injectie

In de `<head>` van je `index.html` **moet** de placeholder tag `<!-- SEO_TAGS_INJECTION -->` staan:

```html
<!DOCTYPE html>
<html lang="nl">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    
    <!-- HIER INJECTEERT DE GO-SERVER DYNAMISCHE SEO: -->
    <!-- SEO_TAGS_INJECTION -->
    
    <!-- Jouw fonts en stylesheets -->
    <link rel="preconnect" href="https://fonts.googleapis.com" />
</head>
<body>
    <div id="root"></div>
    <script type="module" src="/src/index.tsx"></script>
</body>
</html>
```

### Wat Go hier automatisch injecteert voor zoekmachines en social media:
* `<title>` (Dynamisch gebaseerd op product- of categorienaam)
* `<meta name="description" content="...">`
* `<link rel="canonical" href="...">`
* OpenGraph tags (`og:title`, `og:image`, `og:price:amount`, `og:type`) voor Facebook, WhatsApp en iMessage previews.
* Twitter Cards (`twitter:card`, `twitter:image`).
* **Schema.org Product JSON-LD**: Zorgt voor officiële Google Rich Snippets (prijs, voorraad, merk en gele review-sterretjes).

---

## 5. Het "Thema-Contract": Alle Beschikbare API-Endpoints

Elke frontend communiceert met de Go-backend via de onderstaande endpoints. 
**Belangrijk voor authenticatie:** Gebruik bij alle fetch-verzoeken altijd `credentials: 'include'`, zodat de beveiligde HTTP-only sessiecookie (`session_token`) automatisch wordt meegestuurd!

### A. Publieke Catalogus & Producten
| Methode | Endpoint | Omschrijving |
| :--- | :--- | :--- |
| `GET` | `/api/products` | Lijst van producten. Query params: `category`, `search`, `min_price`, `max_price`, `size`, `color`, `sort`, `page`, `limit`. |
| `GET` | `/api/products/:slug` | Volledig product inclusief varianten (maten, kleuren, prijzen, voorraad) en fotogalerij. |
| `GET` | `/api/categories` | Alle actieve categorieën. |
| `GET` | `/api/categories/:slug` | Specifieke categorie info. |
| `GET` | `/api/widgets` | Actieve homepage widgets (Hero, Trust badges, Populaire items, Nieuwsbrief). |
| `GET` | `/api/products/:id/reviews` | Reviews en gemiddelde sterrenscore per product. |
| `POST` | `/api/products/:id/reviews` | Nieuwe review plaatsen (rating 1-5, auteur, toelichting). |

### B. Winkelmand, Verzendkosten & Afrekenen (Checkout)
| Methode | Endpoint | Omschrijving |
| :--- | :--- | :--- |
| `POST` | `/api/shipping-quotes` | Berekent beschikbare verzendmethoden + realtime Gelato POD tarieven op basis van adres en artikelen. |
| `POST` | `/api/coupons/validate` | Valideert een kortingscode (`code`, `subtotal`) en geeft kortingsbedrag terug. |
| `POST` | `/api/taxes/calculate` | Berekent Btw volgens EU OSS regels en past automatische B2B Btw-verlegging toe. |
| `POST` | `/api/taxes/validate-vies` | Controleert een Europees Btw-nummer live bij de VIES-database. |
| `GET` | `/api/payment-methods` | Geeft alle actieve betaalopties terug (Mollie, Stripe, Directe Testbetaler). |
| `POST` | `/api/orders` | Maakt de bestelling definitief aan in de database en boekt voorraad af. |
| `POST` | `/api/payments/initiate` | Start de betaalsessie (`order_number`, `method`) en geeft de redirect-URL naar de bank/Stripe terug. |

### C. Bestelling Volgen & Facturen
| Methode | Endpoint | Omschrijving |
| :--- | :--- | :--- |
| `GET` | `/api/orders/:orderNumber` | Live status van bestelling inclusief tijdlijn-events en Track & Trace link. |
| `POST` | `/api/orders/:orderNumber/pay-mock`| Simuleert een succesvolle testbetaling (handig tijdens testen). |
| `GET` | `/api/orders/:orderNumber/invoice` | Downloadbare / printklare A4 HTML/PDF factuur. |

### D. Klantaccounts & Sessies
| Methode | Endpoint | Omschrijving |
| :--- | :--- | :--- |
| `POST` | `/api/auth/register` | Nieuw klantaccount registreren (`email`, `password`, `full_name`). |
| `POST` | `/api/auth/login` | Inloggen met e-mail en wachtwoord. Start server-side sessie. |
| `POST` | `/api/auth/logout` | Uitloggen (verwijdert sessie direct uit PostgreSQL). |
| `GET` | `/api/auth/me` | Gegevens van de op dat moment ingelogde gebruiker. |
| `GET` | `/api/user/orders` | Bestelgeschiedenis van de ingelogde klant. |
| `GET` | `/api/user/wishlist` | Favorietenlijst van de klant. |
| `POST` | `/api/user/wishlist/:productId` | Product toevoegen aan favorieten. |
| `DELETE`| `/api/user/wishlist/:productId` | Product verwijderen uit favorieten. |
| `GET` | `/api/user/addresses` | Opgeslagen bezorgadressen van de klant. |

---

## 6. Lokale Ontwikkeling & Hot-Reloading

Om aan een thema te werken met live updates zonder de Go-server telkens te herstarten:

1. **Start de Go backend:**
   ```bash
   go run cmd/server/main.go
   ```
2. **Start de dev server in je themamap:**
   ```bash
   cd themes/<thema-naam>
   npm run dev
   ```
3. Zorg dat je `vite.config.ts` de API route proxyt:
   ```ts
   export default defineConfig({
     server: {
       port: 3000,
       proxy: {
         '/api': 'http://localhost:8080',
         '/uploads': 'http://localhost:8080',
         '/wp-json': 'http://localhost:8080',
       }
     }
   });
   ```
   Elke wijziging in je SCSS of componenten is binnen 10 milliseconden live te zien op `http://localhost:3000`.

---

## 7. Productie-Build & Activering in de Admin

Zodra je tevreden bent met het ontwerp:

1. **Compileer de frontend:**
   ```bash
   npm run build
   ```
   Dit maakt de map `themes/<thema-naam>/dist/` aan.
2. **Thema Activeren:**
   * Open je browser en ga naar `http://localhost:8080/admin`.
   * Klik op het tabblad **"Thema's"**.
   * Je nieuwe thema verschijnt automatisch in de lijst met de naam en omschrijving uit `theme.json`.
   * Klik op **"Activeren"**. Vanaf dat moment serveert de Go-server dit thema live aan alle bezoekers!

---

## 8. Checklist van Aanbevolen Pagina's per Thema

Een volwaardig thema bevat componenten en routes voor:
- [ ] **Homepagina (`/`)**: Hero banner, actieve widgets, uitgelichte collectie, trust badges.
- [ ] **Catalogus (`/shop`)**: Productenlijst met filters (categorie, maat, kleur, prijs) en sortering.
- [ ] **Product Detail Pagina (`/product/:slug`)**: Variantenselectie (maten/kleuren met dynamische prijs en voorraadstatus), fotogalerij, review-score en toevoegen aan mandje.
- [ ] **Winkelmandje (Drawer of `/cart`)**: Overzicht van artikelen, aantallen aanpassen, subtotaal.
- [ ] **Checkout (`/checkout`)**: Verzendadres, realtime verzendmethode selectie, kortingscode invoer, zakelijke B2B Btw-verlegging, betaalmethode keuze.
- [ ] **Bestelling Volgen (`/track/:orderNumber`)**: Tijdlijn van de bestelling, koeriergegevens, knop voor PDF-factuur.
- [ ] **Klantprofiel (`/account`)**: Tabbladen voor bestelhistoriek, favorieten en adresboek.
- [ ] **Authenticatie (`/login`)**: Inlog- en registratieformulier.