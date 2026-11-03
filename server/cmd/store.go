package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
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

func NewSQLRestaurantStore(db *sql.DB) *SQLRestaurantStore {
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


