package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"webshop/internal/database"
)

type OrderItem struct {
	ID           string  `json:"id"`
	OrderID      string  `json:"order_id"`
	ProductID    *string `json:"product_id,omitempty"`
	VariantID    *string `json:"variant_id,omitempty"`
	ProductName  string  `json:"product_name"`
	VariantTitle string  `json:"variant_title"`
	SKU          string  `json:"sku"`
	UnitPrice    float64 `json:"unit_price"`
	Quantity     int     `json:"quantity"`
	TotalPrice   float64 `json:"total_price"`
}

type OrderTimelineItem struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Order struct {
	ID              string              `json:"id"`
	OrderNumber     string              `json:"order_number"`
	UserID          *string             `json:"user_id,omitempty"`
	GuestEmail      *string             `json:"guest_email,omitempty"`
	Status          string              `json:"status"` // pending, processing, shipped, delivered, cancelled
	Subtotal        float64             `json:"subtotal"`
	ShippingCost    float64             `json:"shipping_cost"`
	TotalAmount     float64             `json:"total_amount"`
	ShippingAddress json.RawMessage     `json:"shipping_address"`
	BillingAddress  json.RawMessage     `json:"billing_address"`
	TrackingCode    *string             `json:"tracking_code,omitempty"`
	Carrier         *string             `json:"carrier,omitempty"`
	PaymentStatus   string              `json:"payment_status"` // unpaid, paid, failed, refunded
	PaymentProvider string              `json:"payment_provider"`
	Items           []OrderItem         `json:"items,omitempty"`
	Timeline        []OrderTimelineItem `json:"timeline,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type OrderCreateInput struct {
	UserID          *string
	GuestEmail      *string
	ShippingAddress json.RawMessage
	BillingAddress  json.RawMessage
	Items           []OrderItemInput
	ShippingCost    float64
	PaymentProvider string
}

type OrderItemInput struct {
	ProductID *string `json:"product_id"`
	VariantID *string `json:"variant_id"`
	Quantity  int     `json:"quantity"`
}

type OrderRepo struct {
	db *database.DB
}

func NewOrderRepo(db *database.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func generateOrderNumber() string {
	bytes := make([]byte, 4)
	_, _ = rand.Read(bytes)
	datePart := time.Now().Format("20060102")
	return fmt.Sprintf("ORD-%s-%s", datePart, strings.ToUpper(hex.EncodeToString(bytes)))
}

// CreateOrder voert de bestelling atomisch uit (transactie, voorraad afboeken, timeline initialiseren)
func (r *OrderRepo) CreateOrder(ctx context.Context, input OrderCreateInput) (*Order, error) {
	if len(input.Items) == 0 {
		return nil, errors.New("winkelmandje is leeg")
	}

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	orderNumber := generateOrderNumber()
	var subtotal float64
	var resolvedItems []OrderItem

	for _, item := range input.Items {
		if item.Quantity <= 0 {
			return nil, errors.New("ongeldig aantal artikelen")
		}

		var productName string
		var variantTitle string
		var sku string
		var unitPrice float64
		var currentStock int

		if item.VariantID != nil && *item.VariantID != "" {
			q := `
				SELECT p.name, pv.title, pv.sku, COALESCE(pv.price_override, p.base_price), pv.stock_quantity
				FROM product_variants pv
				JOIN products p ON p.id = pv.product_id
				WHERE pv.id = $1 AND pv.is_active = true AND p.is_active = true
				FOR UPDATE
			`
			err = tx.QueryRow(ctx, q, *item.VariantID).Scan(&productName, &variantTitle, &sku, &unitPrice, &currentStock)
			if err != nil {
				return nil, fmt.Errorf("variant niet gevonden of niet meer actief: %w", err)
			}

			if currentStock < item.Quantity {
				return nil, fmt.Errorf("onvoldoende voorraad voor %s (%s)", productName, variantTitle)
			}

			// Voorraad verminderen
			_, err = tx.Exec(ctx, `UPDATE product_variants SET stock_quantity = stock_quantity - $1 WHERE id = $2`, item.Quantity, *item.VariantID)
			if err != nil {
				return nil, err
			}
		} else if item.ProductID != nil && *item.ProductID != "" {
			q := `SELECT name, 'Standaard', 'SKU-GEN', base_price FROM products WHERE id = $1 AND is_active = true`
			err = tx.QueryRow(ctx, q, *item.ProductID).Scan(&productName, &variantTitle, &sku, &unitPrice)
			if err != nil {
				return nil, fmt.Errorf("product niet gevonden: %w", err)
			}
		} else {
			return nil, errors.New("geen geldig product of variant id meegegeven")
		}

		itemTotal := unitPrice * float64(item.Quantity)
		subtotal += itemTotal

		resolvedItems = append(resolvedItems, OrderItem{
			ProductID:    item.ProductID,
			VariantID:    item.VariantID,
			ProductName:  productName,
			VariantTitle: variantTitle,
			SKU:          sku,
			UnitPrice:    unitPrice,
			Quantity:     item.Quantity,
			TotalPrice:   itemTotal,
		})
	}

	totalAmount := subtotal + input.ShippingCost
	provider := input.PaymentProvider
	if provider == "" {
		provider = "mock"
	}

	orderQuery := `
		INSERT INTO orders (order_number, user_id, guest_email, status, subtotal, shipping_cost, total_amount, shipping_address, billing_address, payment_status, payment_provider)
		VALUES ($1, $2, $3, 'pending', $4, $5, $6, $7, $8, 'unpaid', $9)
		RETURNING id, created_at, updated_at
	`
	order := &Order{
		OrderNumber:     orderNumber,
		UserID:          input.UserID,
		GuestEmail:      input.GuestEmail,
		Status:          "pending",
		Subtotal:        subtotal,
		ShippingCost:    input.ShippingCost,
		TotalAmount:     totalAmount,
		ShippingAddress: input.ShippingAddress,
		BillingAddress:  input.BillingAddress,
		PaymentStatus:   "unpaid",
		PaymentProvider: provider,
	}

	err = tx.QueryRow(ctx, orderQuery, order.OrderNumber, order.UserID, order.GuestEmail, order.Subtotal, order.ShippingCost, order.TotalAmount, order.ShippingAddress, order.BillingAddress, order.PaymentProvider).
		Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Items invoegen
	for i := range resolvedItems {
		resolvedItems[i].OrderID = order.ID
		itemQuery := `
			INSERT INTO order_items (order_id, product_id, variant_id, product_name, variant_title, sku, unit_price, quantity, total_price)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`
		err = tx.QueryRow(ctx, itemQuery, order.ID, resolvedItems[i].ProductID, resolvedItems[i].VariantID, resolvedItems[i].ProductName, resolvedItems[i].VariantTitle, resolvedItems[i].SKU, resolvedItems[i].UnitPrice, resolvedItems[i].Quantity, resolvedItems[i].TotalPrice).
			Scan(&resolvedItems[i].ID)
		if err != nil {
			return nil, err
		}
	}
	order.Items = resolvedItems

	// Eerste timeline event
	_, err = tx.Exec(ctx, `INSERT INTO order_timeline (order_id, status, message) VALUES ($1, $2, $3)`, order.ID, "pending", "Bestelling geplaatst en in afwachting van betaling")
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return order, nil
}

// GetByOrderNumber haalt een order op inclusief items en live status timeline
func (r *OrderRepo) GetByOrderNumber(ctx context.Context, orderNumber string) (*Order, error) {
	query := `
		SELECT id, order_number, user_id, guest_email, status, subtotal, shipping_cost, total_amount, shipping_address, billing_address, tracking_code, carrier, payment_status, payment_provider, created_at, updated_at
		FROM orders
		WHERE order_number = $1
	`
	o := &Order{}
	err := r.db.Pool.QueryRow(ctx, query, orderNumber).Scan(
		&o.ID, &o.OrderNumber, &o.UserID, &o.GuestEmail, &o.Status,
		&o.Subtotal, &o.ShippingCost, &o.TotalAmount, &o.ShippingAddress, &o.BillingAddress,
		&o.TrackingCode, &o.Carrier, &o.PaymentStatus, &o.PaymentProvider, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Items ophalen
	itemsQuery := `
		SELECT id, order_id, product_id, variant_id, product_name, variant_title, sku, unit_price, quantity, total_price
		FROM order_items
		WHERE order_id = $1
	`
	iRows, err := r.db.Pool.Query(ctx, itemsQuery, o.ID)
	if err == nil {
		defer iRows.Close()
		for iRows.Next() {
			var it OrderItem
			if err := iRows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.VariantID, &it.ProductName, &it.VariantTitle, &it.SKU, &it.UnitPrice, &it.Quantity, &it.TotalPrice); err == nil {
				o.Items = append(o.Items, it)
			}
		}
	}

	// Timeline ophalen
	timeQuery := `
		SELECT id, order_id, status, message, created_at
		FROM order_timeline
		WHERE order_id = $1
		ORDER BY created_at ASC
	`
	tRows, err := r.db.Pool.Query(ctx, timeQuery, o.ID)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var tl OrderTimelineItem
			if err := tRows.Scan(&tl.ID, &tl.OrderID, &tl.Status, &tl.Message, &tl.CreatedAt); err == nil {
				o.Timeline = append(o.Timeline, tl)
			}
		}
	}

	return o, nil
}

// ListByUser haalt de bestelgeschiedenis van een klant op
func (r *OrderRepo) ListByUser(ctx context.Context, userID string) ([]Order, error) {
	query := `
		SELECT id, order_number, user_id, guest_email, status, subtotal, shipping_cost, total_amount, shipping_address, billing_address, tracking_code, carrier, payment_status, payment_provider, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.UserID, &o.GuestEmail, &o.Status,
			&o.Subtotal, &o.ShippingCost, &o.TotalAmount, &o.ShippingAddress, &o.BillingAddress,
			&o.TrackingCode, &o.Carrier, &o.PaymentStatus, &o.PaymentProvider, &o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

// Admin: Alle bestellingen filteren en beheren
func (r *OrderRepo) ListAll(ctx context.Context, status string, limit, offset int) ([]Order, int, error) {
	where := ""
	var args []interface{}
	if status != "" {
		where = "WHERE status = $1"
		args = append(args, status)
	}

	countQ := fmt.Sprintf("SELECT COUNT(*) FROM orders %s", where)
	var total int
	if err := r.db.Pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQ := fmt.Sprintf(`
		SELECT id, order_number, user_id, guest_email, status, subtotal, shipping_cost, total_amount, shipping_address, billing_address, tracking_code, carrier, payment_status, payment_provider, created_at, updated_at
		FROM orders
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	args = append(args, limit, offset)
	rows, err := r.db.Pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.UserID, &o.GuestEmail, &o.Status,
			&o.Subtotal, &o.ShippingCost, &o.TotalAmount, &o.ShippingAddress, &o.BillingAddress,
			&o.TrackingCode, &o.Carrier, &o.PaymentStatus, &o.PaymentProvider, &o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

// UpdateStatus werkt de status bij en voegt een tijdlijnevent toe
func (r *OrderRepo) UpdateStatus(ctx context.Context, orderID, newStatus, message string, trackingCode, carrier *string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := `
		UPDATE orders
		SET status = $2, tracking_code = COALESCE($3, tracking_code), carrier = COALESCE($4, carrier), updated_at = now()
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, q, orderID, newStatus, trackingCode, carrier)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO order_timeline (order_id, status, message) VALUES ($1, $2, $3)`, orderID, newStatus, message)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// UpdatePaymentStatus registreert de betalingsstatus
func (r *OrderRepo) UpdatePaymentStatus(ctx context.Context, orderID, paymentStatus string) error {
	q := `UPDATE orders SET payment_status = $2, updated_at = now() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, q, orderID, paymentStatus)
	return err
}