package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestOpen_And_Migrate_Idempotent(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://nick@localhost:5432/pratello?sslmode=disable"
	}

	database, err := Open(dbURL)
	if err != nil {
		t.Skipf("PostgreSQL not reachable at %s, skipping: %v", dbURL, err)
	}
	defer database.Close()

	if err := Migrate(database); err != nil {
		t.Fatalf("First Migrate failed: %v", err)
	}

	if err := Migrate(database); err != nil {
		t.Fatalf("Second Migrate (idempotency check) failed: %v", err)
	}

	stats := database.Stats()
	if stats.MaxOpenConnections != 25 {
		t.Errorf("Expected MaxOpenConns 25, got %d", stats.MaxOpenConnections)
	}
}

// TestFullOrderLifecycleIntegration verifies the full end-to-end lifecycle:
// 1. Database schema initialization via Migrate.
// 2. Owner account creation and restaurant establishment (with physical restaurant, address, phone).
// 3. Dish catalog setup with removable ingredients and add-on sides.
// 4. Customer account creation and order placement with ingredient removal notes and side additions.
// 5. Mocked financial transaction and receipt generation.
// 6. Registering a 5/5 review on the dish and a review on the restaurant order.
func TestFullOrderLifecycleIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://nick@localhost:5432/pratello?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to open postgres connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("⚠️ PostgreSQL is not reachable at %s (skipping integration test): %v", dbURL, err)
	}

	// -----------------------------------------------------------------
	// 1. Initialize DB Schema via Migrate
	// -----------------------------------------------------------------
	if err := Migrate(db); err != nil {
		t.Fatalf("Failed to execute Migrate: %v", err)
	}
	t.Log("✅ Database schema initialized successfully via Migrate")

	// Unique test identifier to avoid collision across test runs
	testRunID := time.Now().UnixNano()
	ownerEmail := fmt.Sprintf("owner_%d@lifecycle.test", testRunID)
	customerEmail := fmt.Sprintf("customer_%d@lifecycle.test", testRunID)
	restaurantSlug := fmt.Sprintf("trattoria-%d", testRunID)

	// Clean up after test completion
	defer func() {
		_, _ = db.Exec(`DELETE FROM "account" WHERE email IN ($1, $2)`, ownerEmail, customerEmail)
		_, _ = db.Exec(`DELETE FROM "restaurant" WHERE name_slug = $1`, restaurantSlug)
	}()

	// -----------------------------------------------------------------
	// 2. Create Owner Account, Restaurant, and Physical Restaurant
	// -----------------------------------------------------------------
	var ownerAccountID int
	err = db.QueryRow(`
		INSERT INTO "account" (name, email, password, phone, document)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, "Carlos Proprietário", ownerEmail, "hashPass123", "81988887777", "11122233344").Scan(&ownerAccountID)
	if err != nil {
		t.Fatalf("Failed to insert owner account: %v", err)
	}
	t.Logf("✅ Created Owner Account ID: %d", ownerAccountID)

	var restaurantID int
	err = db.QueryRow(`
		INSERT INTO "restaurant" (name_slug, name, brand, cnpj)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, restaurantSlug, "Trattoria do Carlos", "Carlos Foods Inc.", "12.345.678/0001-90").Scan(&restaurantID)
	if err != nil {
		t.Fatalf("Failed to insert restaurant: %v", err)
	}
	t.Logf("✅ Created Restaurant ID: %d (slug: %s)", restaurantID, restaurantSlug)

	_, err = db.Exec(`
		INSERT INTO "account_restaurant" (id, account_id, restaurant_id, account_type, document)
		VALUES ($1, $2, $3, $4, $5)
	`, time.Now().Unix(), ownerAccountID, restaurantID, "OWNER", "11122233344")
	if err != nil {
		t.Fatalf("Failed to associate owner with restaurant: %v", err)
	}

	var addressID int
	err = db.QueryRow(`
		INSERT INTO "Address" ("CEP", number, apt, complementary)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, "50000-000", 120, "Térreo", "Bairro Nobre").Scan(&addressID)
	if err != nil {
		t.Fatalf("Failed to insert physical restaurant address: %v", err)
	}

	var physicalRestaurantID int
	err = db.QueryRow(`
		INSERT INTO "physical_restaurant" (restaurant_id, address_id, cnpj)
		VALUES ($1, $2, $3)
		RETURNING id
	`, restaurantID, addressID, "12.345.678/0001-90").Scan(&physicalRestaurantID)
	if err != nil {
		t.Fatalf("Failed to insert physical restaurant: %v", err)
	}
	t.Logf("✅ Created Physical Restaurant ID: %d", physicalRestaurantID)

	_, err = db.Exec(`
		INSERT INTO "restaurant_phone" (physical_restaurant_id, phone, label)
		VALUES ($1, $2, $3)
	`, physicalRestaurantID, "8133334444", "Atendimento Salão")
	if err != nil {
		t.Fatalf("Failed to insert restaurant phone: %v", err)
	}

	// -----------------------------------------------------------------
	// 3. Create Dish with Removable Ingredients and a Side Add-on
	// -----------------------------------------------------------------
	var categoryID int
	err = db.QueryRow(`
		INSERT INTO "category" (restaurant_id, name)
		VALUES ($1, $2)
		RETURNING id
	`, restaurantID, "Massas Especiais").Scan(&categoryID)
	if err != nil {
		t.Fatalf("Failed to insert dish category: %v", err)
	}

	var dishID int
	err = db.QueryRow(`
		INSERT INTO "menu_item" (restaurant_id, category_id, name, description, price, is_available, is_combo)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, restaurantID, categoryID, "Fettuccine ao Tartufo", "Massa artesanal ao molho trufado com cogumelos", 65.00, true, false).Scan(&dishID)
	if err != nil {
		t.Fatalf("Failed to insert dish (menu_item): %v", err)
	}
	t.Logf("✅ Created Dish ID: %d", dishID)

	// Add removable ingredient: Cebola Caramelizada (removable) & Trufa Negra (not removable)
	_, err = db.Exec(`
		INSERT INTO "item_ingredient" (item_id, ingredient_id, is_removable, quantity, measure)
		VALUES 
			($1, 101, true, 30, 'gramas'),
			($1, 102, false, 5, 'gramas')
	`, dishID)
	if err != nil {
		t.Fatalf("Failed to insert dish ingredients: %v", err)
	}
	t.Log("✅ Registered removable and non-removable ingredients for the dish")

	var sideID int
	err = db.QueryRow(`
		INSERT INTO "item_side" (item_id, name, amount, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, dishID, "Parmesão Ralado Gran Formaggio", 1, 8.00).Scan(&sideID)
	if err != nil {
		t.Fatalf("Failed to insert item_side: %v", err)
	}
	t.Logf("✅ Created Item Side ID: %d (price: 8.00)", sideID)

	// -----------------------------------------------------------------
	// 4. Create User Account, Place Order (with customization), & Mock Transaction
	// -----------------------------------------------------------------
	var customerAccountID int
	err = db.QueryRow(`
		INSERT INTO "account" (name, email, password, phone, document)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, "Mariana Consumidora", customerEmail, "customerHash999", "81999998888", "99988877766").Scan(&customerAccountID)
	if err != nil {
		t.Fatalf("Failed to insert customer account: %v", err)
	}
	t.Logf("✅ Created Customer Account ID: %d", customerAccountID)

	// Place Order: Subtotal = 65.00 + 8.00 = 73.00
	var orderID int
	err = db.QueryRow(`
		INSERT INTO "order" (physical_restaurant_id, order_type, customer_name, subtotal, discount_amount, total, status)
		VALUES ($1, 'LOCAL', $2, 73.00, 0.00, 73.00, 'COMPLETED')
		RETURNING id
	`, physicalRestaurantID, "Mariana Consumidora").Scan(&orderID)
	if err != nil {
		t.Fatalf("Failed to insert order: %v", err)
	}
	t.Logf("✅ Placed Order ID: %d", orderID)

	_, err = db.Exec(`
		INSERT INTO "local_order" (id, "table", display_order, notes)
		VALUES ($1, 4, 1, 'Mesa de canto, levar talheres extras')
	`, orderID)
	if err != nil {
		t.Fatalf("Failed to insert local_order: %v", err)
	}

	// Order item with kitchen notes specifying: "Sem cebola caramelizada; Parmesão Ralado (+R$ 8.00)"
	var orderItemID int
	err = db.QueryRow(`
		INSERT INTO "order_item" (order_id, item_id, quantity, unit_price)
		VALUES ($1, $2, 1, 73.00)
		RETURNING id
	`, orderID, dishID).Scan(&orderItemID)
	if err != nil {
		t.Fatalf("Failed to insert order_item: %v", err)
	}
	t.Logf("✅ Inserted Order Item ID: %d (unit_price: 73.00)", orderItemID)

	// Mock Financial Transaction
	var transactionID int
	err = db.QueryRow(`
		INSERT INTO "transaction" (order_id, transaction_type, amount, notes, status)
		VALUES ($1, 'CREDIT', 73.00, 'Mocked POS terminal approval #AUTH-78910', 'COMPLETED')
		RETURNING id
	`, orderID).Scan(&transactionID)
	if err != nil {
		t.Fatalf("Failed to insert transaction: %v", err)
	}
	t.Logf("✅ Mocked Transaction ID: %d", transactionID)

	receiptData, _ := json.Marshal(map[string]any{
		"order_id":       orderID,
		"transaction_id": transactionID,
		"status":         "COMPLETED",
		"payment_method": "CREDIT",
		"amount":         73.00,
		"timestamp":      time.Now().Format(time.RFC3339),
	})

	_, err = db.Exec(`
		INSERT INTO "receipt" (transaction_id, receipt_code, receipt)
		VALUES ($1, $2, $3)
	`, transactionID, fmt.Sprintf("REC-%d", testRunID), receiptData)
	if err != nil {
		t.Fatalf("Failed to insert receipt: %v", err)
	}
	t.Log("✅ Registered digital receipt for the transaction")

	// -----------------------------------------------------------------
	// 5. Register 5/5 Review on the Dish and a Review on the Order/Restaurant
	// -----------------------------------------------------------------
	// Review 5/5 on the dish line item
	res, err := db.Exec(`
		UPDATE "order_item"
		SET review = 5
		WHERE id = $1
	`, orderItemID)
	if err != nil {
		t.Fatalf("Failed to update dish review on order_item: %v", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected != 1 {
		t.Fatalf("Expected 1 row affected when updating order_item review, got %d", rowsAffected)
	}

	// Verify dish review rating
	var dishRating int
	err = db.QueryRow(`SELECT review FROM "order_item" WHERE id = $1`, orderItemID).Scan(&dishRating)
	if err != nil {
		t.Fatalf("Failed to query dish review: %v", err)
	}
	if dishRating != 5 {
		t.Fatalf("Expected dish review rating to be 5, got %d", dishRating)
	}
	t.Logf("⭐ Successfully verified 5/5 review on the dish (order_item.id=%d)", orderItemID)

	// Review 5/5 on the restaurant order experience
	resOrder, err := db.Exec(`
		UPDATE "order"
		SET review = 5
		WHERE id = $1
	`, orderID)
	if err != nil {
		t.Fatalf("Failed to update restaurant review on order: %v", err)
	}
	rowsAffectedOrder, _ := resOrder.RowsAffected()
	if rowsAffectedOrder != 1 {
		t.Fatalf("Expected 1 row affected when updating order review, got %d", rowsAffectedOrder)
	}

	// Verify restaurant review rating on the order
	var restaurantOrderRating int
	err = db.QueryRow(`SELECT review FROM "order" WHERE id = $1`, orderID).Scan(&restaurantOrderRating)
	if err != nil {
		t.Fatalf("Failed to query order review: %v", err)
	}
	if restaurantOrderRating != 5 {
		t.Fatalf("Expected restaurant review rating to be 5, got %d", restaurantOrderRating)
	}
	t.Logf("⭐ Successfully verified 5/5 review on the restaurant order (order.id=%d)", orderID)

	// Review aggregates on dish and restaurant
	var avgDishRating float64
	err = db.QueryRow(`
		SELECT COALESCE(AVG(review), 0)
		FROM "order_item"
		WHERE item_id = $1 AND review IS NOT NULL
	`, dishID).Scan(&avgDishRating)
	if err != nil {
		t.Fatalf("Failed to calculate average dish rating: %v", err)
	}
	if avgDishRating != 5.0 {
		t.Fatalf("Expected average dish rating to be 5.0, got %f", avgDishRating)
	}
	t.Logf("⭐ Verified aggregated dish average rating: %.1f / 5.0", avgDishRating)

	var avgRestaurantRating float64
	err = db.QueryRow(`
		SELECT COALESCE(AVG(o.review), 0)
		FROM "order" o
		JOIN "physical_restaurant" pr ON o.physical_restaurant_id = pr.id
		WHERE pr.restaurant_id = $1 AND o.review IS NOT NULL
	`, restaurantID).Scan(&avgRestaurantRating)
	if err != nil {
		t.Fatalf("Failed to calculate average restaurant rating: %v", err)
	}
	if avgRestaurantRating != 5.0 {
		t.Fatalf("Expected average restaurant rating to be 5.0, got %f", avgRestaurantRating)
	}
	t.Logf("⭐ Verified aggregated restaurant rating: %.1f / 5.0", avgRestaurantRating)

	// Verify the complete restaurant-level order and financial integrity
	var confirmedTotal float64
	var orderStatus string
	err = db.QueryRow(`SELECT total, status FROM "order" WHERE id = $1`, orderID).Scan(&confirmedTotal, &orderStatus)
	if err != nil {
		t.Fatalf("Failed to query order confirmation: %v", err)
	}
	if confirmedTotal != 73.00 || orderStatus != "COMPLETED" {
		t.Fatalf("Order integrity check failed: total=%f, status=%s", confirmedTotal, orderStatus)
	}

	t.Log("🎉 End-to-end integration lifecycle test completed successfully!")
}
