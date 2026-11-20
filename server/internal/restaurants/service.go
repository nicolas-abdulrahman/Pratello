package restaurants

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	errInvalid     = errors.New("invalid input")
	errForbidden   = errors.New("forbidden")
	errMissing     = errors.New("resource not found")
	errConflict    = errors.New("resource conflicts with existing data")
	errLastOwner   = errors.New("cannot remove the last owner")
	errUnavailable = errors.New("restaurant database unavailable")
)

type operation string

const (
	listOrg          operation = "restaurant.list"
	myOrg            operation = "restaurant.mine"
	getOrg           operation = "restaurant.get"
	createOrg        operation = "restaurant.create"
	updateOrg        operation = "restaurant.update"
	archiveOrg       operation = "restaurant.archive"
	listBranch       operation = "branch.list"
	getBranch        operation = "branch.get"
	createBranch     operation = "branch.create"
	updateBranch     operation = "branch.update"
	archiveBranch    operation = "branch.archive"
	verifyBranch     operation = "branch.verify"
	setAddress       operation = "branch.address.update"
	listPhones       operation = "branch.phone.list"
	addPhone         operation = "branch.phone.add"
	removePhone      operation = "branch.phone.remove"
	getHours         operation = "branch.hours.get"
	setHours         operation = "branch.hours.update"
	listTables       operation = "branch.table.list"
	addTable         operation = "branch.table.add"
	updateTable      operation = "branch.table.update"
	removeTable      operation = "branch.table.remove"
	listOwners       operation = "restaurant.owner.list"
	addOwner         operation = "restaurant.owner.add"
	removeOwner      operation = "restaurant.owner.remove"
	publishedCatalog operation = "restaurant.catalog.published"
	assignManager    operation = "branch.manager.assign"
	revokeManager    operation = "branch.manager.revoke"
	listOrgManagers  operation = "restaurant.manager.list"
	addOrgManager    operation = "restaurant.manager.add"
	removeOrgManager operation = "restaurant.manager.remove"
)

type address struct {
	CEP        string `json:"cep"`
	Street     string `json:"street"`
	Number     string `json:"number"`
	Apt        string `json:"apt,omitempty"`
	Complement string `json:"complement,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
}

type details struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type newOrg struct {
	details
	Slug string `json:"slug,omitempty"`
}

type branchInput struct {
	Name      string  `json:"name"`
	CNPJ      string  `json:"cnpj"`
	OwnerCNPJ string  `json:"owner_cnpj"`
	Address   address `json:"address"`
}

type phoneInput struct {
	Number string `json:"phone_number"`
	Type   string `json:"phone_type"`
}

type hour struct {
	Weekday *int    `json:"weekday"`
	Opens   *string `json:"opens_at"`
	Closes  *string `json:"closes_at"`
	Closed  *bool   `json:"is_closed"`
}

type tableInput struct {
	Number int   `json:"number"`
	Seats  *int  `json:"seats"`
	Order  int   `json:"display_order"`
	Active *bool `json:"is_active"`
}

type accountInput struct {
	ID int `json:"account_id"`
}

type request struct {
	op                                operation
	slug                              string
	actor, branch, target, page, size int
	body                              any
}

type access struct {
	orgID                                  int
	orgRole                                string
	manager, employee, sysadmin, published bool
}

func (a access) managesOrg() bool { return a.orgRole == "OWNER" || a.orgRole == "MANAGER" }

func (a access) managesBranch() bool { return a.managesOrg() || a.manager }

type repository interface {
	transact(context.Context, bool, func(transaction) error) error
}

type transaction interface {
	scope(context.Context, request, bool) (access, error)
	execute(context.Context, request, access) (any, error)
	audit(context.Context, request, access, any) error
}

type service struct{ repo repository }

func (s service) run(ctx context.Context, r request) (any, error) {
	if err := validate(&r); err != nil {
		return nil, err
	}
	write := isWrite(r.op)
	var result any
	err := s.repo.transact(ctx, write, func(tx transaction) error {
		a, err := tx.scope(ctx, r, write)
		if err != nil {
			return err
		}
		if err = authorize(r, a); err != nil {
			return err
		}
		result, err = tx.execute(ctx, r, a)
		if err != nil || !write {
			return err
		}
		return tx.audit(ctx, r, a, result)
	})
	return result, err
}

func isWrite(op operation) bool {
	switch op {
	case listOrg, myOrg, getOrg, listBranch, getBranch, listPhones, getHours, listTables, listOwners, listOrgManagers:
		return false
	default:
		return true
	}
}

func authorize(r request, a access) error {
	allowed := false
	switch r.op {
	case listOrg, getOrg, listBranch:
		return nil
	case createOrg, myOrg:
		allowed = r.actor > 0
	case getBranch, getHours:
		if a.published || a.managesBranch() || a.sysadmin {
			return nil
		}
		return errMissing
	case updateOrg, createBranch, archiveBranch, assignManager, revokeManager:
		allowed = a.managesOrg()
	case listOwners, listOrgManagers:
		allowed = a.managesOrg()
	case archiveOrg, addOwner, removeOwner, addOrgManager, removeOrgManager:
		allowed = a.orgRole == "OWNER"
	case verifyBranch:
		allowed = a.sysadmin
	case listPhones, listTables:
		allowed = a.managesBranch() || a.employee
	case updateBranch, setAddress, addPhone, removePhone, setHours, addTable, updateTable, removeTable:
		allowed = a.managesBranch()
	}
	if !allowed {
		return errForbidden
	}
	return nil
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

var slugSeparator = regexp.MustCompile(`[^a-z0-9]+`)

var cnpjPattern = regexp.MustCompile(`^[0-9]{14}$`)

var cepPattern = regexp.MustCompile(`^[0-9]{8}$`)

var phonePattern = regexp.MustCompile(`^\+?[0-9]{8,15}$`)

func text(s string, max int, required bool) bool {
	return (!required || strings.TrimSpace(s) != "") && utf8.RuneCountInString(s) <= max && !strings.ContainsRune(s, 0)
}

func validAddress(a address) bool {
	return cepPattern.MatchString(a.CEP) && text(a.Street, 255, true) && text(a.Number, 20, true) && text(a.Apt, 50, false) && text(a.Complement, 255, false) && text(a.City, 255, false) && text(a.State, 2, false)
}

func validate(r *request) error {
	valid := true
	switch b := r.body.(type) {
	case *newOrg:
		b.Name = strings.TrimSpace(b.Name)
		if b.Slug == "" {
			b.Slug = strings.Trim(slugSeparator.ReplaceAllString(strings.ToLower(b.Name), "-"), "-")
		}
		valid = slugPattern.MatchString(b.Slug) && len(b.Slug) <= 255 && text(b.Name, 255, true) && text(b.Description, 10000, false)
	case *details:
		b.Name = strings.TrimSpace(b.Name)
		valid = text(b.Name, 255, true) && text(b.Description, 10000, false)
	case *branchInput:
		b.Name = strings.TrimSpace(b.Name)
		valid = text(b.Name, 255, true) && cnpjPattern.MatchString(b.CNPJ) && cnpjPattern.MatchString(b.OwnerCNPJ) && validAddress(b.Address)
	case *address:
		valid = validAddress(*b)
	case *phoneInput:
		if b.Type == "" {
			b.Type = "mobile"
		}
		valid = phonePattern.MatchString(b.Number) && (b.Type == "mobile" || b.Type == "landline" || b.Type == "whatsapp" || b.Type == "other")
	case *[]hour:
		valid = len(*b) == 7
		seen := [7]bool{}
		for _, h := range *b {
			if h.Weekday == nil || *h.Weekday < 0 || *h.Weekday > 6 || seen[*h.Weekday] || h.Closed == nil {
				return errInvalid
			}
			seen[*h.Weekday] = true
			if *h.Closed {
				if h.Opens != nil || h.Closes != nil {
					return errInvalid
				}
				continue
			}
			if h.Opens == nil || h.Closes == nil || *h.Opens == *h.Closes {
				return errInvalid
			}
			for _, value := range []string{*h.Opens, *h.Closes} {
				if _, err := time.Parse("15:04", value); err != nil || len(value) != 5 {
					return errInvalid
				}
			}
		}
	case *tableInput:
		valid = b.Number > 0 && b.Number <= 2147483647 && (b.Seats == nil || (*b.Seats > 0 && *b.Seats <= 2147483647)) && b.Order >= 0 && b.Order <= 2147483647
	case *accountInput:
		valid = b.ID > 0 && b.ID <= 2147483647
	}
	if !valid {
		return errInvalid
	}
	return nil
}

// Domain Errors: Exported as vars so http.go or callers can match with errors.Is.
var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation failed")
)

// Restaurant represents a logical restaurant entity.
type Restaurant struct {
	Slug         string
	Name         string
	Description  string
	Cuisine      string
	IsOpen       bool
	OpenTime     string
	ClosingTime  string
	Rating       float64
	ReviewsCount int
	DeliveryFee  float64
	CEP          string
	DeliveryTime string
	Neighborhood string
	Phone        string
	BannerURL    string
	LogoURL      string
	MenuItems    []string
}

// Dish represents a menu item belonging to a restaurant.
type Dish struct {
	ID             any
	RestaurantSlug string
	Name           string
	Description    string
	Category       string
	Tags           []string
	Price          float64
	ImageURL       string
}

// NewDish holds the business fields required to create a dish.
type NewDish struct {
	RestaurantSlug string
	Name           string
	Description    string
	Category       string
	Tags           []string
	Price          float64
	ImageURL       string
}

// UpdateDishInput holds modifications for an existing dish.
type UpdateDishInput struct {
	Name        string
	Description string
	Category    string
	Tags        []string
	Price       float64
	ImageURL    string
}

// store defines the consumer-side interface for restaurant persistence.
type catalogStore interface {
	GetRestaurantBySlug(ctx context.Context, slug string) (*Restaurant, error)

	ListDishes(ctx context.Context, restaurantSlug string) ([]Dish, error)
	CreateDish(ctx context.Context, input NewDish) (*Dish, error)
	UpdateDish(ctx context.Context, restaurantSlug string, dishID int, input UpdateDishInput) (*Dish, error)
	DeleteDish(ctx context.Context, restaurantSlug string, dishID int) error
	UpdateDishImage(ctx context.Context, restaurantSlug string, dishID int, imageURL string) (*Dish, error)
}

type Service struct{ store catalogStore }

// ListDishes retrieves all dishes belonging to a restaurant.
func (s *Service) ListDishes(ctx context.Context, restaurantSlug string) ([]Dish, error) {
	restaurantSlug = strings.ToLower(strings.TrimSpace(restaurantSlug))
	return s.store.ListDishes(ctx, restaurantSlug)
}

// AddDish validates price, name, and restaurant existence before adding.
func (s *Service) AddDish(ctx context.Context, input NewDish) (*Dish, error) {
	name := strings.TrimSpace(input.Name)
	if len(name) < 2 {
		return nil, errors.Join(ErrValidation, errors.New("dish name must have at least 2 characters"))
	}
	if input.Price <= 0 {
		return nil, errors.Join(ErrValidation, errors.New("dish price must be greater than zero"))
	}

	slug := strings.ToLower(strings.TrimSpace(input.RestaurantSlug))
	if slug == "" {
		return nil, errors.Join(ErrValidation, errors.New("restaurant slug is required"))
	}

	// Verify restaurant exists
	_, err := s.store.GetRestaurantBySlug(ctx, slug)
	if err != nil {
		return nil, ErrNotFound
	}

	return s.store.CreateDish(ctx, NewDish{
		RestaurantSlug: slug,
		Name:           name,
		Description:    strings.TrimSpace(input.Description),
		Category:       strings.TrimSpace(input.Category),
		Tags:           input.Tags,
		Price:          input.Price,
		ImageURL:       strings.TrimSpace(input.ImageURL),
	})
}

// UpdateDish updates dish information.
func (s *Service) UpdateDish(ctx context.Context, slug string, id int, input UpdateDishInput) (*Dish, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	return s.store.UpdateDish(ctx, slug, id, input)
}

// DeleteDish removes a dish from the menu.
func (s *Service) DeleteDish(ctx context.Context, slug string, id int) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	return s.store.DeleteDish(ctx, slug, id)
}

// UpdateDishImage updates the image URL for a dish.
func (s *Service) UpdateDishImage(ctx context.Context, slug string, id int, imageURL string) (*Dish, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	return s.store.UpdateDishImage(ctx, slug, id, imageURL)
}

// GetRestaurantMenu returns the restaurant profile, ordered unique categories ("Mais Pedidos" first), and dishes.
func (s *Service) GetRestaurantMenu(ctx context.Context, slug string) (*Restaurant, []string, []Dish, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	res, err := s.store.GetRestaurantBySlug(ctx, slug)
	if err != nil {
		return nil, nil, nil, err
	}

	dishes, err := s.store.ListDishes(ctx, slug)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, nil, nil, err
	}
	if dishes == nil {
		dishes = make([]Dish, 0)
	}

	catSet := make(map[string]bool)
	var categories []string
	categories = append(categories, "Mais Pedidos")
	catSet["Mais Pedidos"] = true

	for _, d := range dishes {
		cat := d.Category
		if cat == "" {
			cat = "Pizzas"
		}
		if !catSet[cat] {
			catSet[cat] = true
			categories = append(categories, cat)
		}
	}

	return res, categories, dishes, nil
}
