-- PostgreSQL baseline corresponding to docs/db.dbml (installed in the current schema).
DO $$ BEGIN CREATE TYPE account_type_enum AS ENUM ('OWNER','ANALYST'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE card_brand_enum AS ENUM ('VISA','MASTERCARD','ELO','HIPERCARD','OTHER'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE discount_type_enum AS ENUM ('percent','fixed_cents'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE menu_style_enum AS ENUM ('CLASSICO','EXPERIENCIA'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE order_status_enum AS ENUM ('pending','preparing','ready','delivered','cancelled'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE order_type_enum AS ENUM ('local','remote'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE transaction_status_enum AS ENUM ('pending','paid','failed','refunded','partially_refunded'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE transaction_type_enum AS ENUM ('credit_card','debit_card','pix','cash'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE employee_position_enum AS ENUM ('WAITER','COOK','CASHIER','BARTENDER','HOST','OTHER'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE restaurant_phone_type_enum AS ENUM ('whatsapp','landline','mobile','other'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE verification_purpose_enum AS ENUM ('register_phone','login_phone','add_phone','claim_cpf','verify_email'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE ingredient_unit_enum AS ENUM ('g','kg','ml','l','un','cs','colher'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE ux_channel_enum AS ENUM ('qr_table','counter','remote_link'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE device_type_enum AS ENUM ('mobile','tablet','desktop'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE ux_event_type_enum AS ENUM ('menu_open','category_view','item_impression','item_detail_open','item_detail_close','item_side_select','cart_item_add','cart_item_remove','cart_qty_change','checkout_open','checkout_abandon','order_submitted'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE price_change_reason_enum AS ENUM ('base_price_update','sale_created','sale_ended'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS "Address" (
    id SERIAL PRIMARY KEY, "CEP" VARCHAR(8) NOT NULL CHECK ("CEP" ~ '^[0-9]{8}$'),
    street VARCHAR(255) NOT NULL, number VARCHAR(20) NOT NULL, apt VARCHAR(50),
    complementary VARCHAR(255), city VARCHAR(255), state VARCHAR(2)
);
CREATE TABLE IF NOT EXISTS account (
    id SERIAL PRIMARY KEY, email VARCHAR(255), email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    password VARCHAR(255), phone VARCHAR(20), phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    name VARCHAR(500), address_id INT REFERENCES "Address"(id) ON DELETE SET NULL,
    cpf VARCHAR(11) CHECK (cpf ~ '^[0-9]{11}$'), cpf_claimed_at TIMESTAMPTZ,
    is_sysadmin BOOLEAN NOT NULL DEFAULT FALSE, two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    last_password_change TIMESTAMPTZ, token_version INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (email IS NOT NULL OR phone IS NOT NULL OR cpf IS NOT NULL),
    CHECK (email IS NULL OR password IS NOT NULL)
);
COMMENT ON COLUMN account.password IS 'Password hash only.';
ALTER TABLE account ADD COLUMN IF NOT EXISTS is_administrator BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE account ADD COLUMN IF NOT EXISTS avatar_url TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_account_cpf ON account(cpf);
CREATE UNIQUE INDEX IF NOT EXISTS idx_account_email ON account(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_account_phone ON account(phone);
CREATE UNIQUE INDEX IF NOT EXISTS account_email_lower_unique ON account(lower(email));

CREATE TABLE IF NOT EXISTS verification_code (
    id BIGSERIAL PRIMARY KEY, account_id INT REFERENCES account(id) ON DELETE CASCADE,
    phone VARCHAR(20), email VARCHAR(255), code_hash VARCHAR(255) NOT NULL,
    purpose verification_purpose_enum NOT NULL, attempts INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL, consumed_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((phone IS NULL) <> (email IS NULL))
);
CREATE INDEX IF NOT EXISTS idx_verification_code_email_purpose ON verification_code(email,purpose);
CREATE INDEX IF NOT EXISTS idx_verification_code_phone_purpose ON verification_code(phone,purpose);
CREATE UNIQUE INDEX IF NOT EXISTS verification_active_email ON verification_code(email,purpose) WHERE email IS NOT NULL AND consumed_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS verification_active_phone ON verification_code(phone,purpose) WHERE phone IS NOT NULL AND consumed_at IS NULL;
CREATE TABLE IF NOT EXISTS refresh_token (
    id BIGSERIAL PRIMARY KEY, account_id INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL, expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ, user_agent VARCHAR(255), ip VARCHAR(45), created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_refresh_token_account_id_expires_at ON refresh_token(account_id,expires_at);

CREATE TABLE IF NOT EXISTS restaurant (
    id SERIAL PRIMARY KEY, name_slug VARCHAR(255) UNIQUE NOT NULL, name VARCHAR(255) NOT NULL,
    description TEXT, loyalty_earn_rate NUMERIC(5,2) NOT NULL DEFAULT 1 CHECK (loyalty_earn_rate >= 0),
    archived_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE restaurant ADD COLUMN IF NOT EXISTS icon_url TEXT;
ALTER TABLE restaurant ADD COLUMN IF NOT EXISTS favicon_url TEXT;
ALTER TABLE restaurant ADD COLUMN IF NOT EXISTS cover_url TEXT;
CREATE TABLE IF NOT EXISTS account_restaurant (
    id BIGSERIAL PRIMARY KEY, account_id INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    restaurant_id INT NOT NULL REFERENCES restaurant(id) ON DELETE CASCADE,
    account_type account_type_enum NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_account_restaurant_account_id_restaurant_id ON account_restaurant(account_id,restaurant_id);
CREATE INDEX IF NOT EXISTS idx_account_restaurant_restaurant_id_account_type ON account_restaurant(restaurant_id,account_type);
CREATE TABLE IF NOT EXISTS physical_restaurant (
    id SERIAL PRIMARY KEY, restaurant_id INT NOT NULL REFERENCES restaurant(id),
    address_id INT NOT NULL REFERENCES "Address"(id), name VARCHAR(255) NOT NULL,
    cnpj VARCHAR(14) NOT NULL CHECK (cnpj ~ '^[0-9]{14}$'),
    owner_cnpj VARCHAR(14) NOT NULL CHECK (owner_cnpj ~ '^[0-9]{14}$'),
    cnpj_verified BOOLEAN NOT NULL DEFAULT FALSE, operating_hours JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE, archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (jsonb_typeof(operating_hours)='array')
);
CREATE UNIQUE INDEX IF NOT EXISTS physical_restaurant_cnpj_unique ON physical_restaurant(cnpj);
ALTER TABLE physical_restaurant ADD COLUMN IF NOT EXISTS delivery_fee_cents BIGINT NOT NULL DEFAULT 0 CHECK (delivery_fee_cents >= 0);
CREATE TABLE IF NOT EXISTS restaurant_phone (
    id SERIAL PRIMARY KEY, physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id),
    phone VARCHAR(20) NOT NULL, phone_type restaurant_phone_type_enum NOT NULL DEFAULT 'mobile'
);
CREATE TABLE IF NOT EXISTS dining_table (
    id SERIAL PRIMARY KEY, physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id),
    number INT NOT NULL CHECK (number > 0), seats INT CHECK (seats > 0),
    display_order INT NOT NULL DEFAULT 0, is_active BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_dining_table_physical_restaurant_id_number ON dining_table(physical_restaurant_id,number);
CREATE TABLE IF NOT EXISTS order_qr (
    id BIGSERIAL PRIMARY KEY, branch_id INT NOT NULL REFERENCES physical_restaurant(id),
    table_id INT REFERENCES dining_table(id), token_hash BYTEA NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS employee (
    id SERIAL PRIMARY KEY, physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id),
    account_id INT NOT NULL REFERENCES account(id), name VARCHAR(500) NOT NULL, cpf VARCHAR(11),
    position employee_position_enum NOT NULL, is_active BOOLEAN NOT NULL DEFAULT TRUE, hired_at DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_employee_physical_restaurant_id ON employee(physical_restaurant_id);
DROP INDEX IF EXISTS idx_employee_account_id;
CREATE UNIQUE INDEX IF NOT EXISTS idx_employee_account_branch ON employee(account_id,physical_restaurant_id);

CREATE TABLE IF NOT EXISTS menu (
    id SERIAL PRIMARY KEY, restaurant_id INT UNIQUE NOT NULL REFERENCES restaurant(id), name VARCHAR(255),
    style menu_style_enum NOT NULL DEFAULT 'CLASSICO', accent_color VARCHAR(7), typography VARCHAR(50),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS category (
    id SERIAL PRIMARY KEY, restaurant_id INT NOT NULL REFERENCES restaurant(id), name VARCHAR(100) NOT NULL,
    display_order INT NOT NULL DEFAULT 0, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS ingredient (
    id SERIAL PRIMARY KEY, restaurant_id INT NOT NULL REFERENCES restaurant(id), name VARCHAR(255) NOT NULL,
    default_unit ingredient_unit_enum
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ingredient_restaurant_id_name ON ingredient(restaurant_id,name);
CREATE TABLE IF NOT EXISTS menu_item (
    id SERIAL PRIMARY KEY, restaurant_id INT NOT NULL REFERENCES restaurant(id), category_id INT NOT NULL REFERENCES category(id),
    name VARCHAR(255) NOT NULL, description TEXT, base_price_cents BIGINT NOT NULL CHECK (base_price_cents >= 0),
    base_cost_cents BIGINT CHECK (base_cost_cents >= 0), image_url TEXT, is_available BOOLEAN NOT NULL DEFAULT TRUE,
    is_combo BOOLEAN NOT NULL DEFAULT FALSE, allows_review BOOLEAN NOT NULL DEFAULT TRUE, archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS item_ingredient (
    id SERIAL PRIMARY KEY, item_id INT NOT NULL REFERENCES menu_item(id), ingredient_id INT NOT NULL REFERENCES ingredient(id),
    quantity NUMERIC(10,3), measure ingredient_unit_enum, is_removable BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_item_ingredient_item_id_ingredient_id ON item_ingredient(item_id,ingredient_id);
CREATE TABLE IF NOT EXISTS combo_item (
    combo_id INT NOT NULL REFERENCES menu_item(id), included_item_id INT NOT NULL REFERENCES menu_item(id),
    quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0), PRIMARY KEY(combo_id,included_item_id), CHECK (combo_id <> included_item_id)
);
CREATE TABLE IF NOT EXISTS item_side (
    id SERIAL PRIMARY KEY, item_id INT NOT NULL REFERENCES menu_item(id), title VARCHAR(255) NOT NULL,
    min_options INT NOT NULL DEFAULT 0 CHECK (min_options >= 0), max_options INT NOT NULL CHECK (max_options >= min_options)
);
CREATE TABLE IF NOT EXISTS item_side_option (
    id SERIAL PRIMARY KEY, item_side_id INT NOT NULL REFERENCES item_side(id), name VARCHAR(255) NOT NULL,
    additional_price_cents BIGINT NOT NULL DEFAULT 0, is_available BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE TABLE IF NOT EXISTS item_sale (
    id SERIAL PRIMARY KEY, item_id INT NOT NULL REFERENCES menu_item(id), discount_type discount_type_enum NOT NULL,
    discount_value BIGINT NOT NULL CHECK (discount_value > 0), starts_at TIMESTAMPTZ NOT NULL, ends_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), CHECK (ends_at > starts_at)
);
CREATE INDEX IF NOT EXISTS idx_item_sale_item_id_is_active ON item_sale(item_id,is_active);
ALTER TABLE menu ADD COLUMN IF NOT EXISTS secondary_color VARCHAR(7);
ALTER TABLE menu ADD COLUMN IF NOT EXISTS display_config JSONB CHECK (jsonb_typeof(display_config)='object');
ALTER TABLE category ADD COLUMN IF NOT EXISTS is_fallback BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE category ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE category ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;
CREATE UNIQUE INDEX IF NOT EXISTS category_one_fallback ON category(restaurant_id) WHERE is_fallback;
ALTER TABLE combo_item ADD COLUMN IF NOT EXISTS id SERIAL;
CREATE UNIQUE INDEX IF NOT EXISTS combo_item_id_unique ON combo_item(id);
ALTER TABLE item_sale ADD COLUMN IF NOT EXISTS branch_id INT REFERENCES physical_restaurant(id);
CREATE TABLE IF NOT EXISTS branch_item_availability (
    branch_id INT NOT NULL REFERENCES physical_restaurant(id), item_id INT NOT NULL REFERENCES menu_item(id),
    is_available BOOLEAN NOT NULL, PRIMARY KEY(branch_id,item_id)
);
CREATE TABLE IF NOT EXISTS branch_option_availability (
    branch_id INT NOT NULL REFERENCES physical_restaurant(id), option_id INT NOT NULL REFERENCES item_side_option(id) ON DELETE CASCADE,
    is_available BOOLEAN NOT NULL, PRIMARY KEY(branch_id,option_id)
);
CREATE TABLE IF NOT EXISTS coupon (
    id SERIAL PRIMARY KEY, restaurant_id INT NOT NULL REFERENCES restaurant(id), code VARCHAR(50) NOT NULL,
    discount_type discount_type_enum NOT NULL, discount_value BIGINT NOT NULL CHECK (discount_value > 0),
    min_order_cents BIGINT NOT NULL DEFAULT 0, max_uses INT, used_count INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE, expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_coupons_restaurant_id_code ON coupon(restaurant_id,code);
CREATE UNIQUE INDEX IF NOT EXISTS coupon_code_folded_unique ON coupon(restaurant_id,upper(code));
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS name VARCHAR(255);
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS title VARCHAR(255);
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS target_item_id INT REFERENCES menu_item(id);
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS target_scope VARCHAR(8) NOT NULL DEFAULT 'order';
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS is_banner BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS background_image_url TEXT;
DO $$ BEGIN
    ALTER TABLE coupon ADD CONSTRAINT coupon_target_scope_check CHECK (
        (target_scope='order' AND target_item_id IS NULL) OR
        (target_scope IN ('item','combo') AND target_item_id IS NOT NULL));
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
CREATE TABLE IF NOT EXISTS coupon_claim (
    account_id INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    coupon_id INT NOT NULL REFERENCES coupon(id), claimed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(account_id,coupon_id)
);

CREATE TABLE IF NOT EXISTS "order" (
    id SERIAL PRIMARY KEY, physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id),
    account_id INT REFERENCES account(id) ON DELETE SET NULL, customer_cpf VARCHAR(11), coupon_id INT REFERENCES coupon(id),
    coupon_code VARCHAR(50), order_type order_type_enum NOT NULL, status order_status_enum NOT NULL DEFAULT 'pending',
    subtotal_cents BIGINT NOT NULL DEFAULT 0, discount_cents BIGINT NOT NULL DEFAULT 0,
    freight_cents BIGINT NOT NULL DEFAULT 0, total_cents BIGINT NOT NULL DEFAULT 0,
    rating SMALLINT CHECK (rating BETWEEN 1 AND 5), feedback TEXT, rated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (account_id IS NULL OR customer_cpf IS NULL)
);
CREATE INDEX IF NOT EXISTS idx_order_account_id_created_at ON "order"(account_id,created_at);
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ DEFAULT now();
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS guest_token_hash BYTEA;
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS qr_id BIGINT REFERENCES order_qr(id);
CREATE TABLE IF NOT EXISTS order_draft_item (
    id SERIAL PRIMARY KEY, order_id INT NOT NULL REFERENCES "order"(id),
    item_id INT NOT NULL REFERENCES menu_item(id), selection JSONB NOT NULL, snapshot JSONB NOT NULL,
    voided_at TIMESTAMPTZ, voided_reason TEXT
);
CREATE INDEX IF NOT EXISTS order_draft_item_order_id ON order_draft_item(order_id,id);
CREATE TABLE IF NOT EXISTS order_promotion (
    order_id INT PRIMARY KEY REFERENCES "order"(id), result JSONB NOT NULL
);
CREATE TABLE IF NOT EXISTS order_coupon (
    order_id INT NOT NULL REFERENCES "order"(id), coupon_id INT NOT NULL REFERENCES coupon(id),
    coupon_code VARCHAR(50) NOT NULL, discount_cents BIGINT NOT NULL CHECK(discount_cents > 0),
    PRIMARY KEY(order_id,coupon_id)
);
CREATE INDEX IF NOT EXISTS idx_order_customer_cpf_created_at ON "order"(customer_cpf,created_at);
CREATE INDEX IF NOT EXISTS idx_order_physical_restaurant_id_status_created_at ON "order"(physical_restaurant_id,status,created_at);
CREATE TABLE IF NOT EXISTS local_order (
    id INT PRIMARY KEY REFERENCES "order"(id), table_id INT REFERENCES dining_table(id), table_number INT,
    display_order INT NOT NULL DEFAULT 0, notes TEXT
);
CREATE TABLE IF NOT EXISTS remote_order (
    id INT PRIMARY KEY REFERENCES "order"(id), contact_phone VARCHAR(20) NOT NULL, estimated_minutes INT,
    delivery_notes TEXT, delivery_cep VARCHAR(8) NOT NULL, delivery_street VARCHAR(255) NOT NULL,
    delivery_number VARCHAR(20) NOT NULL, delivery_apt VARCHAR(50), delivery_complement VARCHAR(255),
    delivery_city VARCHAR(255), delivery_state VARCHAR(2)
);
CREATE TABLE IF NOT EXISTS order_item (
    id SERIAL PRIMARY KEY, order_id INT NOT NULL REFERENCES "order"(id), item_id INT NOT NULL REFERENCES menu_item(id),
    item_name VARCHAR(255) NOT NULL, unit_price_cents BIGINT NOT NULL, unit_cost_cents BIGINT, line_total_cents BIGINT NOT NULL,
    quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0), notes TEXT, selected_options JSONB NOT NULL DEFAULT '[]'::jsonb,
    parent_order_item_id INT REFERENCES order_item(id), voided_at TIMESTAMPTZ, voided_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), CHECK (jsonb_typeof(selected_options)='array')
);
CREATE TABLE IF NOT EXISTS transaction (
    id SERIAL PRIMARY KEY, order_id INT NOT NULL REFERENCES "order"(id), transaction_type transaction_type_enum NOT NULL,
    card_brand card_brand_enum, amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    refunded_cents BIGINT NOT NULL DEFAULT 0 CHECK (refunded_cents >= 0 AND refunded_cents <= amount_cents),
    status transaction_status_enum NOT NULL DEFAULT 'pending', created_by_account_id INT REFERENCES account(id),
    notes VARCHAR(100), created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_transaction_order_id_status ON transaction(order_id,status);
ALTER TABLE transaction ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(128);
ALTER TABLE transaction ADD COLUMN IF NOT EXISTS request_fingerprint TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS transaction_order_idempotency ON transaction(order_id,idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE TABLE IF NOT EXISTS pix_charge (
    id SERIAL PRIMARY KEY, transaction_id INT UNIQUE NOT NULL REFERENCES transaction(id), copy_paste_code TEXT NOT NULL,
    qr_code_base64 TEXT NOT NULL, expires_at TIMESTAMPTZ NOT NULL, paid_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS refund (
    id SERIAL PRIMARY KEY, transaction_id INT NOT NULL REFERENCES transaction(id), amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    reason TEXT, created_by_account_id INT REFERENCES account(id), created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_refund_transaction_id ON refund(transaction_id);
ALTER TABLE refund ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(128);
CREATE UNIQUE INDEX IF NOT EXISTS refund_idempotency ON refund(transaction_id,idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE TABLE IF NOT EXISTS receipt (
    id SERIAL PRIMARY KEY, order_id INT UNIQUE NOT NULL REFERENCES "order"(id), receipt_code VARCHAR(100) UNIQUE NOT NULL,
    receipt JSONB NOT NULL, issued_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS item_review (
    id BIGSERIAL PRIMARY KEY, item_id INT NOT NULL REFERENCES menu_item(id), order_item_id INT UNIQUE NOT NULL REFERENCES order_item(id),
    account_id INT REFERENCES account(id), rating SMALLINT NOT NULL CONSTRAINT item_review_face_rating CHECK (rating BETWEEN 1 AND 3), comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_item_review_item_id_created_at ON item_review(item_id,created_at);
CREATE TABLE IF NOT EXISTS restaurant_review (
    id BIGSERIAL PRIMARY KEY, restaurant_id INT NOT NULL REFERENCES restaurant(id), account_id INT NOT NULL REFERENCES account(id),
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5), comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_restaurant_review_restaurant_id_created_at ON restaurant_review(restaurant_id,created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_restaurant_review_restaurant_id_account_id ON restaurant_review(restaurant_id,account_id);
DO $$ BEGIN
    ALTER TABLE item_review DROP CONSTRAINT IF EXISTS item_review_rating_check;
    DELETE FROM item_review WHERE rating NOT BETWEEN 1 AND 3;
    ALTER TABLE item_review ADD CONSTRAINT item_review_face_rating CHECK (rating BETWEEN 1 AND 3);
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
ALTER TABLE restaurant_review ALTER COLUMN account_id DROP NOT NULL;
ALTER TABLE restaurant_review ADD COLUMN IF NOT EXISTS order_id INT REFERENCES "order"(id);
CREATE UNIQUE INDEX IF NOT EXISTS restaurant_review_purchase_unique ON restaurant_review(order_id) WHERE order_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS order_item_sales_by_item ON order_item(item_id,order_id) INCLUDE (quantity) WHERE voided_at IS NULL;
CREATE TABLE IF NOT EXISTS item_review_totals (
    target_id INT PRIMARY KEY REFERENCES menu_item(id), review_count BIGINT NOT NULL, rating_sum BIGINT NOT NULL
);
CREATE TABLE IF NOT EXISTS restaurant_review_totals (
    target_id INT PRIMARY KEY REFERENCES restaurant(id), review_count BIGINT NOT NULL, rating_sum BIGINT NOT NULL
);
INSERT INTO item_review_totals SELECT item_id,count(*),sum(rating) FROM item_review GROUP BY item_id ON CONFLICT DO NOTHING;
INSERT INTO restaurant_review_totals SELECT restaurant_id,count(*),sum(rating) FROM restaurant_review GROUP BY restaurant_id ON CONFLICT DO NOTHING;
CREATE OR REPLACE FUNCTION update_review_totals() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE target INT; count_delta BIGINT; rating_delta BIGINT;
BEGIN
    IF TG_OP = 'UPDATE' AND (to_jsonb(OLD)->>TG_ARGV[0]) IS DISTINCT FROM (to_jsonb(NEW)->>TG_ARGV[0]) THEN
        RAISE EXCEPTION 'review target cannot change' USING ERRCODE='23514';
    END IF;
    IF TG_OP = 'INSERT' THEN
        target := (to_jsonb(NEW)->>TG_ARGV[0])::int; count_delta := 1; rating_delta := NEW.rating;
    ELSIF TG_OP = 'DELETE' THEN
        target := (to_jsonb(OLD)->>TG_ARGV[0])::int; count_delta := -1; rating_delta := -OLD.rating;
    ELSE
        target := (to_jsonb(NEW)->>TG_ARGV[0])::int; count_delta := 0; rating_delta := NEW.rating - OLD.rating;
    END IF;
    EXECUTE format('INSERT INTO %I(target_id,review_count,rating_sum) VALUES($1,$2,$3)
        ON CONFLICT(target_id) DO UPDATE SET review_count=%I.review_count+EXCLUDED.review_count,
        rating_sum=%I.rating_sum+EXCLUDED.rating_sum', TG_ARGV[1],TG_ARGV[1],TG_ARGV[1]) USING target,count_delta,rating_delta;
    RETURN NULL;
END $$;
DROP TRIGGER IF EXISTS item_review_rollup ON item_review;
CREATE TRIGGER item_review_rollup AFTER INSERT OR UPDATE OR DELETE ON item_review
    FOR EACH ROW EXECUTE FUNCTION update_review_totals('item_id','item_review_totals');
DROP TRIGGER IF EXISTS restaurant_review_rollup ON restaurant_review;
CREATE TRIGGER restaurant_review_rollup AFTER INSERT OR UPDATE OR DELETE ON restaurant_review
    FOR EACH ROW EXECUTE FUNCTION update_review_totals('restaurant_id','restaurant_review_totals');
CREATE TABLE IF NOT EXISTS loyalty_wallet (
    id SERIAL PRIMARY KEY, account_id INT NOT NULL REFERENCES account(id), restaurant_id INT NOT NULL REFERENCES restaurant(id),
    coins_amount BIGINT NOT NULL DEFAULT 0 CHECK (coins_amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_loyalty_wallet_account_restaurant ON loyalty_wallet(account_id,restaurant_id);
CREATE TABLE IF NOT EXISTS wallet_transaction (
    id BIGSERIAL PRIMARY KEY, loyalty_wallet_id INT NOT NULL REFERENCES loyalty_wallet(id), points_delta BIGINT NOT NULL,
    balance_after BIGINT NOT NULL, reason VARCHAR(100) NOT NULL, order_id INT REFERENCES "order"(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_wallet_transaction_order_id ON wallet_transaction(order_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transaction_wallet_id_created_at ON wallet_transaction(loyalty_wallet_id,created_at);

CREATE TABLE IF NOT EXISTS menu_session (
    id BIGSERIAL PRIMARY KEY, restaurant_id INT NOT NULL REFERENCES restaurant(id),
    physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id), dining_table_id INT REFERENCES dining_table(id),
    account_id INT REFERENCES account(id), visitor_fingerprint VARCHAR(128), channel ux_channel_enum NOT NULL,
    device_type device_type_enum NOT NULL, total_duration_ms INT CHECK (total_duration_ms >= 0),
    max_scroll_depth_pct SMALLINT CHECK (max_scroll_depth_pct BETWEEN 0 AND 100), converted_order_id INT REFERENCES "order"(id),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(), last_activity_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_menu_session_branch_started_at ON menu_session(physical_restaurant_id,started_at);
CREATE INDEX IF NOT EXISTS idx_menu_session_converted_order_id ON menu_session(converted_order_id);
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS menu_session_id BIGINT REFERENCES menu_session(id);
CREATE INDEX IF NOT EXISTS idx_menu_session_visitor_fingerprint ON menu_session(visitor_fingerprint);
CREATE TABLE IF NOT EXISTS ux_event (
    id BIGSERIAL PRIMARY KEY, session_id BIGINT NOT NULL REFERENCES menu_session(id), restaurant_id INT NOT NULL REFERENCES restaurant(id),
    physical_restaurant_id INT NOT NULL REFERENCES physical_restaurant(id), event_type ux_event_type_enum NOT NULL,
    category_id INT REFERENCES category(id), item_id INT REFERENCES menu_item(id), duration_ms INT, quantity INT,
    cart_total_cents BIGINT, metadata JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ux_event_branch_created_at ON ux_event(physical_restaurant_id,created_at);
CREATE INDEX IF NOT EXISTS idx_ux_event_item_event_created_at ON ux_event(item_id,event_type,created_at);
CREATE INDEX IF NOT EXISTS idx_ux_event_session_id ON ux_event(session_id);
CREATE INDEX IF NOT EXISTS idx_ux_event_type_created_at ON ux_event(event_type,created_at);
CREATE TABLE IF NOT EXISTS menu_item_price_history (
    id BIGSERIAL PRIMARY KEY, restaurant_id INT NOT NULL REFERENCES restaurant(id), physical_restaurant_id INT REFERENCES physical_restaurant(id),
    item_id INT NOT NULL REFERENCES menu_item(id), changed_by_account_id INT REFERENCES account(id), old_price_cents BIGINT NOT NULL,
    new_price_cents BIGINT NOT NULL, reason price_change_reason_enum NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_price_history_item_id_created_at ON menu_item_price_history(item_id,created_at);
CREATE INDEX IF NOT EXISTS idx_price_history_reason_created_at ON menu_item_price_history(reason,created_at);
CREATE INDEX IF NOT EXISTS idx_price_history_restaurant_id_created_at ON menu_item_price_history(restaurant_id,created_at);
CREATE TABLE IF NOT EXISTS audit_entry (
    id BIGSERIAL PRIMARY KEY, actor_account_id INT REFERENCES account(id), actor_role VARCHAR(20), restaurant_id INT REFERENCES restaurant(id),
    branch_id INT REFERENCES physical_restaurant(id), action VARCHAR(100) NOT NULL, entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100) NOT NULL, metadata JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_entry_action_created_at ON audit_entry(action,created_at);
CREATE INDEX IF NOT EXISTS idx_audit_entry_actor_account_id_created_at ON audit_entry(actor_account_id,created_at);
CREATE INDEX IF NOT EXISTS idx_audit_entry_entity_type_entity_id_created_at ON audit_entry(entity_type,entity_id,created_at);
CREATE INDEX IF NOT EXISTS idx_audit_entry_restaurant_id_created_at ON audit_entry(restaurant_id,created_at);

-- Runtime extensions for scoped branch management and table archival.
CREATE TABLE IF NOT EXISTS branch_manager (
    branch_id INT NOT NULL REFERENCES physical_restaurant(id) ON DELETE CASCADE,
    account_id INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    PRIMARY KEY(branch_id,account_id)
);
-- Existing organization managers retain operational access to their current
-- branches, but are never promoted into a global administrator.
INSERT INTO branch_manager(branch_id,account_id)
SELECT b.id,m.account_id FROM account_restaurant m JOIN physical_restaurant b ON b.restaurant_id=m.restaurant_id
WHERE m.account_type::text='MANAGER' AND b.archived_at IS NULL ON CONFLICT DO NOTHING;
DELETE FROM account_restaurant WHERE account_type::text='MANAGER';
DO $$ BEGIN
    ALTER TABLE account_restaurant ADD CONSTRAINT organization_membership_role CHECK (account_type::text IN ('OWNER','ANALYST'));
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
ALTER TABLE dining_table ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN NEW.updated_at := now(); RETURN NEW; END $$;
DO $$ DECLARE t TEXT; BEGIN
    FOREACH t IN ARRAY ARRAY['account','restaurant','physical_restaurant','employee','menu','menu_item',
        'coupon','order','transaction','restaurant_review','loyalty_wallet'] LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS touch_updated_at ON %I',t);
        EXECUTE format('CREATE TRIGGER touch_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION set_updated_at()',t);
    END LOOP;
END $$;
