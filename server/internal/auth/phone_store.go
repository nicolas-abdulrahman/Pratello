package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

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
