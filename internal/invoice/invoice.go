package invoice

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"

	"webshop/internal/repository"
)

type InvoiceGenerator struct {
	settingsRepo *repository.SettingsRepo
}

func NewInvoiceGenerator(settingsRepo *repository.SettingsRepo) *InvoiceGenerator {
	return &InvoiceGenerator{settingsRepo: settingsRepo}
}

type CompanyInfo struct {
	StoreName    string `json:"store_name"`
	StoreAddress string `json:"store_address"`
	TaxNumber    string `json:"tax_number"`
	ContactEmail string `json:"contact_email"`
}

// GenerateHTML bouwt een A4 printklare PDF/HTML factuur met officiële Btw-uitsplitsing
func (g *InvoiceGenerator) GenerateHTML(ctx context.Context, order *repository.Order) (string, error) {
	comp := CompanyInfo{
		StoreName:    "Aesthetic Apparel",
		StoreAddress: "Kerkstraat 12, 1000 Brussel, België",
		TaxNumber:    "BE0123.456.789",
		ContactEmail: "contact@aesthetic.be",
	}

	allSettings, err := g.settingsRepo.GetAll(ctx)
	if err == nil {
		if raw, ok := allSettings["store_name"]; ok { _ = json.Unmarshal(raw, &comp.StoreName) }
		if raw, ok := allSettings["store_address"]; ok { _ = json.Unmarshal(raw, &comp.StoreAddress) }
		if raw, ok := allSettings["tax_number"]; ok { _ = json.Unmarshal(raw, &comp.TaxNumber) }
		if raw, ok := allSettings["contact_email"]; ok { _ = json.Unmarshal(raw, &comp.ContactEmail) }
	}

	// Bereken 21% Btw
	subtotalExclVat := order.Subtotal / 1.21
	vatAmount := order.Subtotal - subtotalExclVat

	var addr struct {
		FullName    string `json:"full_name"`
		Street      string `json:"street"`
		HouseNumber string `json:"house_number"`
		Bus         string `json:"bus"`
		City        string `json:"city"`
		PostalCode  string `json:"postal_code"`
		Country     string `json:"country"`
	}
	_ = json.Unmarshal(order.ShippingAddress, &addr)

	tmpl := `
	<!DOCTYPE html>
	<html lang="nl">
	<head>
		<meta charset="utf-8">
		<title>Factuur {{.Order.OrderNumber}}</title>
		<style>
			@page { size: A4; margin: 20mm; }
			* { box-sizing: border-box; }
			body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; color: #0f172a; margin: 0; padding: 40px; background: #fff; line-height: 1.5; }
			.invoice-header { display: flex; justify-content: space-between; border-bottom: 2px solid #e2e8f0; padding-bottom: 24px; margin-bottom: 32px; }
			.company-info h1 { margin: 0 0 8px; font-size: 24px; font-weight: 800; }
			.invoice-meta { text-align: right; }
			.invoice-meta h2 { margin: 0 0 4px; font-size: 20px; color: #64748b; }
			.address-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 40px; margin-bottom: 40px; }
			.address-box h3 { font-size: 13px; text-transform: uppercase; color: #64748b; margin: 0 0 8px; }
			table { width: 100%; border-collapse: collapse; margin-bottom: 32px; }
			th, td { padding: 12px 14px; text-align: left; }
			th { background: #f8fafc; border-bottom: 2px solid #e2e8f0; font-size: 12px; text-transform: uppercase; color: #64748b; }
			td { border-bottom: 1px solid #f1f5f9; font-size: 14px; }
			.totals-table { width: 320px; margin-left: auto; margin-bottom: 32px; }
			.totals-table td { border: none; padding: 6px 14px; }
			.total-row td { border-top: 2px solid #0f172a; font-weight: 800; font-size: 16px; padding-top: 10px; }
			.no-print { margin-bottom: 24px; text-align: right; }
			.btn-print { padding: 10px 20px; background: #0f172a; color: #fff; border: none; border-radius: 8px; font-weight: 600; cursor: pointer; }
			@media print { .no-print { display: none; } body { padding: 0; } }
		</style>
	</head>
	<body>
		<div class="no-print">
			<button class="btn-print" onclick="window.print()">🖨️ Factuur Afdrukken / Opslaan als PDF</button>
		</div>

		<div class="invoice-header">
			<div class="company-info">
				<h1>{{.Company.StoreName}}</h1>
				<div>{{.Company.StoreAddress}}</div>
				<div>Btw/KVK: <strong>{{.Company.TaxNumber}}</strong></div>
				<div>E-mail: {{.Company.ContactEmail}}</div>
			</div>
			<div class="invoice-meta">
				<h2>FACTUUR</h2>
				<div>Factuurnummer: <strong>INV-{{.Order.OrderNumber}}</strong></div>
				<div>Datum: {{.Order.CreatedAt.Format "02-01-2006"}}</div>
				<div>Betaalstatus: <strong>{{.Order.PaymentStatus}}</strong> ({{.Order.PaymentProvider}})</div>
			</div>
		</div>

		<div class="address-grid">
			<div class="address-box">
				<h3>Factuuradres</h3>
				<div><strong>{{.Addr.FullName}}</strong></div>
				<div>{{.Addr.Street}} {{.Addr.HouseNumber}} {{.Addr.Bus}}</div>
				<div>{{.Addr.PostalCode}} {{.Addr.City}}</div>
				<div>{{.Addr.Country}}</div>
			</div>
			<div class="address-box">
				<h3>Afleveradres</h3>
				<div><strong>{{.Addr.FullName}}</strong></div>
				<div>{{.Addr.Street}} {{.Addr.HouseNumber}} {{.Addr.Bus}}</div>
				<div>{{.Addr.PostalCode}} {{.Addr.City}}</div>
				<div>{{.Addr.Country}}</div>
			</div>
		</div>

		<table>
			<thead>
				<tr><th>Omschrijving</th><th>SKU</th><th style="text-align: center;">Aantal</th><th style="text-align: right;">Stukprijs</th><th style="text-align: right;">Totaal</th></tr>
			</thead>
			<tbody>
				{{range .Order.Items}}
				<tr>
					<td><strong>{{.ProductName}}</strong> - {{.VariantTitle}}</td>
					<td><code>{{.SKU}}</code></td>
					<td style="text-align: center;">{{.Quantity}}</td>
					<td style="text-align: right;">€ {{printf "%.2f" .UnitPrice}}</td>
					<td style="text-align: right;">€ {{printf "%.2f" .TotalPrice}}</td>
				</tr>
				{{end}}
			</tbody>
		</table>

		<table class="totals-table">
			<tr><td>Subtotaal (excl. Btw)</td><td style="text-align: right;">€ {{printf "%.2f" .SubtotalExclVat}}</td></tr>
			<tr><td>21% Btw</td><td style="text-align: right;">€ {{printf "%.2f" .VatAmount}}</td></tr>
			<tr><td>Verzendkosten</td><td style="text-align: right;">€ {{printf "%.2f" .Order.ShippingCost}}</td></tr>
			<tr class="total-row"><td>Totaal (incl. Btw)</td><td style="text-align: right;">€ {{printf "%.2f" .Order.TotalAmount}}</td></tr>
		</table>

		<div style="margin-top: 40px; font-size: 12px; color: #94a3b8; text-align: center; border-top: 1px solid #f1f5f9; padding-top: 20px;">
			Dank voor je bestelling bij {{.Company.StoreName}}. Vragen over deze factuur? Neem contact op via {{.Company.ContactEmail}}.
		</div>
	</body>
	</html>`

	t, err := template.New("invoice").Parse(tmpl)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, map[string]interface{}{
		"Company":         comp,
		"Order":           order,
		"Addr":            addr,
		"SubtotalExclVat": subtotalExclVat,
		"VatAmount":       vatAmount,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}