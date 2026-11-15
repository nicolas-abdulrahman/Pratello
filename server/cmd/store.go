package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound       = errors.New("resource not found")
	ErrConflict       = errors.New("resource already exists")
	ErrTerminalLocked = errors.New("terminal is locked due to too many failed attempts")
	ErrInvalidPIN     = errors.New("invalid terminal PIN")
	ErrOrphanEntity   = errors.New("cannot remove the last owner: entity cannot be orphaned")
)

// Domain inputs (free of HTTP transport concerns and JSON annotations):

// NewRestaurant holds the information needed to register a new restaurant.
type NewRestaurant struct {
	Name        string
	Slug        string
	Description string
}

// UpdateRestaurantInput holds the fields for modifying a restaurant.
type UpdateRestaurantInput struct {
	Name        string
	Description string
}

// NewDish holds all the required data to create a dish under a restaurant.
type NewDish struct {
	RestaurantSlug string
	Name           string
	Description    string
	Category       string
	Tags           []string
	Price          float64
	ImageURL       string
}

// UpdateDishInput holds the fields for modifying a dish.
type UpdateDishInput struct {
	Name        string
	Description string
	Category    string
	Tags        []string
	Price       float64
	ImageURL    string
}

// RestaurantStore defines the contract for restaurant data persistence.
type RestaurantStore interface {
	List(ctx context.Context) ([]Restaurant, error)
	GetBySlug(ctx context.Context, slug string) (*Restaurant, error)
	Create(ctx context.Context, r *Restaurant) error
	Update(ctx context.Context, slug string, input UpdateRestaurantInput) (*Restaurant, error)
	Delete(ctx context.Context, slug string) error
}

// DishStore defines the contract for dish data persistence.
type DishStore interface {
	ListByRestaurant(ctx context.Context, restaurantSlug string) ([]Dish, error)
	Create(ctx context.Context, input NewDish) (*Dish, error)
	Update(ctx context.Context, restaurantSlug string, dishID int, input UpdateDishInput) (*Dish, error)
	Delete(ctx context.Context, restaurantSlug string, dishID int) error
	UpdateImage(ctx context.Context, restaurantSlug string, dishID int, imageURL string) (*Dish, error)
}

// SQLRestaurantStore implements RestaurantStore backed by a SQL database.
type SQLRestaurantStore struct {
	DB *sql.DB
}

func NewSQLRestaurantStore(db *sql.DB) RestaurantStore {
	if db == nil {
		return NewMemoryRestaurantStore()
	}
	return &SQLRestaurantStore{DB: db}
}

func (s *SQLRestaurantStore) List(ctx context.Context) ([]Restaurant, error) {
	query := `
		SELECT slug, name, COALESCE(description, ''), COALESCE(cuisine, ''),
		       COALESCE(is_open, true), COALESCE(open_time, '07:00'), COALESCE(closing_time, '22:00'),
		       COALESCE(rating, 5.0), COALESCE(reviews_count, 0), COALESCE(delivery_fee, 0.0),
		       COALESCE(cep, ''), COALESCE(delivery_time, '20–30 min'), COALESCE(neighborhood, ''),
		       COALESCE(phone, ''), COALESCE(banner_url, ''), COALESCE(logo_url, '')
		FROM restaurants
	`
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Restaurant
	for rows.Next() {
		var r Restaurant
		if err := rows.Scan(
			&r.Slug, &r.Name, &r.Description, &r.Cuisine,
			&r.IsOpen, &r.OpenTime, &r.ClosingTime,
			&r.Rating, &r.ReviewsCount, &r.DeliveryFee,
			&r.CEP, &r.DeliveryTime, &r.Neighborhood,
			&r.Phone, &r.BannerURL, &r.LogoURL,
		); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s *SQLRestaurantStore) GetBySlug(ctx context.Context, slug string) (*Restaurant, error) {
	var r Restaurant
	query := `
		SELECT slug, name, COALESCE(description, ''), COALESCE(cuisine, ''),
		       COALESCE(is_open, true), COALESCE(open_time, '07:00'), COALESCE(closing_time, '22:00'),
		       COALESCE(rating, 5.0), COALESCE(reviews_count, 0), COALESCE(delivery_fee, 0.0),
		       COALESCE(cep, ''), COALESCE(delivery_time, '20–30 min'), COALESCE(neighborhood, ''),
		       COALESCE(phone, ''), COALESCE(banner_url, ''), COALESCE(logo_url, '')
		FROM restaurants WHERE slug = $1
	`
	err := s.DB.QueryRowContext(ctx, query, slug).Scan(
		&r.Slug, &r.Name, &r.Description, &r.Cuisine,
		&r.IsOpen, &r.OpenTime, &r.ClosingTime,
		&r.Rating, &r.ReviewsCount, &r.DeliveryFee,
		&r.CEP, &r.DeliveryTime, &r.Neighborhood,
		&r.Phone, &r.BannerURL, &r.LogoURL,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *SQLRestaurantStore) Create(ctx context.Context, r *Restaurant) error {
	var existing string
	err := s.DB.QueryRowContext(ctx, "SELECT slug FROM restaurants WHERE slug = $1", r.Slug).Scan(&existing)
	if err == nil {
		return ErrConflict
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	query := `
		INSERT INTO restaurants (
			slug, name, description, cuisine, is_open, open_time, closing_time,
			rating, reviews_count, delivery_fee, cep, delivery_time, neighborhood,
			phone, banner_url, logo_url
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16
		)
	`
	_, err = s.DB.ExecContext(ctx, query,
		r.Slug, r.Name, r.Description, r.Cuisine, r.IsOpen, r.OpenTime, r.ClosingTime,
		r.Rating, r.ReviewsCount, r.DeliveryFee, r.CEP, r.DeliveryTime, r.Neighborhood,
		r.Phone, r.BannerURL, r.LogoURL,
	)
	return err
}

func (s *SQLRestaurantStore) Update(ctx context.Context, slug string, input UpdateRestaurantInput) (*Restaurant, error) {
	res, err := s.DB.ExecContext(ctx,
		"UPDATE restaurants SET name = $1, description = $2 WHERE slug = $3",
		input.Name, input.Description, slug,
	)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	return &Restaurant{
		Slug:        slug,
		Name:        input.Name,
		Description: input.Description,
	}, nil
}

func (s *SQLRestaurantStore) Delete(ctx context.Context, slug string) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM restaurants WHERE slug = $1", slug)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// SQLDishStore implements DishStore backed by a SQL database.
type SQLDishStore struct {
	DB *sql.DB
}

func NewSQLDishStore(db *sql.DB) DishStore {
	if db == nil {
		return NewMemoryDishStore()
	}
	return &SQLDishStore{DB: db}
}

func (s *SQLDishStore) ListByRestaurant(ctx context.Context, restaurantSlug string) ([]Dish, error) {
	var dummy string
	if err := s.DB.QueryRowContext(ctx, "SELECT slug FROM restaurants WHERE slug = $1", restaurantSlug).Scan(&dummy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	query := `
		SELECT id, name, COALESCE(description, ''), COALESCE(category, 'Pizzas'),
		       COALESCE(tags, ''), price, COALESCE(image_url, '')
		FROM dishes WHERE restaurant_slug = $1 ORDER BY id
	`
	rows, err := s.DB.QueryContext(ctx, query, restaurantSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Dish
	for rows.Next() {
		var d Dish
		var tagsStr string
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.Category, &tagsStr, &d.Price, &d.ImageURL); err != nil {
			return nil, err
		}
		d.RestaurantSlug = restaurantSlug
		if tagsStr != "" {
			d.Tags = strings.Split(tagsStr, ",")
		} else {
			d.Tags = []string{}
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

func (s *SQLDishStore) Create(ctx context.Context, input NewDish) (*Dish, error) {
	var dummy string
	if err := s.DB.QueryRowContext(ctx, "SELECT slug FROM restaurants WHERE slug = $1", input.RestaurantSlug).Scan(&dummy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	tagsStr := strings.Join(input.Tags, ",")
	if input.Category == "" {
		input.Category = "Pizzas"
	}

	query := `
		INSERT INTO dishes (restaurant_slug, name, description, category, tags, price, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var newID int
	err := s.DB.QueryRowContext(ctx, query,
		input.RestaurantSlug, input.Name, input.Description, input.Category, tagsStr, input.Price, input.ImageURL,
	).Scan(&newID)
	if err != nil {
		return nil, err
	}

	return &Dish{
		ID:             newID,
		RestaurantSlug: input.RestaurantSlug,
		Name:           input.Name,
		Description:    input.Description,
		Category:       input.Category,
		Tags:           input.Tags,
		Price:          input.Price,
		ImageURL:       input.ImageURL,
	}, nil
}

func (s *SQLDishStore) Update(ctx context.Context, restaurantSlug string, dishID int, input UpdateDishInput) (*Dish, error) {
	res, err := s.DB.ExecContext(ctx,
		"UPDATE dishes SET name = $1, price = $2, image_url = CASE WHEN $3 = '' THEN image_url ELSE $3 END WHERE id = $4 AND restaurant_slug = $5",
		input.Name, input.Price, input.ImageURL, dishID, restaurantSlug,
	)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	var d Dish
	err = s.DB.QueryRowContext(ctx,
		"SELECT id, name, price, COALESCE(image_url, '') FROM dishes WHERE id = $1 AND restaurant_slug = $2",
		dishID, restaurantSlug,
	).Scan(&d.ID, &d.Name, &d.Price, &d.ImageURL)
	if err != nil {
		return nil, err
	}
	d.RestaurantSlug = restaurantSlug
	return &d, nil
}

func (s *SQLDishStore) UpdateImage(ctx context.Context, restaurantSlug string, dishID int, imageURL string) (*Dish, error) {
	res, err := s.DB.ExecContext(ctx,
		"UPDATE dishes SET image_url = $1 WHERE id = $2 AND restaurant_slug = $3",
		imageURL, dishID, restaurantSlug,
	)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	var d Dish
	err = s.DB.QueryRowContext(ctx,
		"SELECT id, name, price, COALESCE(image_url, '') FROM dishes WHERE id = $1 AND restaurant_slug = $2",
		dishID, restaurantSlug,
	).Scan(&d.ID, &d.Name, &d.Price, &d.ImageURL)
	if err != nil {
		return nil, err
	}
	d.RestaurantSlug = restaurantSlug
	return &d, nil
}

func (s *SQLDishStore) Delete(ctx context.Context, restaurantSlug string, dishID int) error {
	res, err := s.DB.ExecContext(ctx,
		"DELETE FROM dishes WHERE id = $1 AND restaurant_slug = $2",
		dishID, restaurantSlug,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------------- ACCOUNTS, MANAGERS & SYSADMINS PERSISTENCE ----------------

type Account struct {
	ID        int       `json:"id"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Sysadmin struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Manager struct {
	ID             int    `json:"id"`
	RestaurantSlug string `json:"restaurant_slug"`
	AccountID      int    `json:"account_id"`
}

type Owner struct {
	ID             int    `json:"id"`
	RestaurantSlug string `json:"restaurant_slug"`
	AccountID      int    `json:"account_id"`
}

type PhysicalRestaurant struct {
	ID                int        `json:"id"`
	RestaurantID      int        `json:"restaurant_id"`
	RestaurantSlug    string     `json:"restaurant_slug"`
	AddressID         int        `json:"address_id"`
	Name              string     `json:"name"`
	CNPJ              string     `json:"cnpj"`
	OwnerCNPJ         string     `json:"owner_cnpj"`
	TerminalPIN       string     `json:"-"`
	FailedPINAttempts int        `json:"-"`
	LockedUntil       *time.Time `json:"locked_until,omitempty"`
}

type Employee struct {
	ID                   int       `json:"id"`
	PhysicalRestaurantID int       `json:"physical_restaurant_id"`
	Name                 string    `json:"name"`
	CPF                  string    `json:"cpf"`
	Position             string    `json:"position"`
	AccountID            *int      `json:"account_id,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

type TerminalSession struct {
	ID                   int       `json:"id"`
	PhysicalRestaurantID int       `json:"physical_restaurant_id"`
	Token                string    `json:"token"`
	DeviceID             string    `json:"device_id"`
	ExpiresAt            time.Time `json:"expires_at"`
	CreatedAt            time.Time `json:"created_at"`
}

type AccountStore interface {
	CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountByID(ctx context.Context, id int) (*Account, error)
	ListSysadmins(ctx context.Context) ([]Sysadmin, error)
	AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error)
	ListManagers(ctx context.Context, restaurantSlug string) ([]Manager, error)
	AddManager(ctx context.Context, restaurantSlug string, accountID int) (*Manager, error)
	RemoveManager(ctx context.Context, restaurantSlug string, accountID int) error

	// Owner Entity-level authority
	ListOwners(ctx context.Context, restaurantSlug string) ([]Owner, error)
	AddOwner(ctx context.Context, restaurantSlug string, accountID int) (*Owner, error)
	RemoveOwner(ctx context.Context, restaurantSlug string, accountID int) error
	GetAccountRoles(ctx context.Context, restaurantSlug string, accountID int) ([]string, error)

	// Physical Restaurants (Branches)
	CreatePhysicalRestaurant(ctx context.Context, restaurantSlug, name, cnpj, ownerCnpj string) (*PhysicalRestaurant, error)
	ListPhysicalRestaurants(ctx context.Context, restaurantSlug string) ([]PhysicalRestaurant, error)
	GetPhysicalRestaurant(ctx context.Context, branchID int) (*PhysicalRestaurant, error)
	GetRestaurantCNPJ(ctx context.Context, restaurantSlug string) (string, error)

	// Employees (HR records without auto-login)
	ListEmployees(ctx context.Context, branchID int) ([]Employee, error)
	CreateEmployee(ctx context.Context, branchID int, name, cpf, position string) (*Employee, error)
	LinkEmployeeAccount(ctx context.Context, employeeID int, accountID int) (*Employee, error)

	// Shift Terminal Sessions (Shared 6-digit PIN, device-bound, 8h expiry)
	VerifyAndOpenTerminalSession(ctx context.Context, branchID int, pin string, deviceID string) (*TerminalSession, error)
	GetTerminalSession(ctx context.Context, token string) (*TerminalSession, error)
	RotateTerminalPIN(ctx context.Context, branchID int, newPIN string) error
}

func NewSQLAccountStore(db *sql.DB) AccountStore {
	if db == nil {
		return NewMemoryAccountStore()
	}
	return &SQLAccountStore{DB: db}
}

type SQLAccountStore struct {
	DB *sql.DB
}

func (s *SQLAccountStore) CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error) {
	var a Account
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO "account" (name, email, password, phone, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 RETURNING id, COALESCE(name, ''), email, password, COALESCE(phone, ''), COALESCE(created_at, NOW())`,
		name, strings.ToLower(strings.TrimSpace(email)), passwordHash, phone,
	).Scan(&a.ID, &a.Name, &a.Email, &a.Password, &a.Phone, &a.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, ErrConflict
		}
		return nil, err
	}
	return &a, nil
}

func (s *SQLAccountStore) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	var a Account
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, COALESCE(name, ''), email, password, COALESCE(phone, ''), COALESCE(created_at, NOW())
		 FROM "account"
		 WHERE LOWER(email) = $1`,
		strings.ToLower(strings.TrimSpace(email)),
	).Scan(&a.ID, &a.Name, &a.Email, &a.Password, &a.Phone, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (s *SQLAccountStore) GetAccountByID(ctx context.Context, id int) (*Account, error) {
	var a Account
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, COALESCE(name, ''), email, password, COALESCE(phone, ''), COALESCE(created_at, NOW())
		 FROM "account"
		 WHERE id = $1`,
		id,
	).Scan(&a.ID, &a.Name, &a.Email, &a.Password, &a.Phone, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (s *SQLAccountStore) ListSysadmins(ctx context.Context) ([]Sysadmin, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, account_id, created_at FROM "sysadmin" ORDER BY id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Sysadmin
	for rows.Next() {
		var sa Sysadmin
		if err := rows.Scan(&sa.ID, &sa.AccountID, &sa.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, sa)
	}
	if list == nil {
		list = make([]Sysadmin, 0)
	}
	return list, rows.Err()
}

func (s *SQLAccountStore) AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error) {
	if _, err := s.GetAccountByID(ctx, accountID); err != nil {
		return nil, err
	}

	var sa Sysadmin
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO "sysadmin" (account_id, created_at)
		 VALUES ($1, NOW())
		 ON CONFLICT (account_id) DO UPDATE SET account_id = EXCLUDED.account_id
		 RETURNING id, account_id, created_at`,
		accountID,
	).Scan(&sa.ID, &sa.AccountID, &sa.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *SQLAccountStore) getRestaurantID(ctx context.Context, slug string) (int, error) {
	var id int
	err := s.DB.QueryRowContext(ctx, `SELECT id FROM "restaurant" WHERE name_slug = $1`, slug).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	// Legacy or fallback check
	var legacyName string
	err = s.DB.QueryRowContext(ctx, `SELECT name FROM restaurants WHERE slug = $1`, slug).Scan(&legacyName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO "restaurant" (name_slug, name) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING
		 RETURNING id`,
		slug, legacyName,
	).Scan(&id)
	if err != nil {
		_ = s.DB.QueryRowContext(ctx, `SELECT id FROM "restaurant" WHERE name_slug = $1`, slug).Scan(&id)
	}
	if id == 0 {
		return 0, ErrNotFound
	}
	return id, nil
}

func (s *SQLAccountStore) ListManagers(ctx context.Context, restaurantSlug string) ([]Manager, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.QueryContext(ctx,
		`SELECT ar.id, ar.account_id
		 FROM "account_restaurant" ar
		 WHERE ar.restaurant_id = $1 AND ar.account_type = 'manager'
		 ORDER BY ar.id ASC`,
		restID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var managers []Manager
	for rows.Next() {
		var m Manager
		if err := rows.Scan(&m.ID, &m.AccountID); err != nil {
			return nil, err
		}
		m.RestaurantSlug = restaurantSlug
		managers = append(managers, m)
	}
	if managers == nil {
		managers = make([]Manager, 0)
	}
	return managers, rows.Err()
}

func (s *SQLAccountStore) AddManager(ctx context.Context, restaurantSlug string, accountID int) (*Manager, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	if _, err := s.GetAccountByID(ctx, accountID); err != nil {
		return nil, err
	}

	var id int
	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO "account_restaurant" (account_id, restaurant_id, account_type)
		 VALUES ($1, $2, 'manager')
		 RETURNING id`,
		accountID, restID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &Manager{
		ID:             id,
		RestaurantSlug: restaurantSlug,
		AccountID:      accountID,
	}, nil
}

func (s *SQLAccountStore) RemoveManager(ctx context.Context, restaurantSlug string, accountID int) error {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return err
	}

	res, err := s.DB.ExecContext(ctx,
		`DELETE FROM "account_restaurant"
		 WHERE restaurant_id = $1 AND account_id = $2 AND account_type = 'manager'`,
		restID, accountID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLAccountStore) ListOwners(ctx context.Context, restaurantSlug string) ([]Owner, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.QueryContext(ctx,
		`SELECT ar.id, ar.account_id
		 FROM "account_restaurant" ar
		 WHERE ar.restaurant_id = $1 AND ar.account_type = 'owner'
		 ORDER BY ar.id ASC`,
		restID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var owners []Owner
	for rows.Next() {
		var o Owner
		if err := rows.Scan(&o.ID, &o.AccountID); err != nil {
			return nil, err
		}
		o.RestaurantSlug = restaurantSlug
		owners = append(owners, o)
	}
	if owners == nil {
		owners = make([]Owner, 0)
	}
	return owners, rows.Err()
}

func (s *SQLAccountStore) AddOwner(ctx context.Context, restaurantSlug string, accountID int) (*Owner, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	if _, err := s.GetAccountByID(ctx, accountID); err != nil {
		return nil, err
	}

	var id int
	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO "account_restaurant" (account_id, restaurant_id, account_type)
		 VALUES ($1, $2, 'owner')
		 RETURNING id`,
		accountID, restID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &Owner{
		ID:             id,
		RestaurantSlug: restaurantSlug,
		AccountID:      accountID,
	}, nil
}

func (s *SQLAccountStore) RemoveOwner(ctx context.Context, restaurantSlug string, accountID int) error {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return err
	}

	// Entity-level authority rule: entity cannot be orphaned. Multiple owners are allowed so entity can't be orphaned.
	var count int
	if err := s.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM "account_restaurant" WHERE restaurant_id = $1 AND account_type = 'owner'`,
		restID,
	).Scan(&count); err != nil {
		return err
	}
	if count <= 1 {
		return ErrOrphanEntity
	}

	res, err := s.DB.ExecContext(ctx,
		`DELETE FROM "account_restaurant"
		 WHERE restaurant_id = $1 AND account_id = $2 AND account_type = 'owner'`,
		restID, accountID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLAccountStore) GetAccountRoles(ctx context.Context, restaurantSlug string, accountID int) ([]string, error) {
	var roles []string

	// Check sysadmin
	var saCount int
	_ = s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM "sysadmin" WHERE account_id = $1`, accountID).Scan(&saCount)
	if saCount > 0 {
		roles = append(roles, "sysadmin")
	}

	if restaurantSlug != "" {
		restID, err := s.getRestaurantID(ctx, restaurantSlug)
		if err == nil {
			rows, err := s.DB.QueryContext(ctx,
				`SELECT DISTINCT account_type FROM "account_restaurant" WHERE restaurant_id = $1 AND account_id = $2`,
				restID, accountID,
			)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var rType string
					if err := rows.Scan(&rType); err == nil && rType != "" {
						roles = append(roles, rType)
					}
				}
			}
		}
	}

	return roles, nil
}

func (s *SQLAccountStore) GetRestaurantCNPJ(ctx context.Context, restaurantSlug string) (string, error) {
	var cnpj sql.NullString
	err := s.DB.QueryRowContext(ctx,
		`SELECT cnpj FROM "restaurant" WHERE name_slug = $1`,
		restaurantSlug,
	).Scan(&cnpj)
	if err == nil && cnpj.Valid && cnpj.String != "" {
		return cnpj.String, nil
	}

	// Legacy table fallback
	err = s.DB.QueryRowContext(ctx,
		`SELECT COALESCE(cep, '') FROM restaurants WHERE slug = $1`,
		restaurantSlug,
	).Scan(&cnpj)
	if err == nil && cnpj.Valid {
		return cnpj.String, nil
	}
	return "", nil
}

func (s *SQLAccountStore) CreatePhysicalRestaurant(ctx context.Context, restaurantSlug, name, cnpj, ownerCnpj string) (*PhysicalRestaurant, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	var addressID int
	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO "Address" ("CEP", number, complementary) VALUES ('00000-000', 0, 'Sede') RETURNING id`,
	).Scan(&addressID)
	if err != nil {
		_ = s.DB.QueryRowContext(ctx, `SELECT id FROM "Address" LIMIT 1`).Scan(&addressID)
		if addressID == 0 {
			addressID = 1
		}
	}

	var pr PhysicalRestaurant
	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO "physical_restaurant" (restaurant_id, address_id, name, cnpj, owner_cnpj, terminal_pin, pin_updated_at, failed_pin_attempts)
		 VALUES ($1, $2, $3, $4, $5, '123456', NOW(), 0)
		 RETURNING id, restaurant_id, address_id, name, cnpj, owner_cnpj, terminal_pin, failed_pin_attempts, locked_until`,
		restID, addressID, name, cnpj, ownerCnpj,
	).Scan(&pr.ID, &pr.RestaurantID, &pr.AddressID, &pr.Name, &pr.CNPJ, &pr.OwnerCNPJ, &pr.TerminalPIN, &pr.FailedPINAttempts, &pr.LockedUntil)
	if err != nil {
		return nil, err
	}
	pr.RestaurantSlug = restaurantSlug
	return &pr, nil
}

func (s *SQLAccountStore) ListPhysicalRestaurants(ctx context.Context, restaurantSlug string) ([]PhysicalRestaurant, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, restaurant_id, address_id, COALESCE(name, ''), cnpj, COALESCE(owner_cnpj, ''), terminal_pin, failed_pin_attempts, locked_until
		 FROM "physical_restaurant"
		 WHERE restaurant_id = $1
		 ORDER BY id ASC`,
		restID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PhysicalRestaurant
	for rows.Next() {
		var pr PhysicalRestaurant
		if err := rows.Scan(&pr.ID, &pr.RestaurantID, &pr.AddressID, &pr.Name, &pr.CNPJ, &pr.OwnerCNPJ, &pr.TerminalPIN, &pr.FailedPINAttempts, &pr.LockedUntil); err != nil {
			return nil, err
		}
		pr.RestaurantSlug = restaurantSlug
		list = append(list, pr)
	}
	if list == nil {
		list = make([]PhysicalRestaurant, 0)
	}
	return list, rows.Err()
}

func (s *SQLAccountStore) GetPhysicalRestaurant(ctx context.Context, branchID int) (*PhysicalRestaurant, error) {
	var pr PhysicalRestaurant
	err := s.DB.QueryRowContext(ctx,
		`SELECT pr.id, pr.restaurant_id, COALESCE(r.name_slug, ''), pr.address_id, COALESCE(pr.name, ''), pr.cnpj, COALESCE(pr.owner_cnpj, ''), pr.terminal_pin, pr.failed_pin_attempts, pr.locked_until
		 FROM "physical_restaurant" pr
		 JOIN "restaurant" r ON r.id = pr.restaurant_id
		 WHERE pr.id = $1`,
		branchID,
	).Scan(&pr.ID, &pr.RestaurantID, &pr.RestaurantSlug, &pr.AddressID, &pr.Name, &pr.CNPJ, &pr.OwnerCNPJ, &pr.TerminalPIN, &pr.FailedPINAttempts, &pr.LockedUntil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &pr, nil
}

func (s *SQLAccountStore) ListEmployees(ctx context.Context, branchID int) ([]Employee, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, physical_restaurant_id, name, cpf, position, account_id, created_at
		 FROM "employee"
		 WHERE physical_restaurant_id = $1
		 ORDER BY id ASC`,
		branchID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Employee
	for rows.Next() {
		var emp Employee
		if err := rows.Scan(&emp.ID, &emp.PhysicalRestaurantID, &emp.Name, &emp.CPF, &emp.Position, &emp.AccountID, &emp.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, emp)
	}
	if list == nil {
		list = make([]Employee, 0)
	}
	return list, rows.Err()
}

func (s *SQLAccountStore) CreateEmployee(ctx context.Context, branchID int, name, cpf, position string) (*Employee, error) {
	var emp Employee
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO "employee" (physical_restaurant_id, name, cpf, position, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 RETURNING id, physical_restaurant_id, name, cpf, position, account_id, created_at`,
		branchID, name, cpf, position,
	).Scan(&emp.ID, &emp.PhysicalRestaurantID, &emp.Name, &emp.CPF, &emp.Position, &emp.AccountID, &emp.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (s *SQLAccountStore) LinkEmployeeAccount(ctx context.Context, employeeID int, accountID int) (*Employee, error) {
	var emp Employee
	err := s.DB.QueryRowContext(ctx,
		`UPDATE "employee"
		 SET account_id = $1
		 WHERE id = $2
		 RETURNING id, physical_restaurant_id, name, cpf, position, account_id, created_at`,
		accountID, employeeID,
	).Scan(&emp.ID, &emp.PhysicalRestaurantID, &emp.Name, &emp.CPF, &emp.Position, &emp.AccountID, &emp.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &emp, nil
}

func (s *SQLAccountStore) VerifyAndOpenTerminalSession(ctx context.Context, branchID int, pin string, deviceID string) (*TerminalSession, error) {
	pr, err := s.GetPhysicalRestaurant(ctx, branchID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if pr.LockedUntil != nil && pr.LockedUntil.After(now) {
		return nil, ErrTerminalLocked
	}

	if pr.TerminalPIN != pin {
		attempts := pr.FailedPINAttempts + 1
		if attempts >= 5 {
			lockout := now.Add(15 * time.Minute)
			_, _ = s.DB.ExecContext(ctx,
				`UPDATE "physical_restaurant" SET failed_pin_attempts = $1, locked_until = $2 WHERE id = $3`,
				attempts, lockout, branchID,
			)
			return nil, ErrTerminalLocked
		}
		_, _ = s.DB.ExecContext(ctx,
			`UPDATE "physical_restaurant" SET failed_pin_attempts = $1 WHERE id = $2`,
			attempts, branchID,
		)
		return nil, ErrInvalidPIN
	}

	// Reset attempts on successful PIN
	_, _ = s.DB.ExecContext(ctx,
		`UPDATE "physical_restaurant" SET failed_pin_attempts = 0, locked_until = NULL WHERE id = $1`,
		branchID,
	)

	// 64-char hex session token
	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	tokenStr := hex.EncodeToString(tokenBytes)
	expiresAt := now.Add(8 * time.Hour)

	var ts TerminalSession
	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO "terminal_session" (physical_restaurant_id, token, device_id, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 RETURNING id, physical_restaurant_id, token, device_id, expires_at, created_at`,
		branchID, tokenStr, deviceID, expiresAt,
	).Scan(&ts.ID, &ts.PhysicalRestaurantID, &ts.Token, &ts.DeviceID, &ts.ExpiresAt, &ts.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &ts, nil
}

func (s *SQLAccountStore) GetTerminalSession(ctx context.Context, token string) (*TerminalSession, error) {
	var ts TerminalSession
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, physical_restaurant_id, token, device_id, expires_at, created_at
		 FROM "terminal_session"
		 WHERE token = $1 AND expires_at > NOW()`,
		token,
	).Scan(&ts.ID, &ts.PhysicalRestaurantID, &ts.Token, &ts.DeviceID, &ts.ExpiresAt, &ts.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &ts, nil
}

func (s *SQLAccountStore) RotateTerminalPIN(ctx context.Context, branchID int, newPIN string) error {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE "physical_restaurant"
		 SET terminal_pin = $1, pin_updated_at = NOW(), failed_pin_attempts = 0, locked_until = NULL
		 WHERE id = $2`,
		newPIN, branchID,
	)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------------- MEMORY ACCOUNT STORE ----------------

type MemoryAccountStore struct {
	mu               sync.RWMutex
	accounts         map[int]*Account
	byEmail          map[string]*Account
	nextAcctID       int
	sysadmins        map[int]*Sysadmin
	nextSysID        int
	managers         map[int]*Manager
	nextManID        int
	owners           map[int]*Owner
	nextOwnID        int
	branches         map[int]*PhysicalRestaurant
	nextBranchID     int
	employees        map[int]*Employee
	nextEmpID        int
	terminalSessions map[string]*TerminalSession
}

func NewMemoryAccountStore() *MemoryAccountStore {
	return &MemoryAccountStore{
		accounts:         make(map[int]*Account),
		byEmail:          make(map[string]*Account),
		nextAcctID:       1,
		sysadmins:        make(map[int]*Sysadmin),
		nextSysID:        1,
		managers:         make(map[int]*Manager),
		nextManID:        1,
		owners:           make(map[int]*Owner),
		nextOwnID:        1,
		branches:         make(map[int]*PhysicalRestaurant),
		nextBranchID:     1,
		employees:        make(map[int]*Employee),
		nextEmpID:        1,
		terminalSessions: make(map[string]*TerminalSession),
	}
}

func (m *MemoryAccountStore) CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	normEmail := strings.ToLower(strings.TrimSpace(email))
	if _, exists := m.byEmail[normEmail]; exists {
		return nil, ErrConflict
	}

	a := &Account{
		ID:        m.nextAcctID,
		Name:      name,
		Email:     normEmail,
		Password:  passwordHash,
		Phone:     phone,
		CreatedAt: time.Now(),
	}
	m.nextAcctID++
	m.accounts[a.ID] = a
	m.byEmail[normEmail] = a
	return a, nil
}

func (m *MemoryAccountStore) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	normEmail := strings.ToLower(strings.TrimSpace(email))
	a, ok := m.byEmail[normEmail]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (m *MemoryAccountStore) GetAccountByID(ctx context.Context, id int) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	a, ok := m.accounts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (m *MemoryAccountStore) ListSysadmins(ctx context.Context) ([]Sysadmin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []Sysadmin
	for _, sa := range m.sysadmins {
		list = append(list, *sa)
	}
	if list == nil {
		list = make([]Sysadmin, 0)
	}
	return list, nil
}

func (m *MemoryAccountStore) AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.accounts[accountID]; !ok {
		return nil, ErrNotFound
	}

	for _, sa := range m.sysadmins {
		if sa.AccountID == accountID {
			return sa, nil
		}
	}

	sa := &Sysadmin{
		ID:        m.nextSysID,
		AccountID: accountID,
		CreatedAt: time.Now(),
	}
	m.nextSysID++
	m.sysadmins[sa.ID] = sa
	return sa, nil
}

func (m *MemoryAccountStore) ListManagers(ctx context.Context, restaurantSlug string) ([]Manager, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []Manager
	for _, mgr := range m.managers {
		if mgr.RestaurantSlug == restaurantSlug {
			list = append(list, *mgr)
		}
	}
	if list == nil {
		list = make([]Manager, 0)
	}
	return list, nil
}

func (m *MemoryAccountStore) AddManager(ctx context.Context, restaurantSlug string, accountID int) (*Manager, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.accounts[accountID]; !ok {
		return nil, ErrNotFound
	}

	mgr := &Manager{
		ID:             m.nextManID,
		RestaurantSlug: restaurantSlug,
		AccountID:      accountID,
	}
	m.nextManID++
	m.managers[mgr.ID] = mgr
	return mgr, nil
}

func (m *MemoryAccountStore) RemoveManager(ctx context.Context, restaurantSlug string, accountID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var foundKey int
	found := false
	for k, mgr := range m.managers {
		if mgr.RestaurantSlug == restaurantSlug && mgr.AccountID == accountID {
			foundKey = k
			found = true
			break
		}
	}
	if !found {
		return ErrNotFound
	}
	delete(m.managers, foundKey)
	return nil
}

func (m *MemoryAccountStore) ListOwners(ctx context.Context, restaurantSlug string) ([]Owner, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []Owner
	for _, o := range m.owners {
		if o.RestaurantSlug == restaurantSlug {
			list = append(list, *o)
		}
	}
	if list == nil {
		list = make([]Owner, 0)
	}
	return list, nil
}

func (m *MemoryAccountStore) AddOwner(ctx context.Context, restaurantSlug string, accountID int) (*Owner, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.accounts[accountID]; !ok {
		return nil, ErrNotFound
	}

	o := &Owner{
		ID:             m.nextOwnID,
		RestaurantSlug: restaurantSlug,
		AccountID:      accountID,
	}
	m.nextOwnID++
	m.owners[o.ID] = o
	return o, nil
}

func (m *MemoryAccountStore) RemoveOwner(ctx context.Context, restaurantSlug string, accountID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	var targetKey int
	found := false
	for k, o := range m.owners {
		if o.RestaurantSlug == restaurantSlug {
			count++
			if o.AccountID == accountID {
				targetKey = k
				found = true
			}
		}
	}
	if !found {
		return ErrNotFound
	}
	if count <= 1 {
		return ErrOrphanEntity
	}
	delete(m.owners, targetKey)
	return nil
}

func (m *MemoryAccountStore) GetAccountRoles(ctx context.Context, restaurantSlug string, accountID int) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var roles []string
	for _, sa := range m.sysadmins {
		if sa.AccountID == accountID {
			roles = append(roles, "sysadmin")
			break
		}
	}

	if restaurantSlug != "" {
		for _, o := range m.owners {
			if o.RestaurantSlug == restaurantSlug && o.AccountID == accountID {
				roles = append(roles, "owner")
				break
			}
		}
		for _, mgr := range m.managers {
			if mgr.RestaurantSlug == restaurantSlug && mgr.AccountID == accountID {
				roles = append(roles, "manager")
				break
			}
		}
	}
	return roles, nil
}

func (m *MemoryAccountStore) GetRestaurantCNPJ(ctx context.Context, restaurantSlug string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return "12.345.678/0001-99", nil
}

func (m *MemoryAccountStore) CreatePhysicalRestaurant(ctx context.Context, restaurantSlug, name, cnpj, ownerCnpj string) (*PhysicalRestaurant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pr := &PhysicalRestaurant{
		ID:                m.nextBranchID,
		RestaurantSlug:    restaurantSlug,
		Name:              name,
		CNPJ:              cnpj,
		OwnerCNPJ:         ownerCnpj,
		TerminalPIN:       "123456",
		FailedPINAttempts: 0,
	}
	m.nextBranchID++
	m.branches[pr.ID] = pr
	return pr, nil
}

func (m *MemoryAccountStore) ListPhysicalRestaurants(ctx context.Context, restaurantSlug string) ([]PhysicalRestaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []PhysicalRestaurant
	for _, pr := range m.branches {
		if pr.RestaurantSlug == restaurantSlug {
			list = append(list, *pr)
		}
	}
	if list == nil {
		list = make([]PhysicalRestaurant, 0)
	}
	return list, nil
}

func (m *MemoryAccountStore) GetPhysicalRestaurant(ctx context.Context, branchID int) (*PhysicalRestaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pr, ok := m.branches[branchID]
	if !ok {
		return nil, ErrNotFound
	}
	return pr, nil
}

func (m *MemoryAccountStore) ListEmployees(ctx context.Context, branchID int) ([]Employee, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []Employee
	for _, emp := range m.employees {
		if emp.PhysicalRestaurantID == branchID {
			list = append(list, *emp)
		}
	}
	if list == nil {
		list = make([]Employee, 0)
	}
	return list, nil
}

func (m *MemoryAccountStore) CreateEmployee(ctx context.Context, branchID int, name, cpf, position string) (*Employee, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	emp := &Employee{
		ID:                   m.nextEmpID,
		PhysicalRestaurantID: branchID,
		Name:                 name,
		CPF:                  cpf,
		Position:             position,
		CreatedAt:            time.Now(),
	}
	m.nextEmpID++
	m.employees[emp.ID] = emp
	return emp, nil
}

func (m *MemoryAccountStore) LinkEmployeeAccount(ctx context.Context, employeeID int, accountID int) (*Employee, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	emp, ok := m.employees[employeeID]
	if !ok {
		return nil, ErrNotFound
	}
	if _, ok := m.accounts[accountID]; !ok {
		return nil, ErrNotFound
	}
	emp.AccountID = &accountID
	return emp, nil
}

func (m *MemoryAccountStore) VerifyAndOpenTerminalSession(ctx context.Context, branchID int, pin string, deviceID string) (*TerminalSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pr, ok := m.branches[branchID]
	if !ok {
		return nil, ErrNotFound
	}

	now := time.Now()
	if pr.LockedUntil != nil && pr.LockedUntil.After(now) {
		return nil, ErrTerminalLocked
	}

	if pr.TerminalPIN != pin {
		pr.FailedPINAttempts++
		if pr.FailedPINAttempts >= 5 {
			lockout := now.Add(15 * time.Minute)
			pr.LockedUntil = &lockout
			return nil, ErrTerminalLocked
		}
		return nil, ErrInvalidPIN
	}

	pr.FailedPINAttempts = 0
	pr.LockedUntil = nil

	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	tokenStr := hex.EncodeToString(tokenBytes)
	expiresAt := now.Add(8 * time.Hour)

	ts := &TerminalSession{
		ID:                   len(m.terminalSessions) + 1,
		PhysicalRestaurantID: branchID,
		Token:                tokenStr,
		DeviceID:             deviceID,
		ExpiresAt:            expiresAt,
		CreatedAt:            now,
	}
	m.terminalSessions[tokenStr] = ts
	return ts, nil
}

func (m *MemoryAccountStore) GetTerminalSession(ctx context.Context, token string) (*TerminalSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ts, ok := m.terminalSessions[token]
	if !ok || ts.ExpiresAt.Before(time.Now()) {
		return nil, ErrNotFound
	}
	return ts, nil
}

func (m *MemoryAccountStore) RotateTerminalPIN(ctx context.Context, branchID int, newPIN string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pr, ok := m.branches[branchID]
	if !ok {
		return ErrNotFound
	}
	pr.TerminalPIN = newPIN
	pr.FailedPINAttempts = 0
	pr.LockedUntil = nil
	return nil
}

// ---------------- MEMORY RESTAURANT STORE ----------------

type MemoryRestaurantStore struct {
	mu   sync.RWMutex
	data map[string]*Restaurant
}

func NewMemoryRestaurantStore() *MemoryRestaurantStore {
	return &MemoryRestaurantStore{data: make(map[string]*Restaurant)}
}

func (m *MemoryRestaurantStore) List(ctx context.Context) ([]Restaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]Restaurant, 0, len(m.data))
	for _, r := range m.data {
		list = append(list, *r)
	}
	return list, nil
}

func (m *MemoryRestaurantStore) GetBySlug(ctx context.Context, slug string) (*Restaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, exists := m.data[slug]
	if !exists {
		return nil, ErrNotFound
	}
	return r, nil
}

func (m *MemoryRestaurantStore) Create(ctx context.Context, r *Restaurant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[r.Slug]; exists {
		return ErrConflict
	}
	m.data[r.Slug] = r
	return nil
}

func (m *MemoryRestaurantStore) Update(ctx context.Context, slug string, input UpdateRestaurantInput) (*Restaurant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, exists := m.data[slug]
	if !exists {
		return nil, ErrNotFound
	}
	if input.Name != "" {
		r.Name = input.Name
	}
	r.Description = input.Description
	return r, nil
}

func (m *MemoryRestaurantStore) Delete(ctx context.Context, slug string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[slug]; !exists {
		return ErrNotFound
	}
	delete(m.data, slug)
	return nil
}

// ---------------- MEMORY DISH STORE ----------------

type MemoryDishStore struct {
	mu     sync.RWMutex
	dishes map[int]*Dish
	nextID int
}

func NewMemoryDishStore() *MemoryDishStore {
	return &MemoryDishStore{
		dishes: make(map[int]*Dish),
		nextID: 1,
	}
}

func (m *MemoryDishStore) ListByRestaurant(ctx context.Context, restaurantSlug string) ([]Dish, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []Dish
	for _, d := range m.dishes {
		if d.RestaurantSlug == restaurantSlug {
			list = append(list, *d)
		}
	}
	if list == nil {
		list = make([]Dish, 0)
	}
	return list, nil
}

func (m *MemoryDishStore) Create(ctx context.Context, input NewDish) (*Dish, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d := &Dish{
		ID:             m.nextID,
		RestaurantSlug: input.RestaurantSlug,
		Name:           input.Name,
		Description:    input.Description,
		Category:       input.Category,
		Tags:           input.Tags,
		Price:          input.Price,
		ImageURL:       input.ImageURL,
	}
	m.nextID++
	m.dishes[d.ID.(int)] = d
	return d, nil
}

func (m *MemoryDishStore) Update(ctx context.Context, slug string, id int, input UpdateDishInput) (*Dish, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.dishes[id]
	if !ok || d.RestaurantSlug != slug {
		return nil, ErrNotFound
	}
	if input.Name != "" {
		d.Name = input.Name
	}
	d.Price = input.Price
	if input.ImageURL != "" {
		d.ImageURL = input.ImageURL
	}
	return d, nil
}

func (m *MemoryDishStore) UpdateImage(ctx context.Context, slug string, id int, imageURL string) (*Dish, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.dishes[id]
	if !ok || d.RestaurantSlug != slug {
		return nil, ErrNotFound
	}
	d.ImageURL = imageURL
	return d, nil
}

func (m *MemoryDishStore) Delete(ctx context.Context, slug string, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.dishes[id]
	if !ok || d.RestaurantSlug != slug {
		return ErrNotFound
	}
	delete(m.dishes, id)
	return nil
}




