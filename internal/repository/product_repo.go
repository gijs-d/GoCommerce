package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"webshop/internal/database"
)

type ProductVariant struct {
	ID            string   `json:"id"`
	ProductID     string   `json:"product_id"`
	SKU           string   `json:"sku"`
	Title         string   `json:"title"`
	Size          *string  `json:"size,omitempty"`
	Color         *string  `json:"color,omitempty"`
	ColorHex      *string  `json:"color_hex,omitempty"`
	PriceOverride *float64 `json:"price_override,omitempty"`
	StockQuantity int      `json:"stock_quantity"`
	ImageURL      *string  `json:"image_url,omitempty"`
	IsActive      bool     `json:"is_active"`
}

type ProductImage struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	VariantID *string `json:"variant_id,omitempty"`
	URL       string  `json:"url"`
	AltText   *string `json:"alt_text,omitempty"`
	SortOrder int     `json:"sort_order"`
	IsPrimary bool    `json:"is_primary"`
}

type Product struct {
	ID               string           `json:"id"`
	CategoryID       *string          `json:"category_id,omitempty"`
	CategoryName     *string          `json:"category_name,omitempty"`
	CategorySlug     *string          `json:"category_slug,omitempty"`
	Name             string           `json:"name"`
	Slug             string           `json:"slug"`
	ShortDescription *string          `json:"short_description,omitempty"`
	Description      string           `json:"description"`
	BasePrice        float64          `json:"base_price"`
	Featured         bool             `json:"featured"`
	IsActive         bool             `json:"is_active"`
	SEOTitle         *string          `json:"seo_title,omitempty"`
	SEODescription   *string          `json:"seo_description,omitempty"`
	PrimaryImage     *string          `json:"primary_image,omitempty"`
	Variants         []ProductVariant `json:"variants,omitempty"`
	Images           []ProductImage   `json:"images,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type ProductFilter struct {
	CategorySlug string
	Search       string
	MinPrice     *float64
	MaxPrice     *float64
	Size         string
	Color        string
	FeaturedOnly bool
	ActiveOnly   bool
	SortBy       string // "price_asc", "price_desc", "newest", "name"
	Limit        int
	Offset       int
}

type ProductRepo struct {
	db *database.DB
}

func NewProductRepo(db *database.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

// ListProducts haalt een gefilterde lijst op inclusief de hoofdafbeelding
func (r *ProductRepo) ListProducts(ctx context.Context, filter ProductFilter) ([]Product, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.ActiveOnly {
		conditions = append(conditions, fmt.Sprintf("p.is_active = $%d", argIdx))
		args = append(args, true)
		argIdx++
	}

	if filter.FeaturedOnly {
		conditions = append(conditions, fmt.Sprintf("p.featured = $%d", argIdx))
		args = append(args, true)
		argIdx++
	}

	if filter.CategorySlug != "" {
		conditions = append(conditions, fmt.Sprintf("c.slug = $%d", argIdx))
		args = append(args, filter.CategorySlug)
		argIdx++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("p.base_price >= $%d", argIdx))
		args = append(args, *filter.MinPrice)
		argIdx++
	}

	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("p.base_price <= $%d", argIdx))
		args = append(args, *filter.MaxPrice)
		argIdx++
	}

	if filter.Size != "" {
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM product_variants pv WHERE pv.product_id = p.id AND pv.size = $%d AND pv.is_active = true)", argIdx))
		args = append(args, filter.Size)
		argIdx++
	}

	if filter.Color != "" {
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM product_variants pv WHERE pv.product_id = p.id AND pv.color = $%d AND pv.is_active = true)", argIdx))
		args = append(args, filter.Color)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Totaal aantal voor paginering
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT p.id)
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		%s
	`, whereClause)

	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Sortering
	orderBy := "p.created_at DESC"
	switch filter.SortBy {
	case "price_asc":
		orderBy = "p.base_price ASC"
	case "price_desc":
		orderBy = "p.base_price DESC"
	case "name":
		orderBy = "p.name ASC"
	case "newest":
		orderBy = "p.created_at DESC"
	}

	limitClause := ""
	if filter.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			p.id, p.category_id, c.name, c.slug, p.name, p.slug, 
			p.short_description, p.description, p.base_price, p.featured, 
			p.is_active, p.seo_title, p.seo_description,
			(SELECT url FROM product_images pi WHERE pi.product_id = p.id ORDER BY pi.is_primary DESC, pi.sort_order ASC LIMIT 1) as primary_image,
			p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		%s
		ORDER BY %s
		%s
	`, whereClause, orderBy, limitClause)

	rows, err := r.db.Pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID, &p.CategoryID, &p.CategoryName, &p.CategorySlug, &p.Name, &p.Slug,
			&p.ShortDescription, &p.Description, &p.BasePrice, &p.Featured,
			&p.IsActive, &p.SEOTitle, &p.SEODescription, &p.PrimaryImage,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}

	return products, total, nil
}

// GetBySlug haalt het complete product op inclusief varianten en afbeeldingen
func (r *ProductRepo) GetBySlug(ctx context.Context, slug string) (*Product, error) {
	query := `
		SELECT 
			p.id, p.category_id, c.name, c.slug, p.name, p.slug, 
			p.short_description, p.description, p.base_price, p.featured, 
			p.is_active, p.seo_title, p.seo_description,
			(SELECT url FROM product_images pi WHERE pi.product_id = p.id ORDER BY pi.is_primary DESC, pi.sort_order ASC LIMIT 1) as primary_image,
			p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.slug = $1
	`
	p := &Product{}
	err := r.db.Pool.QueryRow(ctx, query, slug).Scan(
		&p.ID, &p.CategoryID, &p.CategoryName, &p.CategorySlug, &p.Name, &p.Slug,
		&p.ShortDescription, &p.Description, &p.BasePrice, &p.Featured,
		&p.IsActive, &p.SEOTitle, &p.SEODescription, &p.PrimaryImage,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Varianten ophalen
	vQuery := `
		SELECT id, product_id, sku, title, size, color, color_hex, price_override, stock_quantity, image_url, is_active
		FROM product_variants
		WHERE product_id = $1 AND is_active = true
		ORDER BY size ASC, color ASC
	`
	vRows, err := r.db.Pool.Query(ctx, vQuery, p.ID)
	if err == nil {
		defer vRows.Close()
		for vRows.Next() {
			var v ProductVariant
			if err := vRows.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Title, &v.Size, &v.Color, &v.ColorHex, &v.PriceOverride, &v.StockQuantity, &v.ImageURL, &v.IsActive); err == nil {
				p.Variants = append(p.Variants, v)
			}
		}
	}

	// Afbeeldingen ophalen
	imgQuery := `
		SELECT id, product_id, variant_id, url, alt_text, sort_order, is_primary
		FROM product_images
		WHERE product_id = $1
		ORDER BY is_primary DESC, sort_order ASC
	`
	iRows, err := r.db.Pool.Query(ctx, imgQuery, p.ID)
	if err == nil {
		defer iRows.Close()
		for iRows.Next() {
			var img ProductImage
			if err := iRows.Scan(&img.ID, &img.ProductID, &img.VariantID, &img.URL, &img.AltText, &img.SortOrder, &img.IsPrimary); err == nil {
				p.Images = append(p.Images, img)
			}
		}
	}

	return p, nil
}

// Admin: CRUD Operaties
func (r *ProductRepo) Create(ctx context.Context, categoryID *string, name, slug string, shortDesc *string, desc string, basePrice float64, featured, isActive bool, seoTitle, seoDesc *string) (*Product, error) {
	query := `
		INSERT INTO products (category_id, name, slug, short_description, description, base_price, featured, is_active, seo_title, seo_description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, category_id, name, slug, short_description, description, base_price, featured, is_active, seo_title, seo_description, created_at, updated_at
	`
	p := &Product{}
	err := r.db.Pool.QueryRow(ctx, query, categoryID, name, slug, shortDesc, desc, basePrice, featured, isActive, seoTitle, seoDesc).
		Scan(&p.ID, &p.CategoryID, &p.Name, &p.Slug, &p.ShortDescription, &p.Description, &p.BasePrice, &p.Featured, &p.IsActive, &p.SEOTitle, &p.SEODescription, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProductRepo) Update(ctx context.Context, id string, categoryID *string, name, slug string, shortDesc *string, desc string, basePrice float64, featured, isActive bool, seoTitle, seoDesc *string) error {
	query := `
		UPDATE products
		SET category_id = $2, name = $3, slug = $4, short_description = $5, description = $6, base_price = $7, featured = $8, is_active = $9, seo_title = $10, seo_description = $11, updated_at = now()
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id, categoryID, name, slug, shortDesc, desc, basePrice, featured, isActive, seoTitle, seoDesc)
	return err
}

func (r *ProductRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *ProductRepo) UpsertVariant(ctx context.Context, v ProductVariant) error {
	query := `
		INSERT INTO product_variants (product_id, sku, title, size, color, color_hex, price_override, stock_quantity, image_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (sku) DO UPDATE
		SET title = EXCLUDED.title, size = EXCLUDED.size, color = EXCLUDED.color, color_hex = EXCLUDED.color_hex, price_override = EXCLUDED.price_override, stock_quantity = EXCLUDED.stock_quantity, image_url = EXCLUDED.image_url, is_active = EXCLUDED.is_active, updated_at = now()
	`
	_, err := r.db.Pool.Exec(ctx, query, v.ProductID, v.SKU, v.Title, v.Size, v.Color, v.ColorHex, v.PriceOverride, v.StockQuantity, v.ImageURL, v.IsActive)
	return err
}

func (r *ProductRepo) DeleteVariant(ctx context.Context, variantID string) error {
	query := `DELETE FROM product_variants WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, variantID)
	return err
}

func (r *ProductRepo) AddImage(ctx context.Context, productID string, variantID *string, url string, altText *string, sortOrder int, isPrimary bool) error {
	if isPrimary {
		_, _ = r.db.Pool.Exec(ctx, `UPDATE product_images SET is_primary = false WHERE product_id = $1`, productID)
	}
	query := `
		INSERT INTO product_images (product_id, variant_id, url, alt_text, sort_order, is_primary)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Pool.Exec(ctx, query, productID, variantID, url, altText, sortOrder, isPrimary)
	return err
}

func (r *ProductRepo) DeleteImage(ctx context.Context, imageID string) error {
	query := `DELETE FROM product_images WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, imageID)
	return err
}