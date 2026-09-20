-- 1. Uitbreiding voor geavanceerde no-code homepage widgets
ALTER TABLE widgets DROP CONSTRAINT IF EXISTS widgets_type_check_v3 CASCADE;
ALTER TABLE widgets DROP CONSTRAINT IF EXISTS widgets_type_check CASCADE;
ALTER TABLE widgets ADD CONSTRAINT widgets_type_check_v3 
    CHECK (type IN ('hero', 'featured_products', 'categories_grid', 'banner_promo', 'trust_badges', 'newsletter', 'category_spotlight', 'best_sellers', 'recent_products'));

-- 2. Digitale & Downloadbare Bestanden tabel
CREATE TABLE IF NOT EXISTS product_downloads (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    file_url TEXT NOT NULL,
    download_limit INT DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 3. Btw-tarieven per land (EU OSS & Belastingmatrix)
CREATE TABLE IF NOT EXISTS tax_rates (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    country_code VARCHAR(2) NOT NULL,
    state_code VARCHAR(10),
    rate NUMERIC(5, 2) NOT NULL,
    name TEXT NOT NULL,
    is_compound BOOLEAN NOT NULL DEFAULT false,
    priority INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Zorg voor een unieke index op country_code zodat ON CONFLICT altijd slaagt
CREATE UNIQUE INDEX IF NOT EXISTS idx_tax_rates_country_unique ON tax_rates (country_code);

-- Standaard EU Btw-tarieven
INSERT INTO tax_rates (country_code, rate, name)
VALUES 
    ('BE', 21.00, 'Belgische Btw (21%)'),
    ('NL', 21.00, 'Nederlandse Btw (21%)'),
    ('DE', 19.00, 'Duitse MwSt (19%)'),
    ('FR', 20.00, 'Franse TVA (20%)')
ON CONFLICT (country_code) DO NOTHING;

-- 4. Terugbetalingen (Refunds)
CREATE TABLE IF NOT EXISTS order_refunds (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    amount NUMERIC(10, 2) NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 5. Generieke Fulfillment & Drukkerij Providers
CREATE TABLE IF NOT EXISTS fulfillment_providers (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL UNIQUE,
    provider_type TEXT NOT NULL,
    api_endpoint TEXT,
    api_key TEXT,
    is_active BOOLEAN NOT NULL DEFAULT false,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO fulfillment_providers (name, provider_type, api_endpoint, is_active)
VALUES 
    ('gelato', 'pod', 'https://order.gelatoapis.com/v2', false),
    ('printful', 'pod', 'https://api.printful.com', false)
ON CONFLICT (name) DO NOTHING;

-- Indexen
CREATE INDEX IF NOT EXISTS idx_tax_country ON tax_rates(country_code);
CREATE INDEX IF NOT EXISTS idx_downloads_product ON product_downloads(product_id);