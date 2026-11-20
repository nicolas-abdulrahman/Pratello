package auth

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"strings"
	"sync"
	"time"
)

type Account struct {
	ID            int
	Name          string
	Email         string
	EmailVerified bool
	Password      string
	Phone         string
	PhoneVerified bool
	CPF           string
	CPFClaimedAt  *time.Time
	TokenVersion  int
	CreatedAt     time.Time
}

type Sysadmin struct {
	ID        int
	AccountID int
	CreatedAt time.Time
}

type Store interface {
	CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountByID(ctx context.Context, id int) (*Account, error)

	ListSysadmins(ctx context.Context) ([]Sysadmin, error)
	AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error)
	IsSysadmin(context.Context, int) (bool, error)
}

func NewStore(db *sql.DB) Store {
	if db == nil {
		return NewMemoryStore()
	}
	return NewSQLStore(db)
}

func NewSQLStore(db *sql.DB) *sqlStore {
	return &sqlStore{db: db}
}

type sqlStore struct {
	db *sql.DB
}

func (s *sqlStore) IsSysadmin(ctx context.Context, id int) (bool, error) {
	var allowed bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM account WHERE id=$1 AND is_sysadmin)`, id).Scan(&allowed)
	return allowed, err
}

func (m *memoryStore) IsSysadmin(_ context.Context, id int) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, admin := range m.sysadmins {
		if admin.AccountID == id {
			return true, nil
		}
	}
	return false, nil
}

func (s *sqlStore) CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error) {
	var a Account
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO "account" (name, email, email_verified, password, phone, created_at)
		 VALUES (NULLIF($1, ''), $2, TRUE, $3, NULLIF($4, ''), NOW())
		 RETURNING id, COALESCE(name, ''), COALESCE(email, ''), COALESCE(password, ''), COALESCE(phone, ''), created_at`,
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
		`SELECT id, COALESCE(name, ''), COALESCE(email, ''), COALESCE(password, ''), COALESCE(phone, ''), created_at, token_version, email_verified, phone_verified, COALESCE(cpf, ''), cpf_claimed_at
		 FROM "account"
		 WHERE LOWER(email) = $1`,
		strings.ToLower(strings.TrimSpace(email)),
	).Scan(&a.ID, &a.Name, &a.Email, &a.Password, &a.Phone, &a.CreatedAt, &a.TokenVersion, &a.EmailVerified, &a.PhoneVerified, &a.CPF, &a.CPFClaimedAt)
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
		`SELECT id, COALESCE(name, ''), COALESCE(email, ''), COALESCE(password, ''), COALESCE(phone, ''), created_at, token_version, email_verified, phone_verified, COALESCE(cpf, ''), cpf_claimed_at
		 FROM "account"
		 WHERE id = $1`,
		id,
	).Scan(&a.ID, &a.Name, &a.Email, &a.Password, &a.Phone, &a.CreatedAt, &a.TokenVersion, &a.EmailVerified, &a.PhoneVerified, &a.CPF, &a.CPFClaimedAt)
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
		`SELECT id, id, created_at FROM account WHERE is_sysadmin ORDER BY id`,
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
		`UPDATE account SET is_sysadmin=true WHERE id=$1 RETURNING id, id, created_at`,
		accountID,
	).Scan(&sa.ID, &sa.AccountID, &sa.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

type memoryStore struct {
	mu            sync.RWMutex
	accounts      map[int]*Account
	byEmail       map[string]*Account
	nextAcctID    int
	sysadmins     map[int]*Sysadmin
	nextSysID     int
	refreshTokens map[string]refreshRecord
}

func NewMemoryStore() *memoryStore {
	return &memoryStore{
		accounts:      make(map[int]*Account),
		byEmail:       make(map[string]*Account),
		nextAcctID:    1,
		sysadmins:     make(map[int]*Sysadmin),
		nextSysID:     1,
		refreshTokens: make(map[string]refreshRecord),
	}
}

func (m *memoryStore) CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	normEmail := strings.ToLower(strings.TrimSpace(email))
	if _, exists := m.byEmail[normEmail]; exists {
		return nil, ErrConflict
	}
	if phone != "" {
		for _, existing := range m.accounts {
			if existing.Phone == phone {
				return nil, ErrConflict
			}
		}
	}

	a := &Account{
		ID:            m.nextAcctID,
		Name:          name,
		Email:         normEmail,
		EmailVerified: true,
		Password:      passwordHash,
		Phone:         phone,
		CreatedAt:     time.Now(),
	}
	m.nextAcctID++
	m.accounts[a.ID] = a
	m.byEmail[normEmail] = a
	copy := *a
	return &copy, nil
}

func (m *memoryStore) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	normEmail := strings.ToLower(strings.TrimSpace(email))
	a, ok := m.byEmail[normEmail]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *a
	return &copy, nil
}

func (m *memoryStore) GetAccountByID(ctx context.Context, id int) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	a, ok := m.accounts[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *a
	return &copy, nil
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

type cpfStore interface {
	RegisterWithCPF(context.Context, string, string, string, string, string, bool) (*Account, error)
	ClaimCPF(context.Context, int, string) error
	RemoveCPF(context.Context, int) error
}

func cpfDBError(err error) error {
	var e *pq.Error
	if errors.As(err, &e) && e.Code == "23505" {
		return ErrConflict
	}
	return err
}

func lockCPF(ctx context.Context, tx *sql.Tx, cpf string) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('account-cpf:' || $1, 0))`, cpf)
	return err
}

// RegisterWithCPF performs shadow graduation and uniqueness checks atomically.
func (s *sqlStore) RegisterWithCPF(ctx context.Context, name, email, hash, phone, cpf string, phoneOnly bool) (*Account, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = lockCPF(ctx, tx, cpf); err != nil {
		return nil, err
	}
	var id int
	var shadow bool
	err = tx.QueryRowContext(ctx, `SELECT id, email IS NULL AND phone IS NULL AND password IS NULL AND NOT is_sysadmin FROM account WHERE cpf=$1 FOR UPDATE`, cpf).Scan(&id, &shadow)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		if !shadow {
			return nil, ErrConflict
		}
		_, err = tx.ExecContext(ctx, `UPDATE account SET name=NULLIF($2,''), email=NULLIF($3,''), password=NULLIF($4,''), phone=NULLIF($5,''), email_verified=$6, phone_verified=$7, cpf_claimed_at=NOW() WHERE id=$1`, id, name, email, hash, phone, !phoneOnly, phoneOnly)
	} else {
		err = tx.QueryRowContext(ctx, `INSERT INTO account (name,email,password,phone,email_verified,phone_verified,cpf,cpf_claimed_at) VALUES (NULLIF($1,''),NULLIF($2,''),NULLIF($3,''),NULLIF($4,''),$5,$6,$7,NOW()) RETURNING id`, name, email, hash, phone, !phoneOnly, phoneOnly, cpf).Scan(&id)
	}
	if err != nil {
		return nil, cpfDBError(err)
	}
	if err = tx.Commit(); err != nil {
		return nil, cpfDBError(err)
	}
	return s.GetAccountByID(ctx, id)
}

func (s *sqlStore) ClaimCPF(ctx context.Context, id int, cpf string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = lockCPF(ctx, tx, cpf); err != nil {
		return err
	}
	var current string
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(cpf,'') FROM account WHERE id=$1 FOR UPDATE`, id).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if current != "" {
		return ErrConflict
	}
	var shadowID int
	var shadow bool
	err = tx.QueryRowContext(ctx, `SELECT id, email IS NULL AND phone IS NULL AND password IS NULL AND NOT is_sysadmin FROM account WHERE cpf=$1 FOR UPDATE`, cpf).Scan(&shadowID, &shadow)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil {
		if !shadow {
			return ErrConflict
		}
		if err = transferWallets(ctx, tx, id, shadowID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM account WHERE id=$1`, shadowID); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE account SET cpf=$2, cpf_claimed_at=NOW() WHERE id=$1`, id, cpf); err != nil {
		return cpfDBError(err)
	}
	return cpfDBError(tx.Commit())
}

func transferWallets(ctx context.Context, tx *sql.Tx, target, shadow int) error {
	// Lock existing wallet rows before moving balances. Account locks held by
	// the caller also serialize new wallet creation against the claim.
	rows, err := tx.QueryContext(ctx, `SELECT id FROM loyalty_wallet WHERE account_id IN ($1,$2) ORDER BY id FOR UPDATE`, target, shadow)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	queries := []string{
		`UPDATE loyalty_wallet d SET coins_amount=d.coins_amount+s.coins_amount FROM loyalty_wallet s WHERE d.account_id=$1 AND s.account_id=$2 AND d.restaurant_id=s.restaurant_id`,
		// Keep every ledger entry and its original monetary values and timestamp.
		`UPDATE wallet_transaction l SET loyalty_wallet_id=d.id FROM loyalty_wallet s JOIN loyalty_wallet d ON d.restaurant_id=s.restaurant_id WHERE d.account_id=$1 AND s.account_id=$2 AND l.loyalty_wallet_id=s.id`,
		`DELETE FROM loyalty_wallet s USING loyalty_wallet d WHERE d.account_id=$1 AND s.account_id=$2 AND d.restaurant_id=s.restaurant_id`,
		`UPDATE loyalty_wallet SET account_id=$1 WHERE account_id=$2`,
	}
	for _, q := range queries {
		if _, err = tx.ExecContext(ctx, q, target, shadow); err != nil {
			return err
		}
	}
	return nil
}

func (s *sqlStore) RemoveCPF(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, `UPDATE account SET cpf=NULL, cpf_claimed_at=NULL WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *memoryStore) RegisterWithCPF(_ context.Context, name, email, hash, phone, cpf string, phoneOnly bool) (*Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	email = strings.ToLower(strings.TrimSpace(email))
	if email != "" && m.byEmail[email] != nil {
		return nil, ErrConflict
	}
	var a *Account
	for _, existing := range m.accounts {
		if phone != "" && existing.Phone == phone {
			return nil, ErrConflict
		}
		if existing.CPF == cpf {
			a = existing
		}
	}
	if a != nil && (a.Email != "" || a.Phone != "" || a.Password != "") {
		return nil, ErrConflict
	}
	if a == nil {
		a = &Account{ID: m.nextAcctID, CPF: cpf, CreatedAt: time.Now()}
		m.nextAcctID++
		m.accounts[a.ID] = a
	}
	now := time.Now()
	a.Name = name
	a.Email = email
	a.Password = hash
	a.Phone = phone
	a.EmailVerified = !phoneOnly
	a.PhoneVerified = phoneOnly
	a.CPFClaimedAt = &now
	if email != "" {
		m.byEmail[email] = a
	}
	copy := *a
	return &copy, nil
}

func (m *memoryStore) ClaimCPF(_ context.Context, id int, cpf string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.accounts[id]
	if a == nil {
		return ErrNotFound
	}
	if a.CPF != "" {
		return ErrConflict
	}
	for _, other := range m.accounts {
		if other.CPF == cpf {
			if other.Email != "" || other.Phone != "" || other.Password != "" {
				return ErrConflict
			}
			delete(m.accounts, other.ID)
			break
		}
	}
	now := time.Now()
	a.CPF = cpf
	a.CPFClaimedAt = &now
	return nil
}

func (m *memoryStore) RemoveCPF(_ context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.accounts[id]
	if a == nil {
		return ErrNotFound
	}
	a.CPF = ""
	a.CPFClaimedAt = nil
	return nil
}

func (s *sqlStore) SetVerifiedEmail(ctx context.Context, id int, email, hash, oldHash string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE account SET email=$2, password=$3, email_verified=TRUE WHERE id=$1 AND COALESCE(password,'')=$4`, id, email, hash, oldHash)
	var e *pq.Error
	if errors.As(err, &e) && e.Code == "23505" {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func (s *sqlStore) RemoveEmail(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, `UPDATE account SET email=NULL, password=NULL, email_verified=FALSE WHERE id=$1 AND (phone_verified OR email IS NULL)`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func (m *memoryStore) SetVerifiedEmail(_ context.Context, id int, email, hash, oldHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.accounts[id]
	if a == nil {
		return ErrNotFound
	}
	if a.Password != oldHash {
		return ErrConflict
	}
	if other := m.byEmail[email]; other != nil && other.ID != id {
		return ErrConflict
	}
	if a.Email != "" {
		delete(m.byEmail, a.Email)
	}
	a.Email = email
	a.Password = hash
	a.EmailVerified = true
	m.byEmail[email] = a
	return nil
}

func (m *memoryStore) RemoveEmail(_ context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.accounts[id]
	if a == nil {
		return ErrNotFound
	}
	if a.Email != "" && !a.PhoneVerified {
		return ErrConflict
	}
	delete(m.byEmail, a.Email)
	a.Email = ""
	a.Password = ""
	a.EmailVerified = false
	return nil
}

type phoneStore interface {
	GetAccountByPhone(context.Context, string) (*Account, error)
	CreatePhoneAccount(context.Context, string, string) (*Account, error)
	SetVerifiedPhone(context.Context, int, string) error
	RemovePhone(context.Context, int) error
}

func (s *sqlStore) CreatePhoneAccount(ctx context.Context, phone, name string) (*Account, error) {
	var id int
	err := s.db.QueryRowContext(ctx, `INSERT INTO account (phone, phone_verified, name) VALUES ($1, TRUE, NULLIF($2, '')) RETURNING id`, phone, name).Scan(&id)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return nil, ErrConflict
	}
	if err != nil {
		return nil, err
	}
	return s.GetAccountByID(ctx, id)
}

func (m *memoryStore) CreatePhoneAccount(_ context.Context, phone, name string) (*Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.accounts {
		if a.Phone == phone {
			return nil, ErrConflict
		}
	}
	a := &Account{ID: m.nextAcctID, Name: name, Phone: phone, PhoneVerified: true, CreatedAt: time.Now()}
	m.nextAcctID++
	m.accounts[a.ID] = a
	copy := *a
	return &copy, nil
}

func (s *sqlStore) GetAccountByPhone(ctx context.Context, phone string) (*Account, error) {
	var id int
	err := s.db.QueryRowContext(ctx, `SELECT id FROM account WHERE phone=$1`, phone).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.GetAccountByID(ctx, id)
}

func (s *sqlStore) SetVerifiedPhone(ctx context.Context, id int, phone string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE account SET phone=$2, phone_verified=TRUE WHERE id=$1`, id, phone)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *sqlStore) RemovePhone(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, `UPDATE account SET phone=NULL, phone_verified=FALSE WHERE id=$1 AND (NOT phone_verified OR email_verified)`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func (m *memoryStore) GetAccountByPhone(_ context.Context, phone string) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.accounts {
		if a.Phone == phone && phone != "" {
			copy := *a
			return &copy, nil
		}
	}
	return nil, ErrNotFound
}

func (m *memoryStore) SetVerifiedPhone(_ context.Context, id int, phone string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.accounts[id]
	if a == nil {
		return ErrNotFound
	}
	for _, other := range m.accounts {
		if other.ID != id && other.Phone == phone {
			return ErrConflict
		}
	}
	a.Phone = phone
	a.PhoneVerified = true
	return nil
}

func (m *memoryStore) RemovePhone(_ context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.accounts[id]
	if a == nil {
		return ErrNotFound
	}
	if a.PhoneVerified && !a.EmailVerified {
		return ErrConflict
	}
	a.Phone = ""
	a.PhoneVerified = false
	return nil
}

type refreshRecord struct {
	accountID int
	expiresAt time.Time
}

type sessionStore interface {
	SaveRefresh(context.Context, int, int, string, time.Time) error
	ConsumeRefresh(context.Context, string, int) (int, error)
	RevokeRefresh(context.Context, string, int) error
	RevokeAll(context.Context, int) error
}

func (s *sqlStore) SaveRefresh(ctx context.Context, accountID, version int, hash string, expires time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var current int
	err = tx.QueryRowContext(ctx, `SELECT token_version FROM account WHERE id=$1 FOR UPDATE`, accountID).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUnauthorized
	}
	if err != nil {
		return err
	}
	if current != version {
		return ErrUnauthorized
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO refresh_token (account_id, token_hash, expires_at) VALUES ($1,$2,$3)`, accountID, hash, expires); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *sqlStore) ConsumeRefresh(ctx context.Context, hash string, version int) (int, error) {
	var id int
	err := s.db.QueryRowContext(ctx, `UPDATE refresh_token r SET revoked_at=NOW() FROM account a
		WHERE r.token_hash=$1 AND r.account_id=a.id AND a.token_version=$2
		AND r.revoked_at IS NULL AND r.expires_at>NOW() RETURNING r.account_id`, hash, version).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrUnauthorized
	}
	return id, err
}

func (s *sqlStore) RevokeRefresh(ctx context.Context, hash string, accountID int) error {
	res, err := s.db.ExecContext(ctx, `UPDATE refresh_token SET revoked_at=NOW() WHERE token_hash=$1 AND account_id=$2 AND revoked_at IS NULL AND expires_at>NOW()`, hash, accountID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrUnauthorized
	}
	return nil
}

func (s *sqlStore) RevokeAll(ctx context.Context, accountID int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `UPDATE account SET token_version=token_version+1 WHERE id=$1`, accountID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrUnauthorized
	}
	if _, err = tx.ExecContext(ctx, `UPDATE refresh_token SET revoked_at=NOW() WHERE account_id=$1 AND revoked_at IS NULL`, accountID); err != nil {
		return err
	}
	return tx.Commit()
}

func (m *memoryStore) SaveRefresh(_ context.Context, id, version int, hash string, expires time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.accounts[id]
	if a == nil || a.TokenVersion != version {
		return ErrUnauthorized
	}
	m.refreshTokens[hash] = refreshRecord{id, expires}
	return nil
}

func (m *memoryStore) ConsumeRefresh(_ context.Context, hash string, version int) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.refreshTokens[hash]
	if !ok || !r.expiresAt.After(time.Now()) {
		return 0, ErrUnauthorized
	}
	a := m.accounts[r.accountID]
	if a == nil || a.TokenVersion != version {
		return 0, ErrUnauthorized
	}
	delete(m.refreshTokens, hash)
	return r.accountID, nil
}

func (m *memoryStore) RevokeRefresh(_ context.Context, hash string, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.refreshTokens[hash]
	if !ok || r.accountID != id || !r.expiresAt.After(time.Now()) {
		return ErrUnauthorized
	}
	delete(m.refreshTokens, hash)
	return nil
}

func (m *memoryStore) RevokeAll(_ context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.accounts[id]
	if a == nil {
		return ErrUnauthorized
	}
	a.TokenVersion++
	for hash, r := range m.refreshTokens {
		if r.accountID == id {
			delete(m.refreshTokens, hash)
		}
	}
	return nil
}
