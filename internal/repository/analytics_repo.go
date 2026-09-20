package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"webshop/internal/database"
)

type DailySales struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type TopProduct struct {
	ProductName  string  `json:"product_name"`
	TotalSold    int     `json:"total_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

type CategorySales struct {
	CategoryName string  `json:"category_name"`
	TotalRevenue float64 `json:"total_revenue"`
	Percentage   float64 `json:"percentage"`
}

type AnalyticsOverview struct {
	TotalRevenue  float64         `json:"total_revenue"`
	TotalOrders   int             `json:"total_orders"`
	AverageOrder  float64         `json:"average_order"`
	ItemsSold     int             `json:"items_sold"`
	DailySales    []DailySales    `json:"daily_sales"`
	TopProducts   []TopProduct    `json:"top_products"`
	CategorySales []CategorySales `json:"category_sales"`
}

type AnalyticsRepo struct {
	db *database.DB
}

func NewAnalyticsRepo(db *database.DB) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

// GetOverview berekent de complete verkoopanalyse voor een bepaalde periode (in dagen)
func (r *AnalyticsRepo) GetOverview(ctx context.Context, days int) (*AnalyticsOverview, error) {
	if days <= 0 {
		days = 30
	}

	overview := &AnalyticsOverview{
		DailySales:    []DailySales{},
		TopProducts:   []TopProduct{},
		CategorySales: []CategorySales{},
	}

	// 1. Algemene totalen in de periode
	totalsQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(total_amount), 0),
			COUNT(id),
			COALESCE(AVG(total_amount), 0)
		FROM orders
		WHERE payment_status = 'paid' AND created_at >= NOW() - INTERVAL '%d days'
	`, days)
	_ = r.db.Pool.QueryRow(ctx, totalsQuery).Scan(&overview.TotalRevenue, &overview.TotalOrders, &overview.AverageOrder)

	// Totaal aantal stuks verkocht
	itemsSoldQuery := fmt.Sprintf(`
		SELECT COALESCE(SUM(oi.quantity), 0)
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.payment_status = 'paid' AND o.created_at >= NOW() - INTERVAL '%d days'
	`, days)
	_ = r.db.Pool.QueryRow(ctx, itemsSoldQuery).Scan(&overview.ItemsSold)

	// 2. Dagelijkse omzet & orders (met generate_series)
	dailyQuery := fmt.Sprintf(`
		SELECT 
			to_char(d.date, 'YYYY-MM-DD'),
			COALESCE(SUM(o.total_amount), 0),
			COUNT(o.id)
		FROM generate_series(CURRENT_DATE - INTERVAL '%d days', CURRENT_DATE, '1 day'::interval) d(date)
		LEFT JOIN orders o ON date_trunc('day', o.created_at) = d.date AND o.payment_status = 'paid'
		GROUP BY d.date
		ORDER BY d.date ASC
	`, days-1)
	dRows, err := r.db.Pool.Query(ctx, dailyQuery)
	if err == nil {
		defer dRows.Close()
		for dRows.Next() {
			var ds DailySales
			if err := dRows.Scan(&ds.Date, &ds.Revenue, &ds.Orders); err == nil {
				overview.DailySales = append(overview.DailySales, ds)
			}
		}
	}

	// 3. Bestverkochte producten
	topQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(oi.product_name, 'Onbekend'),
			SUM(oi.quantity) as total_qty,
			SUM(oi.total_price) as total_rev
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.payment_status = 'paid' AND o.created_at >= NOW() - INTERVAL '%d days'
		GROUP BY oi.product_name
		ORDER BY total_qty DESC
		LIMIT 5
	`, days)
	tRows, err := r.db.Pool.Query(ctx, topQuery)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var tp TopProduct
			if err := tRows.Scan(&tp.ProductName, &tp.TotalSold, &tp.TotalRevenue); err == nil {
				overview.TopProducts = append(overview.TopProducts, tp)
			}
		}
	}

	// 4. Omzet per categorie
	catQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(c.name, 'Algemeen'),
			COALESCE(SUM(oi.total_price), 0) as cat_rev
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		LEFT JOIN products p ON p.id = oi.product_id
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE o.payment_status = 'paid' AND o.created_at >= NOW() - INTERVAL '%d days'
		GROUP BY c.name
		ORDER BY cat_rev DESC
	`, days)
	cRows, err := r.db.Pool.Query(ctx, catQuery)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var cs CategorySales
			if err := cRows.Scan(&cs.CategoryName, &cs.TotalRevenue); err == nil {
				if overview.TotalRevenue > 0 {
					cs.Percentage = (cs.TotalRevenue / overview.TotalRevenue) * 100.0
				}
				overview.CategorySales = append(overview.CategorySales, cs)
			}
		}
	}

	return overview, nil
}

// GenerateCSVExport bouwt een downloadbare CSV voor de boekhouding
func (r *AnalyticsRepo) GenerateCSVExport(ctx context.Context) ([]byte, error) {
	query := `
		SELECT 
			order_number, to_char(created_at, 'YYYY-MM-DD HH24:MI:SS'),
			COALESCE(guest_email, 'Geregistreerde Klant'),
			subtotal, shipping_cost, total_amount, payment_provider, payment_status, status
		FROM orders
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)
	writer.Comma = ';' // Europese Excel puntkomma

	_ = writer.Write([]string{
		"Ordernummer", "Datum", "Klant / E-mail", "Subtotaal", "Verzendkosten", "Totaalbedrag", "Betaalmethode", "Betaalstatus", "Bestelstatus",
	})

	for rows.Next() {
		var orderNum, date, customer, provider, payStatus, status string
		var subtotal, shipping, total float64
		if err := rows.Scan(&orderNum, &date, &customer, &subtotal, &shipping, &total, &provider, &payStatus, &status); err == nil {
			_ = writer.Write([]string{
				orderNum, date, customer,
				fmt.Sprintf("%.2f", subtotal),
				fmt.Sprintf("%.2f", shipping),
				fmt.Sprintf("%.2f", total),
				provider, payStatus, status,
			})
		}
	}
	writer.Flush()

	return buf.Bytes(), nil
}