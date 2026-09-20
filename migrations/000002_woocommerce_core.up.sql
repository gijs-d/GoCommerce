-- 1. Producttypes en POD (Print-on-demand) velden toevoegen
ALTER TABLE products ADD COLUMN IF NOT EXISTS product_type TEXT NOT NULL DEFAULT 'variable' CHECK (product_type IN ('simple', 'variable', 'digital'));

ALTER TABLE product_variants ADD COLUMN IF NOT EXISTS pod_provider TEXT;       -- bijv. 'gelato', 'printify' of NULL voor eigen voorraad
ALTER TABLE product_variants ADD COLUMN IF NOT EXISTS pod_variant_id TEXT;     -- Uniek artikelnummer van Gelato
ALTER TABLE product_variants ADD COLUMN IF NOT EXISTS print_file_url TEXT;     -- Hoge resolutie drukbestand URL

-- 2. Kortingscodes & Waardebonnen (Coupons)
CREATE TABLE IF NOT EXISTS coupons (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    code TEXT UNIQUE NOT NULL,
    discount_type TEXT NOT NULL CHECK (discount_type IN ('percent', 'fixed')),
    discount_value NUMERIC(10, 2) NOT NULL,
    min_spend NUMERIC(10, 2) DEFAULT 0.00,
    max_uses INT DEFAULT NULL,
    uses_count INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ DEFAULT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 3. Verzendmethoden & Tarieven (Shipping Zones)
CREATE TABLE IF NOT EXISTS shipping_methods (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    title TEXT NOT NULL,
    description TEXT,
    cost NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    free_threshold NUMERIC(10, 2) DEFAULT NULL, -- Gratis vanaf dit bedrag
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 4. Product Reviews & Beoordelingen
CREATE TABLE IF NOT EXISTS product_reviews (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    author_name TEXT NOT NULL,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT NOT NULL,
    is_approved BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 5. Webhooks voor externe automatisering (Zapier, Make, Boekhouding)
CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL,
    target_url TEXT NOT NULL,
    secret TEXT,
    events TEXT[] NOT NULL, -- bijv. ['order.created', 'order.paid']
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 6. REST API Keys voor externe apps
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    description TEXT NOT NULL,
    consumer_key TEXT UNIQUE NOT NULL,
    consumer_secret_hash TEXT NOT NULL,
    permissions TEXT NOT NULL DEFAULT 'read' CHECK (permissions IN ('read', 'write', 'read_write')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Standaard verzendmethoden zaaien indien leeg
INSERT INTO shipping_methods (title, description, cost, free_threshold, sort_order)
VALUES 
    ('Standaard Bezorging', 'Binnen 1-2 werkdagen in huis', 4.95, 50.00, 1),
    ('Express Bezorging', 'Volgende werkdag geleverd', 8.95, NULL, 2),
    ('Gratis Afhalen', 'Afhalen op onze locatie', 0.00, NULL, 3)
ON CONFLICT DO NOTHING;

-- Standaard test-coupon aanmaken (bijv. WELKOM10 voor 10% korting)
INSERT INTO coupons (code, discount_type, discount_value, min_spend, is_active)
VALUES ('WELKOM10', 'percent', 10.00, 25.00, true)
ON CONFLICT DO NOTHING;

-- Indexen voor performantie
CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_reviews_product ON product_reviews(product_id);
CREATE INDEX IF NOT EXISTS idx_reviews_approved ON product_reviews(is_approved);