-- Additive compatibility for databases created by the older server schema.
-- This is executed in the same transaction as schema.sql.
DO $$
BEGIN
    IF to_regclass('account') IS NOT NULL THEN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'account' AND column_name = 'password') THEN
            ALTER TABLE account ALTER COLUMN email DROP NOT NULL;
            ALTER TABLE account ALTER COLUMN password DROP NOT NULL;
            IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'account' AND column_name = 'password_hash') THEN
                ALTER TABLE account ADD COLUMN password_hash VARCHAR(255);
                UPDATE account SET password_hash = password;
            END IF;
            IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'account' AND column_name = 'email_verified') THEN
                ALTER TABLE account ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE;
                UPDATE account SET email_verified = TRUE WHERE email IS NOT NULL AND password_hash IS NOT NULL;
            END IF;
        END IF;
        ALTER TABLE account ADD COLUMN IF NOT EXISTS phone_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE account ADD COLUMN IF NOT EXISTS token_version INT NOT NULL DEFAULT 0;
        ALTER TABLE account ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
    END IF;
END $$;

-- Keep the old password column usable by legacy code without resurrecting a
-- removed password on a later migration. New code uses password_hash.
CREATE OR REPLACE FUNCTION sync_legacy_account_password() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.password_hash IS NULL THEN NEW.password_hash := NEW.password;
        ELSE NEW.password := NEW.password_hash; END IF;
    ELSIF NEW.password_hash IS DISTINCT FROM OLD.password_hash THEN
        NEW.password := NEW.password_hash;
    ELSIF NEW.password IS DISTINCT FROM OLD.password THEN
        NEW.password_hash := NEW.password;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DO $$
BEGIN
    IF to_regclass('account') IS NOT NULL AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'account' AND column_name = 'password') THEN
        DROP TRIGGER IF EXISTS trg_sync_legacy_account_password ON account;
        CREATE TRIGGER trg_sync_legacy_account_password BEFORE INSERT OR UPDATE ON account
            FOR EACH ROW EXECUTE FUNCTION sync_legacy_account_password();
    END IF;
END $$;

ALTER TABLE IF EXISTS category ADD COLUMN IF NOT EXISTS display_order INT NOT NULL DEFAULT 0;
ALTER TABLE IF EXISTS restaurant ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE IF EXISTS menu_item ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;
ALTER TABLE IF EXISTS menu_item ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE IF EXISTS coupon ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE IF EXISTS physical_restaurant ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;
ALTER TABLE IF EXISTS physical_restaurant ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE IF EXISTS employee ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE IF EXISTS employee ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE IF EXISTS "order" ADD COLUMN IF NOT EXISTS account_id INT REFERENCES account(id) ON DELETE SET NULL;
ALTER TABLE IF EXISTS "order" ADD COLUMN IF NOT EXISTS customer_cpf VARCHAR(11);
ALTER TABLE IF EXISTS "order" ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE IF EXISTS transaction ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(64);
ALTER TABLE IF EXISTS transaction ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
