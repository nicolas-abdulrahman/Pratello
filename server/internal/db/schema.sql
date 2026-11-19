-- =============================================================
-- Pratello Database Schema — matches OpenAPI v2.1
-- Rebuild, not a migration.
-- =============================================================

-- -------------------------------------------------------------
-- Enums
-- -------------------------------------------------------------
DO $$ BEGIN
    CREATE TYPE discount_type_enum AS ENUM ('percent', 'fixed_cents');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE order_status_enum AS ENUM (
        'pending', 'preparing', 'ready', 'delivered', 'cancelled'
    );
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE order_type_enum AS ENUM ('local', 'remote');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE transaction_type_enum AS ENUM (
        'credit_card', 'debit_card', 'pix', 'cash'
    );
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE transaction_status_enum AS ENUM (
        'pending', 'paid', 'failed', 'refunded', 'partially_refunded'
    );
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE card_brand_enum AS ENUM (
        'VISA', 'MASTERCARD', 'ELO', 'AMEX', 'HIPERCARD', 'OTHER'
    );
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE membership_role_enum AS ENUM ('OWNER', 'MANAGER');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE employee_position_enum AS ENUM (
        'WAITER', 'COOK', 'CASHIER', 'BARTENDER', 'HOST', 'OTHER'
    );
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE branch_phone_type_enum AS ENUM (
        'whatsapp', 'landline', 'mobile', 'other'
    );
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE verification_purpose_enum AS ENUM (
        'register_phone', 'login_phone', 'add_phone', 'claim_loyalty', 'verify_email'
    );
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
    CREATE TYPE ingredient_unit_enum AS ENUM (
        'g', 'kg', 'ml', 'l', 'un', 'cs', 'colher'
    );
EXCEPTION WHEN duplicate_object THEN null; END $$;

-- -------------------------------------------------------------
-- Utility: updated_at trigger
-- -------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- -------------------------------------------------------------
-- Address — one row per account OR per branch.
-- Order delivery addresses are snapshotted, never referenced.
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS address (
    id          SERIAL PRIMARY KEY,
    cep         VARCHAR(8)  NOT NULL,
    street      VARCHAR(255) NOT NULL,
    number      VARCHAR(20) NOT NULL,
    apt         VARCHAR(50),
    complement  VARCHAR(255),
    city        VARCHAR(255),
    state       VARCHAR(2),
    CONSTRAINT chk_address_cep CHECK (cep ~ '^[0-9]{8}$')
);

-- -------------------------------------------------------------
-- Account + auth
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS account (
    id                   SERIAL PRIMARY KEY,
    email                VARCHAR(255) UNIQUE,
    email_verified       BOOLEAN NOT NULL DEFAULT FALSE,
    password_hash        VARCHAR(255),
    phone                VARCHAR(20) UNIQUE,
    phone_verified       BOOLEAN NOT NULL DEFAULT FALSE,
    name                 VARCHAR(255),
    address_id           INT REFERENCES address(id) ON DELETE SET NULL,
    two_factor_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
    last_password_change TIMESTAMPTZ,
    token_version        INT NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- At least one auth method must exist.
    CONSTRAINT chk_account_has_method CHECK (email IS NOT NULL OR phone IS NOT NULL),
    -- If an email exists, a password must exist too.
    CONSTRAINT chk_account_email_password CHECK (email IS NULL OR password_hash IS NOT NULL)
);
CREATE TRIGGER trg_account_updated BEFORE UPDATE ON account
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS verification_code (
    id          BIGSERIAL PRIMARY KEY,
    account_id  INT REFERENCES account(id) ON DELETE CASCADE,
    phone       VARCHAR(20),
    email       VARCHAR(255),
    code_hash   VARCHAR(255) NOT NULL,
    purpose     verification_purpose_enum NOT NULL,
    attempts    INT NOT NULL DEFAULT 0,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Exactly one contact target, matched to purpose.
    CONSTRAINT chk_verification_target CHECK (
        (phone IS NOT NULL AND email IS NULL)
        OR (phone IS NULL AND email IS NOT NULL)
    )
);
CREATE INDEX IF NOT EXISTS verification_code_phone
    ON verification_code (phone, purpose) WHERE consumed_at IS NULL;
CREATE INDEX IF NOT EXISTS verification_code_email
    ON verification_code (email, purpose) WHERE consumed_at IS NULL;

CREATE TABLE IF NOT EXISTS refresh_token (
    id          BIGSERIAL PRIMARY KEY,
    account_id  INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) UNIQUE NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    user_agent  VARCHAR(255),
    ip          VARCHAR(45),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS refresh_token_account
    ON refresh_token (account_id, expires_at);

-- -------------------------------------------------------------
-- Sysadmin
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sysadmin (
    id          SERIAL PRIMARY KEY,
    account_id  INT NOT NULL UNIQUE REFERENCES account(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -------------------------------------------------------------
-- Restaurant (logical) + memberships + menu layout (inline)
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS restaurant (
    id                   SERIAL PRIMARY KEY,
    slug                 VARCHAR(255) UNIQUE NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    description          TEXT,
    -- Menu layout lives here: 1 per restaurant.
    menu_name            VARCHAR(255),
    menu_template_style  VARCHAR(50) NOT NULL DEFAULT 'classic',
    menu_accent_color    VARCHAR(7),
    menu_typography      VARCHAR(50),
    archived_at          TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_restaurant_accent_color
        CHECK (menu_accent_color IS NULL OR menu_accent_color ~ '^#[0-9A-Fa-f]{6}$')
);
CREATE TRIGGER trg_restaurant_updated BEFORE UPDATE ON restaurant
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS restaurant_membership (
    id            SERIAL PRIMARY KEY,
    account_id    INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    restaurant_id INT NOT NULL REFERENCES restaurant(id) ON DELETE CASCADE,
    role          membership_role_enum NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (account_id, restaurant_id)
);
CREATE INDEX IF NOT EXISTS restaurant_membership_restaurant
    ON restaurant_membership (restaurant_id, role);

-- -------------------------------------------------------------
-- Catalog
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS category (
    id             SERIAL PRIMARY KEY,
    restaurant_id  INT NOT NULL REFERENCES restaurant(id) ON DELETE CASCADE,
    name           VARCHAR(100) NOT NULL,
    display_order  INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS category_restaurant
    ON category (restaurant_id, display_order);

CREATE TABLE IF NOT EXISTS ingredient (
    id             SERIAL PRIMARY KEY,
    restaurant_id  INT NOT NULL REFERENCES restaurant(id) ON DELETE CASCADE,
    name           VARCHAR(255) NOT NULL,
    default_unit   ingredient_unit_enum,
    UNIQUE (restaurant_id, name)
);

CREATE TABLE IF NOT EXISTS menu_item (
    id               SERIAL PRIMARY KEY,
    restaurant_id    INT NOT NULL REFERENCES restaurant(id) ON DELETE CASCADE,
    category_id      INT NOT NULL REFERENCES category(id) ON DELETE RESTRICT,
    name             VARCHAR(255) NOT NULL,
    description      TEXT,
    base_price_cents BIGINT NOT NULL CHECK (base_price_cents >= 0),
    photo_url        TEXT,
    is_available     BOOLEAN NOT NULL DEFAULT TRUE,
    is_combo         BOOLEAN NOT NULL DEFAULT FALSE,
    allows_review    BOOLEAN NOT NULL DEFAULT TRUE,
    archived_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS menu_item_restaurant_category
    ON menu_item (restaurant_id, category_id) WHERE archived_at IS NULL;
CREATE TRIGGER trg_menu_item_updated BEFORE UPDATE ON menu_item
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS item_ingredient (
    id              SERIAL PRIMARY KEY,
    item_id         INT NOT NULL REFERENCES menu_item(id) ON DELETE CASCADE,
    ingredient_id   INT NOT NULL REFERENCES ingredient(id) ON DELETE RESTRICT,
    quantity        NUMERIC(10,3),
    unit            ingredient_unit_enum,
    is_removable    BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (item_id, ingredient_id)
);

CREATE TABLE IF NOT EXISTS side_group (
    id           SERIAL PRIMARY KEY,
    item_id      INT NOT NULL REFERENCES menu_item(id) ON DELETE CASCADE,
    title        VARCHAR(255) NOT NULL,
    min_options  INT NOT NULL DEFAULT 0 CHECK (min_options >= 0),
    max_options  INT NOT NULL CHECK (max_options >= min_options)
);

CREATE TABLE IF NOT EXISTS side_option (
    id                     SERIAL PRIMARY KEY,
    side_group_id          INT NOT NULL REFERENCES side_group(id) ON DELETE CASCADE,
    name                   VARCHAR(255) NOT NULL,
    additional_price_cents BIGINT NOT NULL DEFAULT 0 CHECK (additional_price_cents >= 0),
    is_available           BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS combo_item (
    combo_id          INT NOT NULL REFERENCES menu_item(id) ON DELETE CASCADE,
    included_item_id  INT NOT NULL REFERENCES menu_item(id) ON DELETE CASCADE,
    quantity          INT NOT NULL DEFAULT 1 CHECK (quantity >= 1),
    PRIMARY KEY (combo_id, included_item_id),
    CONSTRAINT chk_combo_not_self CHECK (combo_id <> included_item_id)
);

CREATE TABLE IF NOT EXISTS item_sale (
    id              SERIAL PRIMARY KEY,
    item_id         INT NOT NULL REFERENCES menu_item(id) ON DELETE CASCADE,
    discount_type   discount_type_enum NOT NULL,
    discount_value  BIGINT NOT NULL CHECK (discount_value > 0),
    starts_at       TIMESTAMPTZ NOT NULL,
    ends_at         TIMESTAMPTZ NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_sale_window CHECK (ends_at > starts_at)
);
CREATE INDEX IF NOT EXISTS item_sale_active
    ON item_sale (item_id, is_active);

-- -------------------------------------------------------------
-- Coupons
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS coupon (
    id               SERIAL PRIMARY KEY,
    restaurant_id    INT NOT NULL REFERENCES restaurant(id) ON DELETE CASCADE,
    code             VARCHAR(50) NOT NULL,
    discount_type    discount_type_enum NOT NULL,
    discount_value   BIGINT NOT NULL CHECK (discount_value > 0),
    min_order_cents  BIGINT NOT NULL DEFAULT 0 CHECK (min_order_cents >= 0),
    max_uses         INT,
    used_count       INT NOT NULL DEFAULT 0,
    expires_at       TIMESTAMPTZ,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (restaurant_id, code)
);
CREATE TRIGGER trg_coupon_updated BEFORE UPDATE ON coupon
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -------------------------------------------------------------
-- Branch (physical restaurant)
-- operating_hours stored inline as JSONB:
--   [{"weekday":0,"opens_at":"11:00","closes_at":"22:00","is_closed":false}, ...]
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS physical_restaurant (
    id                  SERIAL PRIMARY KEY,
    restaurant_id       INT NOT NULL REFERENCES restaurant(id) ON DELETE CASCADE,
    address_id          INT NOT NULL REFERENCES address(id) ON DELETE RESTRICT,
    name                VARCHAR(255) NOT NULL,
    cnpj                VARCHAR(14) NOT NULL,
    owner_cnpj          VARCHAR(14) NOT NULL,
    cnpj_verified       BOOLEAN NOT NULL DEFAULT FALSE,
    terminal_pin_hash   VARCHAR(255),
    pin_updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    failed_pin_attempts INT NOT NULL DEFAULT 0,
    locked_until        TIMESTAMPTZ,
    operating_hours     JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    archived_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_physical_cnpj       CHECK (cnpj ~ '^[0-9]{14}$'),
    CONSTRAINT chk_physical_owner_cnpj CHECK (owner_cnpj ~ '^[0-9]{14}$'),
    CONSTRAINT chk_operating_hours_is_array CHECK (jsonb_typeof(operating_hours) = 'array')
);
CREATE INDEX IF NOT EXISTS physical_restaurant_restaurant
    ON physical_restaurant (restaurant_id) WHERE archived_at IS NULL;
CREATE TRIGGER trg_physical_updated BEFORE UPDATE ON physical_restaurant
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS branch_phone (
    id                     SERIAL PRIMARY KEY,
    physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id) ON DELETE CASCADE,
    phone_number           VARCHAR(20) NOT NULL,
    phone_type             branch_phone_type_enum NOT NULL DEFAULT 'mobile'
);

CREATE TABLE IF NOT EXISTS dining_table (
    id                     SERIAL PRIMARY KEY,
    physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id) ON DELETE CASCADE,
    number                 INT NOT NULL CHECK (number >= 1),
    seats                  INT,
    display_order          INT NOT NULL DEFAULT 0,
    is_active              BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (physical_restaurant_id, number)
);

-- -------------------------------------------------------------
-- Employee + sessions
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS employee (
    id                     SERIAL PRIMARY KEY,
    physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id) ON DELETE CASCADE,
    name                   VARCHAR(255) NOT NULL,
    cpf                    VARCHAR(11),
    position               employee_position_enum NOT NULL,
    account_id             INT REFERENCES account(id) ON DELETE SET NULL,
    is_active              BOOLEAN NOT NULL DEFAULT TRUE,
    hired_at               DATE,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS employee_branch
    ON employee (physical_restaurant_id) WHERE is_active = TRUE;
CREATE TRIGGER trg_employee_updated BEFORE UPDATE ON employee
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS terminal_session (
    id                     SERIAL PRIMARY KEY,
    physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id) ON DELETE CASCADE,
    jti                    UUID UNIQUE NOT NULL,
    device_id              VARCHAR(255),
    expires_at             TIMESTAMPTZ NOT NULL,
    revoked_at             TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS terminal_session_branch
    ON terminal_session (physical_restaurant_id, expires_at);

CREATE TABLE IF NOT EXISTS staff_session (
    id                     SERIAL PRIMARY KEY,
    physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id) ON DELETE CASCADE,
    shift_code             VARCHAR(50) NOT NULL,
    qr_token_hash          VARCHAR(255) UNIQUE NOT NULL,
    is_active              BOOLEAN NOT NULL DEFAULT TRUE,
    opened_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at              TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS staff_session_branch_active
    ON staff_session (physical_restaurant_id) WHERE is_active = TRUE;

CREATE TABLE IF NOT EXISTS staff_checkin (
    id               BIGSERIAL PRIMARY KEY,
    staff_session_id INT NOT NULL REFERENCES staff_session(id) ON DELETE CASCADE,
    account_id       INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    checked_in_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    checked_out_at   TIMESTAMPTZ,
    UNIQUE (staff_session_id, account_id, checked_in_at)
);

-- -------------------------------------------------------------
-- Loyalty
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS loyalty_customer (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpf         VARCHAR(11) UNIQUE NOT NULL,
    name        VARCHAR(255),
    phone       VARCHAR(20),
    account_id  INT REFERENCES account(id) ON DELETE SET NULL,
    claimed_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_loyalty_cpf CHECK (cpf ~ '^[0-9]{11}$')
);
CREATE INDEX IF NOT EXISTS loyalty_customer_account
    ON loyalty_customer (account_id) WHERE account_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS loyalty_wallet (
    id                  SERIAL PRIMARY KEY,
    loyalty_customer_id UUID NOT NULL REFERENCES loyalty_customer(id) ON DELETE CASCADE,
    restaurant_id       INT NOT NULL REFERENCES restaurant(id) ON DELETE CASCADE,
    points_balance      BIGINT NOT NULL DEFAULT 0 CHECK (points_balance >= 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (loyalty_customer_id, restaurant_id)
);
CREATE TRIGGER trg_loyalty_wallet_updated BEFORE UPDATE ON loyalty_wallet
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -------------------------------------------------------------
-- Orders
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS "order" (
    id                     SERIAL PRIMARY KEY,
    physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id) ON DELETE RESTRICT,
    account_id             INT REFERENCES account(id) ON DELETE SET NULL,
    customer_cpf           VARCHAR(11),
    coupon_id              INT REFERENCES coupon(id) ON DELETE SET NULL,
    coupon_code            VARCHAR(50),
    order_type             order_type_enum NOT NULL,
    status                 order_status_enum NOT NULL DEFAULT 'pending',
    subtotal_cents         BIGINT NOT NULL DEFAULT 0 CHECK (subtotal_cents >= 0),
    discount_cents         BIGINT NOT NULL DEFAULT 0 CHECK (discount_cents >= 0),
    freight_cents          BIGINT NOT NULL DEFAULT 0 CHECK (freight_cents >= 0),
    total_cents            BIGINT NOT NULL DEFAULT 0 CHECK (total_cents >= 0),
    rating                 SMALLINT CHECK (rating BETWEEN 1 AND 5),
    feedback               TEXT,
    rated_at               TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_order_single_customer
        CHECK (account_id IS NULL OR customer_cpf IS NULL)
);
CREATE INDEX IF NOT EXISTS order_branch_status
    ON "order" (physical_restaurant_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS order_account
    ON "order" (account_id, created_at DESC) WHERE account_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS order_customer_cpf
    ON "order" (customer_cpf, created_at DESC) WHERE customer_cpf IS NOT NULL;
CREATE TRIGGER trg_order_updated BEFORE UPDATE ON "order"
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS local_order (
    order_id      INT PRIMARY KEY REFERENCES "order"(id) ON DELETE CASCADE,
    table_id      INT NOT NULL REFERENCES dining_table(id) ON DELETE RESTRICT,
    table_number  INT NOT NULL,
    display_order INT NOT NULL DEFAULT 0
);

-- Delivery address is a SNAPSHOT at order time, not a FK to `address`.
CREATE TABLE IF NOT EXISTS remote_order (
    order_id            INT PRIMARY KEY REFERENCES "order"(id) ON DELETE CASCADE,
    contact_phone       VARCHAR(20) NOT NULL,
    estimated_minutes   INT,
    delivery_notes      TEXT,
    delivery_cep        VARCHAR(8) NOT NULL,
    delivery_street     VARCHAR(255) NOT NULL,
    delivery_number     VARCHAR(20) NOT NULL,
    delivery_apt        VARCHAR(50),
    delivery_complement VARCHAR(255),
    delivery_city       VARCHAR(255),
    delivery_state      VARCHAR(2)
);

-- selected_options is JSONB:
--   [{"side_id":1,"side_title":"...","option_id":5,"option_name":"...",
--     "additional_price_cents":0,"quantity":1}]
CREATE TABLE IF NOT EXISTS order_item (
    id                   SERIAL PRIMARY KEY,
    order_id             INT NOT NULL REFERENCES "order"(id) ON DELETE CASCADE,
    item_id              INT NOT NULL REFERENCES menu_item(id) ON DELETE RESTRICT,
    item_name            VARCHAR(255) NOT NULL,
    unit_price_cents     BIGINT NOT NULL,
    quantity             INT NOT NULL DEFAULT 1 CHECK (quantity >= 1),
    line_total_cents     BIGINT NOT NULL,
    notes                TEXT,
    selected_options     JSONB NOT NULL DEFAULT '[]'::jsonb,
    parent_order_item_id INT REFERENCES order_item(id) ON DELETE CASCADE,
    voided_at            TIMESTAMPTZ,
    voided_reason        TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_selected_options_is_array
        CHECK (jsonb_typeof(selected_options) = 'array')
);
CREATE INDEX IF NOT EXISTS order_item_order ON order_item (order_id);

-- -------------------------------------------------------------
-- Reviews
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS item_review (
    id            BIGSERIAL PRIMARY KEY,
    item_id       INT NOT NULL REFERENCES menu_item(id) ON DELETE CASCADE,
    order_item_id INT NOT NULL REFERENCES order_item(id) ON DELETE CASCADE,
    account_id    INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    rating        SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment       TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (order_item_id)
);
CREATE INDEX IF NOT EXISTS item_review_item
    ON item_review (item_id, created_at DESC);

CREATE TABLE IF NOT EXISTS restaurant_review (
    id            BIGSERIAL PRIMARY KEY,
    restaurant_id INT NOT NULL REFERENCES restaurant(id) ON DELETE CASCADE,
    account_id    INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    rating        SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment       TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (restaurant_id, account_id)
);
CREATE INDEX IF NOT EXISTS restaurant_review_restaurant
    ON restaurant_review (restaurant_id, created_at DESC);
CREATE TRIGGER trg_restaurant_review_updated BEFORE UPDATE ON restaurant_review
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -------------------------------------------------------------
-- Payments
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS transaction (
    id                SERIAL PRIMARY KEY,
    order_id          INT NOT NULL REFERENCES "order"(id) ON DELETE CASCADE,
    transaction_type  transaction_type_enum NOT NULL,
    card_brand        card_brand_enum,
    amount_cents      BIGINT NOT NULL CHECK (amount_cents > 0),
    refunded_cents    BIGINT NOT NULL DEFAULT 0 CHECK (refunded_cents >= 0),
    status            transaction_status_enum NOT NULL DEFAULT 'pending',
    idempotency_key   VARCHAR(64),
    notes             VARCHAR(255),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_refund_lte_amount CHECK (refunded_cents <= amount_cents)
);
CREATE INDEX IF NOT EXISTS transaction_order
    ON transaction (order_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS transaction_idempotency
    ON transaction (idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE TRIGGER trg_transaction_updated BEFORE UPDATE ON transaction
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS pix_charge (
    id              SERIAL PRIMARY KEY,
    transaction_id  INT UNIQUE NOT NULL REFERENCES transaction(id) ON DELETE CASCADE,
    copy_paste_code TEXT NOT NULL,
    qr_code_base64  TEXT NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refund (
    id              SERIAL PRIMARY KEY,
    transaction_id  INT NOT NULL REFERENCES transaction(id) ON DELETE CASCADE,
    amount_cents    BIGINT NOT NULL CHECK (amount_cents > 0),
    reason          TEXT,
    idempotency_key VARCHAR(64) UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS refund_transaction ON refund (transaction_id);

CREATE TABLE IF NOT EXISTS receipt (
    id           SERIAL PRIMARY KEY,
    order_id     INT UNIQUE NOT NULL REFERENCES "order"(id) ON DELETE CASCADE,
    receipt_code VARCHAR(100) UNIQUE NOT NULL,
    receipt      JSONB NOT NULL,
    issued_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -------------------------------------------------------------
-- Wallet ledger
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS wallet_transaction (
    id              BIGSERIAL PRIMARY KEY,
    wallet_id       INT NOT NULL REFERENCES loyalty_wallet(id) ON DELETE CASCADE,
    points_delta    BIGINT NOT NULL,
    balance_after   BIGINT NOT NULL,
    reason          VARCHAR(255) NOT NULL,
    order_id        INT REFERENCES "order"(id) ON DELETE SET NULL,
    idempotency_key VARCHAR(64),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS wallet_transaction_wallet
    ON wallet_transaction (wallet_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS wallet_transaction_idempotency
    ON wallet_transaction (wallet_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

-- -------------------------------------------------------------
-- Audit trail
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_entry (
    id               BIGSERIAL PRIMARY KEY,
    actor_account_id INT REFERENCES account(id) ON DELETE SET NULL,
    actor_role       VARCHAR(20),
    restaurant_id    INT REFERENCES restaurant(id) ON DELETE SET NULL,
    branch_id        INT REFERENCES physical_restaurant(id) ON DELETE SET NULL,
    action           VARCHAR(100) NOT NULL,
    entity_type      VARCHAR(50) NOT NULL,
    entity_id        VARCHAR(100) NOT NULL,
    metadata         JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS audit_restaurant
    ON audit_entry (restaurant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS audit_entity
    ON audit_entry (entity_type, entity_id, created_at DESC);
