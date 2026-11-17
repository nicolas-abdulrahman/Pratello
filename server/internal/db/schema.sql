-- Pratello Database Schema
-- Generated from docs/db.dbml

-- -----------------------------------------------------
-- Enums
-- -----------------------------------------------------

DO $$ BEGIN
    CREATE TYPE card_brand_enum AS ENUM (
        'VISA',
        'MASTERCARD',
        'ELO',
        'AMEX',
        'HIPERCARD',
        'OTHER'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE discount_type_enum AS ENUM (
        'PERCENTAGE',
        'FIXED_AMOUNT'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE menu_style_enum AS ENUM (
        'CLASSICO',
        'EXPERIENCIA'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE order_status_enum AS ENUM (
        'PENDING',
        'PREPARING',
        'READY',
        'COMPLETED',
        'CANCELLED'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE order_type_enum AS ENUM (
        'LOCAL',
        'REMOTE'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE transaction_status_enum AS ENUM (
        'PENDING',
        'COMPLETED',
        'EXPIRED',
        'FAILED',
        'CANCELLED'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE transaction_type_enum AS ENUM (
        'PIX',
        'CREDIT',
        'DEBIT',
        'CASH'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- -----------------------------------------------------
-- Tables
-- -----------------------------------------------------

-- Table: Address
CREATE TABLE IF NOT EXISTS "Address" (
    "id" SERIAL PRIMARY KEY,
    "CEP" VARCHAR(500) NOT NULL,
    "number" INT NOT NULL,
    "apt" VARCHAR(500),
    "complementary" VARCHAR(500)
);

-- Table: restaurant
CREATE TABLE IF NOT EXISTS "restaurant" (
    "id" SERIAL PRIMARY KEY,
    "name_slug" VARCHAR(500) NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "brand" VARCHAR(500),
    "cnpj" VARCHAR(500),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: account
CREATE TABLE IF NOT EXISTS "account" (
    "id" SERIAL PRIMARY KEY,
    "name" VARCHAR(500),
    "email" VARCHAR(255) UNIQUE NOT NULL,
    "password" VARCHAR(255) NOT NULL,
    "phone" VARCHAR(500),
    "address_id" INT REFERENCES "Address"("id") ON DELETE SET NULL,
    "document" VARCHAR(500),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: sysadmin
CREATE TABLE IF NOT EXISTS "sysadmin" (
    "id" SERIAL PRIMARY KEY,
    "account_id" INT NOT NULL UNIQUE REFERENCES "account"("id") ON DELETE CASCADE,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: physical_restaurant
CREATE TABLE IF NOT EXISTS "physical_restaurant" (
    "id" SERIAL,
    "restaurant_id" INT NOT NULL REFERENCES "restaurant"("id") ON DELETE CASCADE,
    "address_id" INT NOT NULL REFERENCES "Address"("id") ON DELETE RESTRICT,
    "name" VARCHAR(500) NOT NULL DEFAULT '',
    "cnpj" VARCHAR(500) NOT NULL,
    "owner_cnpj" VARCHAR(500) NOT NULL DEFAULT '',
    "terminal_pin" VARCHAR(6) NOT NULL DEFAULT '123456',
    "pin_updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "failed_pin_attempts" INT NOT NULL DEFAULT 0,
    "locked_until" TIMESTAMPTZ,
    PRIMARY KEY ("id", "restaurant_id", "address_id"),
    CONSTRAINT "uq_physical_restaurant_id" UNIQUE ("id")
);

-- Table: restaurant_phone
CREATE TABLE IF NOT EXISTS "restaurant_phone" (
    "physical_restaurant_id" INT PRIMARY KEY REFERENCES "physical_restaurant"("id") ON DELETE CASCADE,
    "phone" VARCHAR(500) NOT NULL,
    "label" VARCHAR(50) NOT NULL DEFAULT 'Principal'
);

-- Table: employee
CREATE TABLE IF NOT EXISTS "employee" (
    "id" SERIAL PRIMARY KEY,
    "physical_restaurant_id" INT NOT NULL REFERENCES "physical_restaurant"("id") ON DELETE CASCADE,
    "name" VARCHAR(500) NOT NULL,
    "cpf" VARCHAR(14) NOT NULL,
    "position" VARCHAR(255) NOT NULL,
    "account_id" INT REFERENCES "account"("id") ON DELETE SET NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: terminal_session
CREATE TABLE IF NOT EXISTS "terminal_session" (
    "id" SERIAL PRIMARY KEY,
    "physical_restaurant_id" INT NOT NULL REFERENCES "physical_restaurant"("id") ON DELETE CASCADE,
    "token" VARCHAR(64) UNIQUE NOT NULL,
    "device_id" VARCHAR(255) NOT NULL,
    "expires_at" TIMESTAMPTZ NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: staff_session
CREATE TABLE IF NOT EXISTS "staff_session" (
    "id" SERIAL PRIMARY KEY,
    "physical_restaurant_id" INT NOT NULL REFERENCES "physical_restaurant"("id") ON DELETE CASCADE,
    "shift_code" VARCHAR(6) NOT NULL,
    "qr_token" VARCHAR(64) UNIQUE NOT NULL,
    "is_active" BOOLEAN NOT NULL DEFAULT TRUE,
    "expires_at" TIMESTAMPTZ NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: account_restaurant
CREATE TABLE IF NOT EXISTS "account_restaurant" (
    "id" BIGSERIAL,
    "account_id" INT REFERENCES "account"("id") ON DELETE CASCADE,
    "restaurant_id" INT NOT NULL REFERENCES "restaurant"("id") ON DELETE CASCADE,
    "account_type" TEXT NOT NULL DEFAULT 'manager',
    "document" VARCHAR(500),
    PRIMARY KEY ("id", "restaurant_id")
);

-- Table: user_restaurant
CREATE TABLE IF NOT EXISTS "user_restaurant" (
    "id" SERIAL PRIMARY KEY,
    "account_id" INT REFERENCES "account"("id") ON DELETE SET NULL,
    "restaurant_id" INT NOT NULL REFERENCES "restaurant"("id") ON DELETE CASCADE,
    "coins_amount" INT NOT NULL DEFAULT 0,
    "cpf" VARCHAR(500)
);

-- Table: category
CREATE TABLE IF NOT EXISTS "category" (
    "id" SERIAL PRIMARY KEY,
    "restaurant_id" INT NOT NULL REFERENCES "restaurant"("id") ON DELETE CASCADE,
    "name" VARCHAR(100) NOT NULL
);

-- Table: menu
CREATE TABLE IF NOT EXISTS "menu" (
    "id" SERIAL,
    "restaurant_id" INT NOT NULL REFERENCES "restaurant"("id") ON DELETE CASCADE,
    "name" VARCHAR(255) NOT NULL,
    "style" menu_style_enum NOT NULL,
    "accent_color" VARCHAR(20),
    PRIMARY KEY ("id", "restaurant_id")
);

-- Table: coupon
CREATE TABLE IF NOT EXISTS "coupon" (
    "id" SERIAL PRIMARY KEY,
    "restaurant_id" INT NOT NULL REFERENCES "restaurant"("id") ON DELETE CASCADE,
    "code" VARCHAR(50) NOT NULL,
    "discount_type" discount_type_enum NOT NULL,
    "discount_value" NUMERIC(10,2) NOT NULL,
    "min_order_value" NUMERIC(10,2) DEFAULT 0,
    "max_uses" INT,
    "is_active" BOOLEAN NOT NULL DEFAULT TRUE,
    "expires_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT "idx_coupons_restaurant_id_code" UNIQUE ("restaurant_id", "code")
);

-- Table: menu_item
CREATE TABLE IF NOT EXISTS "menu_item" (
    "id" SERIAL PRIMARY KEY,
    "restaurant_id" INT NOT NULL REFERENCES "restaurant"("id") ON DELETE CASCADE,
    "category_id" INT NOT NULL REFERENCES "category"("id") ON DELETE RESTRICT,
    "name" VARCHAR(255) NOT NULL,
    "description" TEXT,
    "price" NUMERIC(10,2) NOT NULL,
    "image_url" TEXT,
    "is_available" BOOLEAN DEFAULT TRUE,
    "is_combo" BOOLEAN NOT NULL DEFAULT FALSE
);

-- Table: item_sale
CREATE TABLE IF NOT EXISTS "item_sale" (
    "item_id" INT PRIMARY KEY REFERENCES "menu_item"("id") ON DELETE CASCADE,
    "discount_type" discount_type_enum NOT NULL DEFAULT 'FIXED_AMOUNT',
    "value" NUMERIC(10,2) NOT NULL,
    "is_active" BOOLEAN NOT NULL DEFAULT TRUE,
    "starts_at" TIMESTAMPTZ,
    "ends_at" TIMESTAMPTZ
);

-- Table: item_side
CREATE TABLE IF NOT EXISTS "item_side" (
    "id" SERIAL PRIMARY KEY,
    "item_id" INT NOT NULL REFERENCES "menu_item"("id") ON DELETE CASCADE,
    "name" VARCHAR(255) NOT NULL,
    "amount" INT NOT NULL DEFAULT 0,
    "price" NUMERIC(10,2) NOT NULL
);

-- Table: item_ingredient
CREATE TABLE IF NOT EXISTS "item_ingredient" (
    "item_id" INT NOT NULL REFERENCES "menu_item"("id") ON DELETE CASCADE,
    "ingredient_id" INT NOT NULL,
    "is_removable" BOOLEAN NOT NULL DEFAULT TRUE,
    "quantity" INT,
    "measure" VARCHAR(50),
    PRIMARY KEY ("item_id", "ingredient_id")
);

-- Table: combo_item
CREATE TABLE IF NOT EXISTS "combo_item" (
    "combo_id" INT NOT NULL REFERENCES "menu_item"("id") ON DELETE CASCADE,
    "included_item_id" INT NOT NULL REFERENCES "menu_item"("id") ON DELETE CASCADE,
    "quantity" INT NOT NULL DEFAULT 1,
    PRIMARY KEY ("combo_id", "included_item_id")
);

-- Table: order
CREATE TABLE IF NOT EXISTS "order" (
    "id" SERIAL PRIMARY KEY,
    "physical_restaurant_id" INT NOT NULL REFERENCES "physical_restaurant"("id") ON DELETE RESTRICT,
    "coupon_id" INT REFERENCES "coupon"("id") ON DELETE SET NULL,
    "order_type" order_type_enum NOT NULL,
    "customer_name" VARCHAR(255),
    "subtotal" NUMERIC(10,2) NOT NULL DEFAULT 0,
    "discount_amount" NUMERIC(10,2) DEFAULT 0,
    "total" NUMERIC(10,2) NOT NULL DEFAULT 0,
    "status" order_status_enum NOT NULL DEFAULT 'PENDING',
    "review" INT,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: local_order
CREATE TABLE IF NOT EXISTS "local_order" (
    "id" INT PRIMARY KEY REFERENCES "order"("id") ON DELETE CASCADE,
    "table" INT NOT NULL,
    "display_order" INT NOT NULL,
    "notes" TEXT
);

-- Table: remote_order
CREATE TABLE IF NOT EXISTS "remote_order" (
    "id" INT PRIMARY KEY REFERENCES "order"("id") ON DELETE CASCADE,
    "address_id" INT NOT NULL REFERENCES "Address"("id") ON DELETE RESTRICT,
    "contact_phone" VARCHAR(20) NOT NULL,
    "delivery_fee" NUMERIC(10,2) NOT NULL DEFAULT 0,
    "estimated_minutes" INT,
    "delivery_notes" TEXT
);

-- Table: order_item
CREATE TABLE IF NOT EXISTS "order_item" (
    "id" SERIAL PRIMARY KEY,
    "order_id" INT NOT NULL REFERENCES "order"("id") ON DELETE CASCADE,
    "item_id" INT NOT NULL REFERENCES "menu_item"("id") ON DELETE RESTRICT,
    "parent_order_id" INT REFERENCES "order_item"("id") ON DELETE CASCADE,
    "quantity" INT NOT NULL DEFAULT 1,
    "unit_price" NUMERIC(10,2) NOT NULL,
    "review" INT
);

-- Table: transaction
CREATE TABLE IF NOT EXISTS "transaction" (
    "id" SERIAL PRIMARY KEY,
    "order_id" INT NOT NULL REFERENCES "order"("id") ON DELETE CASCADE,
    "transaction_type" transaction_type_enum NOT NULL,
    "amount" NUMERIC(10,2) NOT NULL,
    "notes" VARCHAR(100),
    "status" transaction_status_enum NOT NULL DEFAULT 'PENDING',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: receipt
CREATE TABLE IF NOT EXISTS "receipt" (
    "transaction_id" INT PRIMARY KEY REFERENCES "transaction"("id") ON DELETE CASCADE,
    "receipt_code" VARCHAR(100) UNIQUE NOT NULL,
    "receipt" JSONB NOT NULL
);
