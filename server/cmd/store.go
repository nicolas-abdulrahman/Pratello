package main

import (
	"context"
	"database/sql"
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource already exists")
)

// RestaurantStore defines the contract for restaurant data persistence.
type RestaurantStore interface {
	List(ctx context.Context) ([]Restaurant, error)
	GetBySlug(ctx context.Context, slug string) (*Restaurant, error)
	Create(ctx context.Context, r *Restaurant) error
	Update(ctx context.Context, slug string, req UpdateRestaurantRequest) (*Restaurant, error)
	Delete(ctx context.Context, slug string) error
}

// DishStore defines the contract for dish data persistence.
type DishStore interface {
	ListByRestaurant(ctx context.Context, restaurantSlug string) ([]Dish, error)
	Create(ctx context.Context, restaurantSlug string, req CreateDishRequest) (*Dish, error)
	Update(ctx context.Context, restaurantSlug string, dishID int, req UpdateDishRequest) (*Dish, error)
	Delete(ctx context.Context, restaurantSlug string, dishID int) error
}

// SQLRestaurantStore implements RestaurantStore backed by a SQL database.
type SQLRestaurantStore struct {
	DB *sql.DB
}

func NewSQLRestaurantStore(db *sql.DB) *SQLRestaurantStore {
	return &SQLRestaurantStore{DB: db}
}

func (s *SQLRestaurantStore) List(ctx context.Context) ([]Restaurant, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT slug, name, COALESCE(description, '') FROM restaurants")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Restaurant
	for rows.Next() {
		var r Restaurant
		if err := rows.Scan(&r.Slug, &r.Name, &r.Description); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s *SQLRestaurantStore) GetBySlug(ctx context.Context, slug string) (*Restaurant, error) {
	var r Restaurant
	err := s.DB.QueryRowContext(ctx,
		"SELECT slug, name, COALESCE(description, '') FROM restaurants WHERE slug = $1",
		slug,
	).Scan(&r.Slug, &r.Name, &r.Description)
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

	_, err = s.DB.ExecContext(ctx,
		"INSERT INTO restaurants (slug, name, description) VALUES ($1, $2, $3)",
		r.Slug, r.Name, r.Description,
	)
	return err
}

func (s *SQLRestaurantStore) Update(ctx context.Context, slug string, req UpdateRestaurantRequest) (*Restaurant, error) {
	res, err := s.DB.ExecContext(ctx,
		"UPDATE restaurants SET name = $1, description = $2 WHERE slug = $3",
		req.Name, req.Description, slug,
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
		Name:        req.Name,
		Description: req.Description,
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

func NewSQLDishStore(db *sql.DB) *SQLDishStore {
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

	rows, err := s.DB.QueryContext(ctx,
		"SELECT id, name, price FROM dishes WHERE restaurant_slug = $1 ORDER BY id",
		restaurantSlug,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Dish
	for rows.Next() {
		var d Dish
		if err := rows.Scan(&d.ID, &d.Name, &d.Price); err != nil {
			return nil, err
		}
		d.RestaurantSlug = restaurantSlug
		result = append(result, d)
	}
	return result, rows.Err()
}

func (s *SQLDishStore) Create(ctx context.Context, restaurantSlug string, req CreateDishRequest) (*Dish, error) {
	var dummy string
	if err := s.DB.QueryRowContext(ctx, "SELECT slug FROM restaurants WHERE slug = $1", restaurantSlug).Scan(&dummy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var newID int
	err := s.DB.QueryRowContext(ctx,
		"INSERT INTO dishes (restaurant_slug, name, price) VALUES ($1, $2, $3) RETURNING id",
		restaurantSlug, req.Name, req.Price,
	).Scan(&newID)
	if err != nil {
		return nil, err
	}

	return &Dish{
		ID:             newID,
		RestaurantSlug: restaurantSlug,
		Name:           req.Name,
		Price:          req.Price,
	}, nil
}

func (s *SQLDishStore) Update(ctx context.Context, restaurantSlug string, dishID int, req UpdateDishRequest) (*Dish, error) {
	res, err := s.DB.ExecContext(ctx,
		"UPDATE dishes SET name = $1, price = $2 WHERE id = $3 AND restaurant_slug = $4",
		req.Name, req.Price, dishID, restaurantSlug,
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

	return &Dish{
		ID:             dishID,
		RestaurantSlug: restaurantSlug,
		Name:           req.Name,
		Price:          req.Price,
	}, nil
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

// MemoryRestaurantStore provides a thread-safe in-memory store for RestaurantStore.
type MemoryRestaurantStore struct {
	mu          sync.RWMutex
	restaurants map[string]Restaurant
}

func NewMemoryRestaurantStore() *MemoryRestaurantStore {
	return &MemoryRestaurantStore{
		restaurants: make(map[string]Restaurant),
	}
}

func (m *MemoryRestaurantStore) List(ctx context.Context) ([]Restaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]Restaurant, 0, len(m.restaurants))
	for _, r := range m.restaurants {
		list = append(list, r)
	}
	return list, nil
}

func (m *MemoryRestaurantStore) GetBySlug(ctx context.Context, slug string) (*Restaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, exists := m.restaurants[slug]
	if !exists {
		return nil, ErrNotFound
	}
	return &r, nil
}

func (m *MemoryRestaurantStore) Create(ctx context.Context, r *Restaurant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.restaurants[r.Slug]; exists {
		return ErrConflict
	}
	m.restaurants[r.Slug] = *r
	return nil
}

func (m *MemoryRestaurantStore) Update(ctx context.Context, slug string, req UpdateRestaurantRequest) (*Restaurant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, exists := m.restaurants[slug]
	if !exists {
		return nil, ErrNotFound
	}
	if req.Name != "" {
		r.Name = req.Name
	}
	r.Description = req.Description
	m.restaurants[slug] = r
	return &r, nil
}

func (m *MemoryRestaurantStore) Delete(ctx context.Context, slug string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.restaurants[slug]; !exists {
		return ErrNotFound
	}
	delete(m.restaurants, slug)
	return nil
}

// MemoryDishStore provides a thread-safe in-memory store for DishStore.
type MemoryDishStore struct {
	mu      sync.RWMutex
	counter int
	dishes  map[int]Dish
}

func NewMemoryDishStore() *MemoryDishStore {
	return &MemoryDishStore{
		dishes: make(map[int]Dish),
	}
}

func (m *MemoryDishStore) ListByRestaurant(ctx context.Context, restaurantSlug string) ([]Dish, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []Dish
	for _, d := range m.dishes {
		if d.RestaurantSlug == restaurantSlug {
			list = append(list, d)
		}
	}
	return list, nil
}

func (m *MemoryDishStore) Create(ctx context.Context, restaurantSlug string, req CreateDishRequest) (*Dish, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counter++
	dish := Dish{
		ID:             m.counter,
		RestaurantSlug: restaurantSlug,
		Name:           req.Name,
		Price:          req.Price,
	}
	m.dishes[dish.ID] = dish
	return &dish, nil
}

func (m *MemoryDishStore) Update(ctx context.Context, restaurantSlug string, dishID int, req UpdateDishRequest) (*Dish, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	dish, exists := m.dishes[dishID]
	if !exists || dish.RestaurantSlug != restaurantSlug {
		return nil, ErrNotFound
	}
	if req.Name != "" {
		dish.Name = req.Name
	}
	if req.Price >= 0 {
		dish.Price = req.Price
	}
	m.dishes[dishID] = dish
	return &dish, nil
}

func (m *MemoryDishStore) Delete(ctx context.Context, restaurantSlug string, dishID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	dish, exists := m.dishes[dishID]
	if !exists || dish.RestaurantSlug != restaurantSlug {
		return ErrNotFound
	}
	delete(m.dishes, dishID)
	return nil
}
