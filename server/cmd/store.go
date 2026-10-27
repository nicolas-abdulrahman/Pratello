package main

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource already exists")
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
	Price          float64
}

// UpdateDishInput holds the fields for modifying a dish.
type UpdateDishInput struct {
	Name  string
	Price float64
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

func (s *SQLDishStore) Create(ctx context.Context, input NewDish) (*Dish, error) {
	var dummy string
	if err := s.DB.QueryRowContext(ctx, "SELECT slug FROM restaurants WHERE slug = $1", input.RestaurantSlug).Scan(&dummy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var newID int
	err := s.DB.QueryRowContext(ctx,
		"INSERT INTO dishes (restaurant_slug, name, price) VALUES ($1, $2, $3) RETURNING id",
		input.RestaurantSlug, input.Name, input.Price,
	).Scan(&newID)
	if err != nil {
		return nil, err
	}

	return &Dish{
		ID:             newID,
		RestaurantSlug: input.RestaurantSlug,
		Name:           input.Name,
		Price:          input.Price,
	}, nil
}

func (s *SQLDishStore) Update(ctx context.Context, restaurantSlug string, dishID int, input UpdateDishInput) (*Dish, error) {
	res, err := s.DB.ExecContext(ctx,
		"UPDATE dishes SET name = $1, price = $2 WHERE id = $3 AND restaurant_slug = $4",
		input.Name, input.Price, dishID, restaurantSlug,
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
		Name:           input.Name,
		Price:          input.Price,
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


