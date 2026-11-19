package auth

import (
	"context"
	"errors"

	"github.com/lib/pq"
)

func (s *sqlStore) SetVerifiedEmail(ctx context.Context, id int, email, hash, oldHash string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE account SET email=$2, password_hash=$3, email_verified=TRUE WHERE id=$1 AND COALESCE(password_hash,'')=$4`, id, email, hash, oldHash)
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
	res, err := s.db.ExecContext(ctx, `UPDATE account SET email=NULL, password_hash=NULL, email_verified=FALSE WHERE id=$1 AND (phone_verified OR email IS NULL)`, id)
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
