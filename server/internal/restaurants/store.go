package restaurants

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
)

// Store defines the database persistence contract for the restaurants domain.
type Store interface {
	ListRestaurants(ctx context.Context) ([]Restaurant, error)
	GetRestaurantBySlug(ctx context.Context, slug string) (*Restaurant, error)
	CreateRestaurant(ctx context.Context, r *Restaurant) error
	UpdateRestaurant(ctx context.Context, slug string, input UpdateRestaurantInput) (*Restaurant, error)
	DeleteRestaurant(ctx context.Context, slug string) error

	ListDishes(ctx context.Context, restaurantSlug string) ([]Dish, error)
	CreateDish(ctx context.Context, input NewDish) (*Dish, error)
	UpdateDish(ctx context.Context, restaurantSlug string, dishID int, input UpdateDishInput) (*Dish, error)
	DeleteDish(ctx context.Context, restaurantSlug string, dishID int) error
	UpdateDishImage(ctx context.Context, restaurantSlug string, dishID int, imageURL string) (*Dish, error)
}

// sqlStore implements Store backed by a SQL database (*sql.DB).
type sqlStore struct {
	db *sql.DB
}

// NewStore creates a new SQL-backed Store. If db is nil, an in-memory store is returned.
func NewStore(db *sql.DB) Store {
	if db == nil {
		return NewMemoryStore()
	}
	return &sqlStore{db: db}
}

// NewSQLiteStore is an alias for NewStore fulfilling the architectural spec.
func NewSQLiteStore(db *sql.DB) Store {
	return NewStore(db)
}

// ---------------- RESTAURANT SQL QUERIES ----------------

func (s *sqlStore) ListRestaurants(ctx context.Context) ([]Restaurant, error) {
	query := `
		SELECT slug, name, COALESCE(description, ''), COALESCE(cuisine, ''),
		       COALESCE(is_open, true), COALESCE(open_time, '07:00'), COALESCE(closing_time, '22:00'),
		       COALESCE(rating, 5.0), COALESCE(reviews_count, 0), COALESCE(delivery_fee, 0.0),
		       COALESCE(cep, ''), COALESCE(delivery_time, '20–30 min'), COALESCE(neighborhood, ''),
		       COALESCE(phone, ''), COALESCE(banner_url, ''), COALESCE(logo_url, '')
		FROM restaurants
	`
	rows, err := s.db.QueryContext(ctx, query)
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

func (s *sqlStore) GetRestaurantBySlug(ctx context.Context, slug string) (*Restaurant, error) {
	var r Restaurant
	query := `
		SELECT slug, name, COALESCE(description, ''), COALESCE(cuisine, ''),
		       COALESCE(is_open, true), COALESCE(open_time, '07:00'), COALESCE(closing_time, '22:00'),
		       COALESCE(rating, 5.0), COALESCE(reviews_count, 0), COALESCE(delivery_fee, 0.0),
		       COALESCE(cep, ''), COALESCE(delivery_time, '20–30 min'), COALESCE(neighborhood, ''),
		       COALESCE(phone, ''), COALESCE(banner_url, ''), COALESCE(logo_url, '')
		FROM restaurants WHERE slug = $1
	`
	err := s.db.QueryRowContext(ctx, query, slug).Scan(
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

func (s *sqlStore) CreateRestaurant(ctx context.Context, r *Restaurant) error {
	var existing string
	err := s.db.QueryRowContext(ctx, "SELECT slug FROM restaurants WHERE slug = $1", r.Slug).Scan(&existing)
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
	_, err = s.db.ExecContext(ctx, query,
		r.Slug, r.Name, r.Description, r.Cuisine, r.IsOpen, r.OpenTime, r.ClosingTime,
		r.Rating, r.ReviewsCount, r.DeliveryFee, r.CEP, r.DeliveryTime, r.Neighborhood,
		r.Phone, r.BannerURL, r.LogoURL,
	)
	return err
}

func (s *sqlStore) UpdateRestaurant(ctx context.Context, slug string, input UpdateRestaurantInput) (*Restaurant, error) {
	res, err := s.db.ExecContext(ctx,
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

func (s *sqlStore) DeleteRestaurant(ctx context.Context, slug string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM restaurants WHERE slug = $1", slug)
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

// ---------------- DISH SQL QUERIES ----------------

func (s *sqlStore) ListDishes(ctx context.Context, restaurantSlug string) ([]Dish, error) {
	var dummy string
	if err := s.db.QueryRowContext(ctx, "SELECT slug FROM restaurants WHERE slug = $1", restaurantSlug).Scan(&dummy); err != nil {
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
	rows, err := s.db.QueryContext(ctx, query, restaurantSlug)
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

func (s *sqlStore) CreateDish(ctx context.Context, input NewDish) (*Dish, error) {
	var dummy string
	if err := s.db.QueryRowContext(ctx, "SELECT slug FROM restaurants WHERE slug = $1", input.RestaurantSlug).Scan(&dummy); err != nil {
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
	err := s.db.QueryRowContext(ctx, query,
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

func (s *sqlStore) UpdateDish(ctx context.Context, restaurantSlug string, dishID int, input UpdateDishInput) (*Dish, error) {
	res, err := s.db.ExecContext(ctx,
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
	err = s.db.QueryRowContext(ctx,
		"SELECT id, name, price, COALESCE(image_url, '') FROM dishes WHERE id = $1 AND restaurant_slug = $2",
		dishID, restaurantSlug,
	).Scan(&d.ID, &d.Name, &d.Price, &d.ImageURL)
	if err != nil {
		return nil, err
	}
	d.RestaurantSlug = restaurantSlug
	return &d, nil
}

func (s *sqlStore) UpdateDishImage(ctx context.Context, restaurantSlug string, dishID int, imageURL string) (*Dish, error) {
	res, err := s.db.ExecContext(ctx,
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
	err = s.db.QueryRowContext(ctx,
		"SELECT id, name, price, COALESCE(image_url, '') FROM dishes WHERE id = $1 AND restaurant_slug = $2",
		dishID, restaurantSlug,
	).Scan(&d.ID, &d.Name, &d.Price, &d.ImageURL)
	if err != nil {
		return nil, err
	}
	d.RestaurantSlug = restaurantSlug
	return &d, nil
}

func (s *sqlStore) DeleteDish(ctx context.Context, restaurantSlug string, dishID int) error {
	res, err := s.db.ExecContext(ctx,
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

// ---------------- IN-MEMORY STORE IMPLEMENTATION ----------------

type memoryStore struct {
	mu          sync.RWMutex
	restaurants map[string]*Restaurant
	dishes      map[int]*Dish
	nextDishID  int
}

// NewMemoryStore constructs a purely thread-safe in-memory store for unit testing.
func NewMemoryStore() Store {
	return &memoryStore{
		restaurants: make(map[string]*Restaurant),
		dishes:      make(map[int]*Dish),
		nextDishID:  1,
	}
}

func (m *memoryStore) ListRestaurants(ctx context.Context) ([]Restaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]Restaurant, 0, len(m.restaurants))
	for _, r := range m.restaurants {
		list = append(list, *r)
	}
	return list, nil
}

func (m *memoryStore) GetRestaurantBySlug(ctx context.Context, slug string) (*Restaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, exists := m.restaurants[slug]
	if !exists {
		return nil, ErrNotFound
	}
	return r, nil
}

func (m *memoryStore) CreateRestaurant(ctx context.Context, r *Restaurant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.restaurants[r.Slug]; exists {
		return ErrConflict
	}
	m.restaurants[r.Slug] = r
	return nil
}

func (m *memoryStore) UpdateRestaurant(ctx context.Context, slug string, input UpdateRestaurantInput) (*Restaurant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, exists := m.restaurants[slug]
	if !exists {
		return nil, ErrNotFound
	}
	if input.Name != "" {
		r.Name = input.Name
	}
	r.Description = input.Description
	return r, nil
}

func (m *memoryStore) DeleteRestaurant(ctx context.Context, slug string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.restaurants[slug]; !exists {
		return ErrNotFound
	}
	delete(m.restaurants, slug)
	return nil
}

func (m *memoryStore) ListDishes(ctx context.Context, restaurantSlug string) ([]Dish, error) {
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

func (m *memoryStore) CreateDish(ctx context.Context, input NewDish) (*Dish, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d := &Dish{
		ID:             m.nextDishID,
		RestaurantSlug: input.RestaurantSlug,
		Name:           input.Name,
		Description:    input.Description,
		Category:       input.Category,
		Tags:           input.Tags,
		Price:          input.Price,
		ImageURL:       input.ImageURL,
	}
	m.nextDishID++
	m.dishes[d.ID.(int)] = d
	return d, nil
}

func (m *memoryStore) UpdateDish(ctx context.Context, slug string, id int, input UpdateDishInput) (*Dish, error) {
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

func (m *memoryStore) UpdateDishImage(ctx context.Context, slug string, id int, imageURL string) (*Dish, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.dishes[id]
	if !ok || d.RestaurantSlug != slug {
		return nil, ErrNotFound
	}
	d.ImageURL = imageURL
	return d, nil
}

func (m *memoryStore) DeleteDish(ctx context.Context, slug string, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.dishes[id]
	if !ok || d.RestaurantSlug != slug {
		return ErrNotFound
	}
	delete(m.dishes, id)
	return nil
}
