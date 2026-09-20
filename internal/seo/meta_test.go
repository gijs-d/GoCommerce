package seo_test

import (
	"html"
	"strings"
	"testing"
	"webshop/internal/seo"
)

func TestDefaultMetaInjection(t *testing.T) {
	meta := seo.DefaultMeta("http://localhost:8080")
	if meta.Title == "" || meta.CanonicalURL == "" {
		t.Errorf("Default meta waarden mogen niet leeg zijn")
	}
}

func TestInjectMeta_Basic(t *testing.T) {
	rawHTML := `<!DOCTYPE html>
<html lang="nl">
<head>
    <meta charset="UTF-8">
    <!-- SEO_TAGS_INJECTION -->
</head>
<body>
    <div id="root"></div>
</body>
</html>`

	meta := seo.DefaultMeta("http://localhost:8080")
	result := seo.InjectMeta(rawHTML, meta)

	escapedTitle := html.EscapeString(meta.Title)
	if !strings.Contains(result, "<title>"+escapedTitle+"</title>") {
		t.Errorf("Verwachtte title tag met ge-escapete inhoud in html, kreeg:\n%s", result)
	}

	if !strings.Contains(result, `<meta property="og:type" content="website">`) {
		t.Errorf("Verwachtte og:type website tag in html")
	}

	if !strings.Contains(result, `application/ld+json`) {
		t.Errorf("Verwachtte JSON-LD script tag in html")
	}
}

func TestInjectMeta_Product(t *testing.T) {
	rawHTML := `<!DOCTYPE html><html><head></head><body><div id="root"></div></body></html>`

	meta := seo.PageMeta{
		Title:        "Oversized Heavyweight T-Shirt - Zwart",
		Description:  "100% biologisch katoen, 240 GSM oversized fit.",
		CanonicalURL: "https://myshop.com/product/oversized-black",
		ImageURL:     "https://myshop.com/images/black.jpg",
		Type:         "product",
		Product: &seo.ProductMeta{
			ID:           "prod-123",
			Name:         "Oversized Heavyweight T-Shirt",
			Description:  "100% biologisch katoen",
			ImageURL:     "https://myshop.com/images/black.jpg",
			Price:        "39.95",
			Currency:     "EUR",
			Availability: "https://schema.org/InStock",
			Brand:        "Aesthetic Studios",
			SKU:          "TSH-OVR-BLK-L",
			URL:          "https://myshop.com/product/oversized-black",
		},
	}

	result := seo.InjectMeta(rawHTML, meta)

	if !strings.Contains(result, `<meta property="og:type" content="product">`) {
		t.Errorf("Verwachtte og:type 'product'")
	}

	if !strings.Contains(result, `"price": "39.95"`) {
		t.Errorf("Verwachtte prijs in JSON-LD")
	}

	if !strings.Contains(result, `"sku": "TSH-OVR-BLK-L"`) {
		t.Errorf("Verwachtte SKU in JSON-LD")
	}
}