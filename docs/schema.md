# Pratello Database Schema & Domain Architecture

> **Target Audience:** LLMs, Autonomous Coding Agents, and Developers.  
> **Purpose:** High-density, unambiguous reference for database schema, relational keys, invariants, and business domain logic.

---

## 1. Domain Map & Relational Graph

```mermaid
erDiagram
    account ||--o| sysadmin : "is"
    account ||--o{ manager : "acts_as"
    restaurant ||--o{ manager : "managed_by"
    restaurant ||--o{ physical_restaurant : "operates"
    restaurant ||--o{ menu : "owns"
    restaurant ||--o{ category : "has"
    restaurant ||--o{ Ingredient_item : "catalogs"
    restaurant ||--o{ coupon : "issues"
    restaurant ||--o{ user_restaurant : "loyalty_for"

    physical_restaurant ||--o| Address : "located_at"
    physical_restaurant ||--o{ p_restaurant_phone : "has_phones"
    physical_restaurant ||--o{ order : "receives"
    physical_restaurant ||--o{ restaurant_staff_session : "hosts_shifts"

    category ||--o{ menu_item : "contains"
    menu_item ||--o{ item_ingredient : "composed_of"
    Ingredient_item ||--o{ item_ingredient : "used_in"
    menu_item ||--o{ item_side : "offers_groups"
    item_side ||--o{ side_option : "has_options"
    menu_item ||--o{ combo_item : "parent_bundle"
    menu_item ||--o{ combo_item : "sub_product"
    menu_item ||--o{ item_sale : "has_promos"

    users ||--o{ user_restaurant : "accumulates_points"
    users ||--o{ order : "places"

    order ||--o| local_order : "specializes_dine_in"
    order ||--o| remote_order : "specializes_delivery"
    order ||--o{ order_item : "contains"
    order ||--o{ transaction : "paid_via"
    order ||--o| receipt : "generates"

    transaction ||--o| transaction_pix : "pix_details"
    restaurant_staff_session ||--o{ staff_session : "attendance_logs"
    account ||--o{ staff_session : "worker"
```

---

## 2. Table Schemas (28 Entities)

### Group A: Identity, Access & Profiles

| Table | Columns (PK, FK, Constraints) | Invariants & Business Rules |
| :--- | :--- | :--- |
| `account` | `id` (PK, INT/UUID)<br>`email` (UNIQUE, VARCHAR)<br>`password_hash` (TEXT)<br>`phone` (VARCHAR)<br>`created_at` (TIMESTAMP) | Global credentials for sysadmins, managers, and staff. Never stores plain passwords. |
| `sysadmin` | `id` (PK)<br>`account_id` (FK -> `account.id`, UNIQUE) | Superuser with global platform access across all tenants/restaurants. |
| `manager` | `id` (PK)<br>`restaurant_slug` (FK -> `restaurant.slug`)<br>`account_id` (FK -> `account.id`)<br>UNIQUE(`restaurant_slug`, `account_id`) | Scopes an account's management permissions to a specific restaurant tenant. |
| `users` | `cpf` (PK, CHAR(11))<br>`name` (VARCHAR)<br>`phone` (VARCHAR)<br>`created_at` (TIMESTAMP) | End consumers. Identified by Brazilian CPF for loyalty point tracking and history across orders. |

### Group B: Tenants & Physical Outlets

| Table | Columns (PK, FK, Constraints) | Invariants & Business Rules |
| :--- | :--- | :--- |
| `restaurant` | `slug` (PK, VARCHAR)<br>`name` (VARCHAR)<br>`description` (TEXT)<br>`created_at` (TIMESTAMP) | Top-level tenant. Owner of menus, products, coupons, and global branding. |
| `physical_restaurant`| `id` (PK)<br>`restaurant_slug` (FK -> `restaurant.slug`)<br>`name` (VARCHAR)<br>`address_id` (FK -> `Address.id`) | Physical branch/store of a brand. Processes orders, delivery dispatch, and work shifts. |
| `Address` | `id` (PK)<br>`cep` (CHAR(8))<br>`street` (VARCHAR)<br>`number` (VARCHAR)<br>`apt` (VARCHAR, NULLABLE)<br>`complement` (VARCHAR, NULLABLE) | Shared postal address entity for branches and delivery destinations. |
| `p_restaurant_phone` | `id` (PK)<br>`branch_id` (FK -> `physical_restaurant.id`)<br>`phone_number` (VARCHAR)<br>`phone_type` (VARCHAR) | Branch contact numbers (e.g. `whatsapp`, `landline`, `customer_service`). |

### Group C: Menu, Catalog & Composition

| Table | Columns (PK, FK, Constraints) | Invariants & Business Rules |
| :--- | :--- | :--- |
| `menu` | `id` (PK)<br>`restaurant_slug` (FK -> `restaurant.slug`, UNIQUE)<br>`name` (VARCHAR)<br>`template_style` (VARCHAR)<br>`accent_color` (CHAR(7))<br>`typography` (VARCHAR) | UI theme & visual layout settings for the interactive digital menu (QR / web). |
| `category` | `id` (PK)<br>`restaurant_slug` (FK -> `restaurant.slug`)<br>`name` (VARCHAR)<br>`display_order` (INT) | Menu visual grouping (e.g., "Starters", "Beverages", "Burgers"). Sorted by `display_order`. |
| `menu_item` | `id` (PK)<br>`category_id` (FK -> `category.id`)<br>`name` (VARCHAR)<br>`description` (TEXT)<br>`base_price` (DECIMAL(10,2))<br>`photo_url` (TEXT)<br>`is_combo` (BOOLEAN)<br>`is_available` (BOOLEAN) | Sellable catalog item. If `is_combo = TRUE`, recipes live in `combo_item`. |
| `Ingredient_item` | `id` (PK)<br>`restaurant_slug` (FK -> `restaurant.slug`)<br>`name` (VARCHAR) | Raw materials catalog (e.g., "Onion", "Cheddar", "Bacon"). |
| `item_ingredient` | `item_id` (PK, FK -> `menu_item.id`)<br>`ingredient_id` (PK, FK -> `Ingredient_item.id`)<br>`proportion` (VARCHAR)<br>`is_removable` (BOOLEAN) | Technical spec. Defines whether customers can customize (remove) an ingredient. |
| `item_side` | `id` (PK)<br>`item_id` (FK -> `menu_item.id`)<br>`title` (VARCHAR)<br>`min_options` (INT)<br>`max_options` (INT) | Group of options/addons (e.g. "Meat Doneness" min=1 max=1, "Extra Addons" min=0 max=3). |
| `side_option` | `id` (PK)<br>`side_id` (FK -> `item_side.id`)<br>`name` (VARCHAR)<br>`additional_price` (DECIMAL(10,2)) | Concrete choices inside a side group (e.g. "+R$ 4.00 Bacon"). `additional_price >= 0`. |
| `combo_item` | `combo_id` (PK, FK -> `menu_item.id`)<br>`sub_item_id` (PK, FK -> `menu_item.id`)<br>`quantity` (INT) | Blueprint bundle: defines which products and quantities compose a combo item. |
| `item_sale` | `id` (PK)<br>`item_id` (FK -> `menu_item.id`)<br>`discount_type` ('percent'\|'fixed')<br>`discount_value` (DECIMAL(10,2))<br>`starts_at` (TIMESTAMP)<br>`ends_at` (TIMESTAMP) | Temporary price discount rules active within the `starts_at` -> `ends_at` window. |

### Group D: Loyalty & Promotions

| Table | Columns (PK, FK, Constraints) | Invariants & Business Rules |
| :--- | :--- | :--- |
| `user_restaurant` | `user_cpf` (PK, FK -> `users.cpf`)<br>`restaurant_slug` (PK, FK -> `restaurant.slug`)<br>`points_balance` (INT) | Tracks customer coins/cashback balance per tenant (`points_balance >= 0`). |
| `coupon` | `id` (PK)<br>`restaurant_slug` (FK -> `restaurant.slug`)<br>`code` (VARCHAR)<br>`discount_type` ('percent'\|'fixed')<br>`discount_value` (DECIMAL(10,2))<br>`min_order_value` (DECIMAL(10,2))<br>`expires_at` (TIMESTAMP)<br>`max_uses` (INT) | Tenant coupon validation. `code` must be unique per restaurant tenant. |

### Group E: Orders & Sales (Dine-in vs. Delivery)

| Table | Columns (PK, FK, Constraints) | Invariants & Business Rules |
| :--- | :--- | :--- |
| `order` | `id` (PK)<br>`branch_id` (FK -> `physical_restaurant.id`)<br>`user_cpf` (FK -> `users.cpf`, NULLABLE)<br>`order_type` ('local'\|'remote')<br>`status` ('pending'\|'preparing'\|'ready'\|'delivered'\|'cancelled')<br>`subtotal` (DECIMAL(10,2))<br>`discount` (DECIMAL(10,2))<br>`total` (DECIMAL(10,2))<br>`created_at` (TIMESTAMP) | Root order. `total = (subtotal - discount + freight_fee)`. Subtotal is computed from `order_item`. |
| `local_order` | `order_id` (PK, FK -> `order.id`)<br>`table_number` (INT)<br>`display_order` (INT) | 1-to-1 specialization for dine-in consumption at tables/tabs. |
| `remote_order` | `order_id` (PK, FK -> `order.id`)<br>`address_id` (FK -> `Address.id`)<br>`contact_phone` (VARCHAR)<br>`freight_fee` (DECIMAL(10,2)) | 1-to-1 specialization for delivery orders. Freight fee added to total. |
| `order_item` | `id` (PK)<br>`order_id` (FK -> `order.id`)<br>`item_id` (FK -> `menu_item.id`)<br>`unit_price_frozen` (DECIMAL(10,2))<br>`quantity` (INT)<br>`notes` (TEXT)<br>`parent_order_item_id` (FK -> `order_item.id`, NULLABLE) | Line items. **`unit_price_frozen` locks the price** at purchase time. Supports hierarchy for combos. |

### Group F: Payments & Digital Receipts

| Table | Columns (PK, FK, Constraints) | Invariants & Business Rules |
| :--- | :--- | :--- |
| `transaction` | `id` (PK)<br>`order_id` (FK -> `order.id`)<br>`amount` (DECIMAL(10,2))<br>`payment_method` ('credit_card'\|'debit_card'\|'pix'\|'cash')<br>`status` ('pending'\|'paid'\|'failed'\|'refunded')<br>`created_at` (TIMESTAMP) | Financial record of payment attempts. An order may have split or retried transactions. |
| `transaction_pix` | `transaction_id` (PK, FK -> `transaction.id`)<br>`copy_paste_code` (TEXT)<br>`qr_code_base64` (TEXT)<br>`expires_at` (TIMESTAMP) | Specialized PIX payload for real-time Brazilian instant payments. |
| `receipt` | `receipt_code` (PK, VARCHAR)<br>`order_id` (FK -> `order.id`, UNIQUE)<br>`receipt_json` (JSONB/TEXT)<br>`issued_at` (TIMESTAMP) | Immutable serialized snapshot of completed order for legal/accounting records. |

### Group G: Shifts & Staff Operations

| Table | Columns (PK, FK, Constraints) | Invariants & Business Rules |
| :--- | :--- | :--- |
| `restaurant_staff_session` | `id` (PK)<br>`branch_id` (FK -> `physical_restaurant.id`)<br>`shift_code` (VARCHAR)<br>`qr_token` (VARCHAR)<br>`opened_at` (TIMESTAMP)<br>`closed_at` (TIMESTAMP, NULLABLE) | Active operational shift in a branch (e.g. Lunch or Dinner). Provides QR token for clock-in. |
| `staff_session` | `id` (PK)<br>`restaurant_staff_session_id` (FK -> `restaurant_staff_session.id`)<br>`account_id` (FK -> `account.id`)<br>`checked_in_at` (TIMESTAMP)<br>`checked_out_at` (TIMESTAMP, NULLABLE) | Individual employee attendance punch tied to the active branch shift. |

---

## 3. Core Business Invariants for Implementation

1. **Multi-Tenancy Isolation:**
   - Every catalog resource (`menu`, `category`, `menu_item`, `Ingredient_item`, `coupon`) is bound directly or transitively to `restaurant.slug`. Queries MUST filter by tenant.
2. **Order Inheritance:**
   - An `order` MUST have exactly one matching subtype record: either `local_order` (if `order_type = 'local'`) or `remote_order` (if `order_type = 'remote'`).
3. **Price Immutability:**
   - Never query `menu_item.base_price` to compute historic order totals. Always use `order_item.unit_price_frozen`.
4. **Combo Composition:**
   - If `menu_item.is_combo == true`, child units are defined in `combo_item`. In `order_item`, children point to parent via `parent_order_item_id`.
5. **Business Model (Take-Rate):**
   - Platform monetizes via a % fee on `order.total`. Platform earnings = `order.total * commission_rate`.
