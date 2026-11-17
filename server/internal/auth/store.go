package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// -------------------------------------------------------------
// Domain Entities (Database Representation)
// -------------------------------------------------------------

type Account struct {
	ID        int
	Name      string
	Email     string
	Password  string
	Phone     string
	CreatedAt time.Time
}

type Sysadmin struct {
	ID        int
	AccountID int
	CreatedAt time.Time
}

type Manager struct {
	ID             int
	RestaurantSlug string
	AccountID      int
}

type Owner struct {
	ID             int
	RestaurantSlug string
	AccountID      int
}

type PhysicalRestaurant struct {
	ID                int
	RestaurantID      int
	RestaurantSlug    string
	AddressID         int
	Name              string
	CNPJ              string
	OwnerCNPJ         string
	TerminalPIN       string
	FailedPINAttempts int
	LockedUntil       *time.Time
}

type Employee struct {
	ID                   int
	PhysicalRestaurantID int
	Name                 string
	CPF                  string
	Position             string
	AccountID            *int
	CreatedAt            time.Time
}

type TerminalSession struct {
	ID                   int
	PhysicalRestaurantID int
	Token                string
	DeviceID             string
	ExpiresAt            time.Time
	CreatedAt            time.Time
}

type TerminalOrder struct {
	ID                   int
	PhysicalRestaurantID int
	AccountID            int
	OrderType            string
	Status               string
	Subtotal             float64
	Discount             float64
	Total                float64
	CreatedAt            time.Time
}

// -------------------------------------------------------------
// Store Interface
// -------------------------------------------------------------

type Store interface {
	CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountByID(ctx context.Context, id int) (*Account, error)

	ListSysadmins(ctx context.Context) ([]Sysadmin, error)
	AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error)

	ListManagers(ctx context.Context, restaurantSlug string) ([]Manager, error)
	AddManager(ctx context.Context, restaurantSlug string, accountID int) (*Manager, error)
	RemoveManager(ctx context.Context, restaurantSlug string, accountID int) error

	ListOwners(ctx context.Context, restaurantSlug string) ([]Owner, error)
	AddOwner(ctx context.Context, restaurantSlug string, accountID int) (*Owner, error)
	RemoveOwner(ctx context.Context, restaurantSlug string, accountID int) error
	GetAccountRoles(ctx context.Context, restaurantSlug string, accountID int) ([]string, error)

	CreatePhysicalRestaurant(ctx context.Context, restaurantSlug, name, cnpj, ownerCnpj string) (*PhysicalRestaurant, error)
	ListPhysicalRestaurants(ctx context.Context, restaurantSlug string) ([]PhysicalRestaurant, error)
	GetPhysicalRestaurant(ctx context.Context, branchID int) (*PhysicalRestaurant, error)
	GetRestaurantCNPJ(ctx context.Context, restaurantSlug string) (string, error)

	ListEmployees(ctx context.Context, branchID int) ([]Employee, error)
	CreateEmployee(ctx context.Context, branchID int, name, cpf, position string) (*Employee, error)
	LinkEmployeeAccount(ctx context.Context, employeeID int, accountID int) (*Employee, error)

	VerifyAndOpenTerminalSession(ctx context.Context, branchID int, pin string, deviceID string) (*TerminalSession, error)
	GetTerminalSession(ctx context.Context, token string) (*TerminalSession, error)
	RotateTerminalPIN(ctx context.Context, branchID int, newPIN string) error
	GetTerminalOrders(ctx context.Context, branchID int) ([]TerminalOrder, error)
}

// -------------------------------------------------------------
// Constructors
// -------------------------------------------------------------

func NewStore(db *sql.DB) Store {
	if db == nil {
		return NewMemoryStore()
	}
	return NewSQLStore(db)
}

func NewSQLiteStore(db *sql.DB) Store {
	return NewStore(db)
}

func NewSQLStore(db *sql.DB) *sqlStore {
	return &sqlStore{db: db}
}

// -------------------------------------------------------------
// SQL Implementation
// -------------------------------------------------------------

type sqlStore struct {
	db *sql.DB
}

func (s *sqlStore) CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error) {
	var a Account
	err := s.db.QueryRowContext(ctx,
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

func (s *sqlStore) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	var a Account
	err := s.db.QueryRowContext(ctx,
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

func (s *sqlStore) GetAccountByID(ctx context.Context, id int) (*Account, error) {
	var a Account
	err := s.db.QueryRowContext(ctx,
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

func (s *sqlStore) ListSysadmins(ctx context.Context) ([]Sysadmin, error) {
	rows, err := s.db.QueryContext(ctx,
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

func (s *sqlStore) AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error) {
	if _, err := s.GetAccountByID(ctx, accountID); err != nil {
		return nil, err
	}

	var sa Sysadmin
	err := s.db.QueryRowContext(ctx,
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

func (s *sqlStore) getRestaurantID(ctx context.Context, slug string) (int, error) {
	var id int
	err := s.db.QueryRowContext(ctx, `SELECT id FROM "restaurant" WHERE name_slug = $1`, slug).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	var legacyName string
	err = s.db.QueryRowContext(ctx, `SELECT name FROM restaurants WHERE slug = $1`, slug).Scan(&legacyName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	err = s.db.QueryRowContext(ctx,
		`INSERT INTO "restaurant" (name_slug, name) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING
		 RETURNING id`,
		slug, legacyName,
	).Scan(&id)
	if err != nil {
		_ = s.db.QueryRowContext(ctx, `SELECT id FROM "restaurant" WHERE name_slug = $1`, slug).Scan(&id)
	}
	if id == 0 {
		return 0, ErrNotFound
	}
	return id, nil
}

func (s *sqlStore) ListManagers(ctx context.Context, restaurantSlug string) ([]Manager, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx,
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

func (s *sqlStore) AddManager(ctx context.Context, restaurantSlug string, accountID int) (*Manager, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	if _, err := s.GetAccountByID(ctx, accountID); err != nil {
		return nil, err
	}

	var id int
	err = s.db.QueryRowContext(ctx,
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

func (s *sqlStore) RemoveManager(ctx context.Context, restaurantSlug string, accountID int) error {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return err
	}

	res, err := s.db.ExecContext(ctx,
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

func (s *sqlStore) ListOwners(ctx context.Context, restaurantSlug string) ([]Owner, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx,
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

func (s *sqlStore) AddOwner(ctx context.Context, restaurantSlug string, accountID int) (*Owner, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	if _, err := s.GetAccountByID(ctx, accountID); err != nil {
		return nil, err
	}

	var id int
	err = s.db.QueryRowContext(ctx,
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

func (s *sqlStore) RemoveOwner(ctx context.Context, restaurantSlug string, accountID int) error {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return err
	}

	var count int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM "account_restaurant" WHERE restaurant_id = $1 AND account_type = 'owner'`,
		restID,
	).Scan(&count); err != nil {
		return err
	}
	if count <= 1 {
		return ErrCannotRemoveLastOwner
	}

	res, err := s.db.ExecContext(ctx,
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

func (s *sqlStore) GetAccountRoles(ctx context.Context, restaurantSlug string, accountID int) ([]string, error) {
	var roles []string

	var saCount int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM "sysadmin" WHERE account_id = $1`, accountID).Scan(&saCount)
	if saCount > 0 {
		roles = append(roles, "sysadmin")
	}

	if restaurantSlug != "" {
		restID, err := s.getRestaurantID(ctx, restaurantSlug)
		if err == nil {
			rows, err := s.db.QueryContext(ctx,
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

func (s *sqlStore) GetRestaurantCNPJ(ctx context.Context, restaurantSlug string) (string, error) {
	var cnpj sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT cnpj FROM "restaurant" WHERE name_slug = $1`,
		restaurantSlug,
	).Scan(&cnpj)
	if err == nil && cnpj.Valid && cnpj.String != "" {
		return cnpj.String, nil
	}

	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(cep, '') FROM restaurants WHERE slug = $1`,
		restaurantSlug,
	).Scan(&cnpj)
	if err == nil && cnpj.Valid {
		return cnpj.String, nil
	}
	return "", nil
}

func (s *sqlStore) CreatePhysicalRestaurant(ctx context.Context, restaurantSlug, name, cnpj, ownerCnpj string) (*PhysicalRestaurant, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	var addressID int
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO "Address" ("CEP", number, complementary) VALUES ('00000-000', 0, 'Sede') RETURNING id`,
	).Scan(&addressID)
	if err != nil {
		_ = s.db.QueryRowContext(ctx, `SELECT id FROM "Address" LIMIT 1`).Scan(&addressID)
		if addressID == 0 {
			addressID = 1
		}
	}

	var pr PhysicalRestaurant
	err = s.db.QueryRowContext(ctx,
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

func (s *sqlStore) ListPhysicalRestaurants(ctx context.Context, restaurantSlug string) ([]PhysicalRestaurant, error) {
	restID, err := s.getRestaurantID(ctx, restaurantSlug)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx,
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

func (s *sqlStore) GetPhysicalRestaurant(ctx context.Context, branchID int) (*PhysicalRestaurant, error) {
	var pr PhysicalRestaurant
	err := s.db.QueryRowContext(ctx,
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

func (s *sqlStore) ListEmployees(ctx context.Context, branchID int) ([]Employee, error) {
	rows, err := s.db.QueryContext(ctx,
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

func (s *sqlStore) CreateEmployee(ctx context.Context, branchID int, name, cpf, position string) (*Employee, error) {
	var emp Employee
	err := s.db.QueryRowContext(ctx,
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

func (s *sqlStore) LinkEmployeeAccount(ctx context.Context, employeeID int, accountID int) (*Employee, error) {
	var emp Employee
	err := s.db.QueryRowContext(ctx,
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

func (s *sqlStore) VerifyAndOpenTerminalSession(ctx context.Context, branchID int, pin string, deviceID string) (*TerminalSession, error) {
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
			_, _ = s.db.ExecContext(ctx,
				`UPDATE "physical_restaurant" SET failed_pin_attempts = $1, locked_until = $2 WHERE id = $3`,
				attempts, lockout, branchID,
			)
			return nil, ErrTerminalLocked
		}
		_, _ = s.db.ExecContext(ctx,
			`UPDATE "physical_restaurant" SET failed_pin_attempts = $1 WHERE id = $2`,
			attempts, branchID,
		)
		return nil, ErrInvalidPIN
	}

	_, _ = s.db.ExecContext(ctx,
		`UPDATE "physical_restaurant" SET failed_pin_attempts = 0, locked_until = NULL WHERE id = $1`,
		branchID,
	)

	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	tokenStr := hex.EncodeToString(tokenBytes)
	expiresAt := now.Add(8 * time.Hour)

	var ts TerminalSession
	err = s.db.QueryRowContext(ctx,
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

func (s *sqlStore) GetTerminalSession(ctx context.Context, token string) (*TerminalSession, error) {
	var ts TerminalSession
	err := s.db.QueryRowContext(ctx,
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

func (s *sqlStore) RotateTerminalPIN(ctx context.Context, branchID int, newPIN string) error {
	res, err := s.db.ExecContext(ctx,
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

func (s *sqlStore) GetTerminalOrders(ctx context.Context, branchID int) ([]TerminalOrder, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, physical_restaurant_id, account_id, order_type, status, subtotal, discount, total, created_at
		 FROM "order"
		 WHERE physical_restaurant_id = $1
		 ORDER BY id DESC`,
		branchID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []TerminalOrder
	for rows.Next() {
		var o TerminalOrder
		if err := rows.Scan(&o.ID, &o.PhysicalRestaurantID, &o.AccountID, &o.OrderType, &o.Status, &o.Subtotal, &o.Discount, &o.Total, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = make([]TerminalOrder, 0)
	}
	return orders, rows.Err()
}

// -------------------------------------------------------------
// Memory Implementation (Fast isolated tests & zero-db fallback)
// -------------------------------------------------------------

type memoryStore struct {
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

func NewMemoryStore() *memoryStore {
	return &memoryStore{
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

func (m *memoryStore) CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error) {
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

func (m *memoryStore) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	normEmail := strings.ToLower(strings.TrimSpace(email))
	a, ok := m.byEmail[normEmail]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (m *memoryStore) GetAccountByID(ctx context.Context, id int) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	a, ok := m.accounts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (m *memoryStore) ListSysadmins(ctx context.Context) ([]Sysadmin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]Sysadmin, 0, len(m.sysadmins))
	for _, sa := range m.sysadmins {
		list = append(list, *sa)
	}
	return list, nil
}

func (m *memoryStore) AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error) {
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

func (m *memoryStore) ListManagers(ctx context.Context, restaurantSlug string) ([]Manager, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]Manager, 0)
	for _, mgr := range m.managers {
		if mgr.RestaurantSlug == restaurantSlug {
			list = append(list, *mgr)
		}
	}
	return list, nil
}

func (m *memoryStore) AddManager(ctx context.Context, restaurantSlug string, accountID int) (*Manager, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.accounts[accountID]; !ok {
		return nil, ErrNotFound
	}

	for _, mgr := range m.managers {
		if mgr.RestaurantSlug == restaurantSlug && mgr.AccountID == accountID {
			return mgr, nil
		}
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

func (m *memoryStore) RemoveManager(ctx context.Context, restaurantSlug string, accountID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, mgr := range m.managers {
		if mgr.RestaurantSlug == restaurantSlug && mgr.AccountID == accountID {
			delete(m.managers, id)
			return nil
		}
	}
	return ErrNotFound
}

func (m *memoryStore) ListOwners(ctx context.Context, restaurantSlug string) ([]Owner, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]Owner, 0)
	for _, own := range m.owners {
		if own.RestaurantSlug == restaurantSlug {
			list = append(list, *own)
		}
	}
	return list, nil
}

func (m *memoryStore) AddOwner(ctx context.Context, restaurantSlug string, accountID int) (*Owner, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.accounts[accountID]; !ok {
		return nil, ErrNotFound
	}

	for _, own := range m.owners {
		if own.RestaurantSlug == restaurantSlug && own.AccountID == accountID {
			return own, nil
		}
	}

	own := &Owner{
		ID:             m.nextOwnID,
		RestaurantSlug: restaurantSlug,
		AccountID:      accountID,
	}
	m.nextOwnID++
	m.owners[own.ID] = own
	return own, nil
}

func (m *memoryStore) RemoveOwner(ctx context.Context, restaurantSlug string, accountID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var ownersForSlug []int
	var targetID int
	for id, own := range m.owners {
		if own.RestaurantSlug == restaurantSlug {
			ownersForSlug = append(ownersForSlug, id)
			if own.AccountID == accountID {
				targetID = id
			}
		}
	}

	if targetID == 0 {
		return ErrNotFound
	}
	if len(ownersForSlug) <= 1 {
		return ErrCannotRemoveLastOwner
	}

	delete(m.owners, targetID)
	return nil
}

func (m *memoryStore) GetAccountRoles(ctx context.Context, restaurantSlug string, accountID int) ([]string, error) {
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
		for _, own := range m.owners {
			if own.RestaurantSlug == restaurantSlug && own.AccountID == accountID {
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

func (m *memoryStore) GetRestaurantCNPJ(ctx context.Context, restaurantSlug string) (string, error) {
	return "12.345.678/0001-90", nil
}

func (m *memoryStore) CreatePhysicalRestaurant(ctx context.Context, restaurantSlug, name, cnpj, ownerCnpj string) (*PhysicalRestaurant, error) {
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

func (m *memoryStore) ListPhysicalRestaurants(ctx context.Context, restaurantSlug string) ([]PhysicalRestaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]PhysicalRestaurant, 0)
	for _, pr := range m.branches {
		if pr.RestaurantSlug == restaurantSlug {
			list = append(list, *pr)
		}
	}
	return list, nil
}

func (m *memoryStore) GetPhysicalRestaurant(ctx context.Context, branchID int) (*PhysicalRestaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pr, ok := m.branches[branchID]
	if !ok {
		return nil, ErrNotFound
	}
	return pr, nil
}

func (m *memoryStore) ListEmployees(ctx context.Context, branchID int) ([]Employee, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]Employee, 0)
	for _, emp := range m.employees {
		if emp.PhysicalRestaurantID == branchID {
			list = append(list, *emp)
		}
	}
	return list, nil
}

func (m *memoryStore) CreateEmployee(ctx context.Context, branchID int, name, cpf, position string) (*Employee, error) {
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

func (m *memoryStore) LinkEmployeeAccount(ctx context.Context, employeeID int, accountID int) (*Employee, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	emp, ok := m.employees[employeeID]
	if !ok {
		return nil, ErrNotFound
	}
	emp.AccountID = &accountID
	return emp, nil
}

func (m *memoryStore) VerifyAndOpenTerminalSession(ctx context.Context, branchID int, pin string, deviceID string) (*TerminalSession, error) {
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

	tokenStr := fmt.Sprintf("test_session_%d_%s", branchID, hex.EncodeToString([]byte(deviceID)))
	ts := &TerminalSession{
		ID:                   len(m.terminalSessions) + 1,
		PhysicalRestaurantID: branchID,
		Token:                tokenStr,
		DeviceID:             deviceID,
		ExpiresAt:            now.Add(8 * time.Hour),
		CreatedAt:            now,
	}
	m.terminalSessions[tokenStr] = ts
	return ts, nil
}

func (m *memoryStore) GetTerminalSession(ctx context.Context, token string) (*TerminalSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ts, ok := m.terminalSessions[token]
	if !ok || ts.ExpiresAt.Before(time.Now()) {
		return nil, ErrNotFound
	}
	return ts, nil
}

func (m *memoryStore) RotateTerminalPIN(ctx context.Context, branchID int, newPIN string) error {
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

func (m *memoryStore) GetTerminalOrders(ctx context.Context, branchID int) ([]TerminalOrder, error) {
	return []TerminalOrder{}, nil
}
