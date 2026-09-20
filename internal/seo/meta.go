package seo

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

type ProductMeta struct {
	ID           string
	Name         string
	Description  string
	ImageURL     string
	Price        string
	Currency     string
	Availability string // "https://schema.org/InStock" of "https://schema.org/OutOfStock"
	Brand        string
	SKU          string
	URL          string
}

type PageMeta struct {
	Title       string
	Description string
	CanonicalURL string
	ImageURL    string
	Type        string // "website" of "product"
	Product     *ProductMeta
}

func DefaultMeta(baseURL string) PageMeta {
	return PageMeta{
		Title:        "Premium T-Shirts & Apparel | Moderne Minimalistische Collectie",
		Description:  "Ontdek onze exclusieve collectie premium t-shirts en basics. Hoogwaardige kwaliteit, perfecte pasvorm en modern design.",
		CanonicalURL: baseURL,
		ImageURL:     baseURL + "/static/og-banner.jpg",
		Type:         "website",
	}
}

func (p PageMeta) GenerateJSONLD() string {
	if p.Type == "product" && p.Product != nil {
		prod := p.Product
		avail := prod.Availability
		if avail == "" {
			avail = "https://schema.org/InStock"
		}
		curr := prod.Currency
		if curr == "" {
			curr = "EUR"
		}
		brandName := prod.Brand
		if brandName == "" {
			brandName = "Shop"
		}

		schema := map[string]interface{}{
			"@context":    "https://schema.org/",
			"@type":       "Product",
			"name":        prod.Name,
			"image":       prod.ImageURL,
			"description": prod.Description,
			"sku":         prod.SKU,
			"brand": map[string]string{
				"@type": "Brand",
				"name":  brandName,
			},
			"offers": map[string]interface{}{
				"@type":         "Offer",
				"url":           prod.URL,
				"priceCurrency": curr,
				"price":         prod.Price,
				"availability":  avail,
			},
		}

		data, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return ""
		}
		return fmt.Sprintf("<script type=\"application/ld+json\">\n%s\n</script>", string(data))
	}

	// Standaard Website / Organisatie schema
	schema := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "WebSite",
		"name":     p.Title,
		"url":      p.CanonicalURL,
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("<script type=\"application/ld+json\">\n%s\n</script>", string(data))
}

// InjectMeta vervangt de placeholders in index.html door geoptimaliseerde meta-tags en JSON-LD
func InjectMeta(htmlContent string, meta PageMeta) string {
	escapedTitle := html.EscapeString(meta.Title)
	escapedDesc := html.EscapeString(meta.Description)
	escapedURL := html.EscapeString(meta.CanonicalURL)
	escapedImage := html.EscapeString(meta.ImageURL)
	ogType := meta.Type
	if ogType == "" {
		ogType = "website"
	}

	jsonLD := meta.GenerateJSONLD()

	tagsBuilder := strings.Builder{}
	tagsBuilder.WriteString(fmt.Sprintf("\t<title>%s</title>\n", escapedTitle))
	tagsBuilder.WriteString(fmt.Sprintf("\t<meta name=\"description\" content=\"%s\">\n", escapedDesc))
	tagsBuilder.WriteString(fmt.Sprintf("\t<link rel=\"canonical\" href=\"%s\">\n", escapedURL))
	tagsBuilder.WriteString("\t<!-- Open Graph / Facebook -->\n")
	tagsBuilder.WriteString(fmt.Sprintf("\t<meta property=\"og:type\" content=\"%s\">\n", ogType))
	tagsBuilder.WriteString(fmt.Sprintf("\t<meta property=\"og:title\" content=\"%s\">\n", escapedTitle))
	tagsBuilder.WriteString(fmt.Sprintf("\t<meta property=\"og:description\" content=\"%s\">\n", escapedDesc))
	tagsBuilder.WriteString(fmt.Sprintf("\t<meta property=\"og:url\" content=\"%s\">\n", escapedURL))
	tagsBuilder.WriteString(fmt.Sprintf("\t<meta property=\"og:image\" content=\"%s\">\n", escapedImage))
	tagsBuilder.WriteString("\t<!-- Twitter Cards -->\n")
	tagsBuilder.WriteString("\t<meta name=\"twitter:card\" content=\"summary_large_image\">\n")
	tagsBuilder.WriteString(fmt.Sprintf("\t<meta name=\"twitter:title\" content=\"%s\">\n", escapedTitle))
	tagsBuilder.WriteString(fmt.Sprintf("\t<meta name=\"twitter:description\" content=\"%s\">\n", escapedDesc))
	tagsBuilder.WriteString(fmt.Sprintf("\t<meta name=\"twitter:image\" content=\"%s\">\n", escapedImage))
	tagsBuilder.WriteString("\t<!-- Rich Snippet JSON-LD -->\n")
	tagsBuilder.WriteString(fmt.Sprintf("\t%s\n", jsonLD))

	replacement := tagsBuilder.String()

	// Als er een placeholder comment in de html zit, vervang die:
	placeholder := "<!-- SEO_TAGS_INJECTION -->"
	if strings.Contains(htmlContent, placeholder) {
		return strings.Replace(htmlContent, placeholder, replacement, 1)
	}

	// Fallback: direct voor de sluitende </head> tag invoegen
	if strings.Contains(htmlContent, "</head>") {
		return strings.Replace(htmlContent, "</head>", replacement+"</head>", 1)
	}

	return htmlContent
}