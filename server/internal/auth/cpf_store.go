package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

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
	err = tx.QueryRowContext(ctx, `SELECT id, email IS NULL AND phone IS NULL AND password_hash IS NULL FROM account WHERE cpf=$1 FOR UPDATE`, cpf).Scan(&id, &shadow)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		if !shadow {
			return nil, ErrConflict
		}
		_, err = tx.ExecContext(ctx, `UPDATE account SET name=NULLIF($2,''), email=NULLIF($3,''), password_hash=NULLIF($4,''), phone=NULLIF($5,''), email_verified=$6, phone_verified=$7, cpf_claimed_at=NOW() WHERE id=$1`, id, name, email, hash, phone, !phoneOnly, phoneOnly)
	} else {
		err = tx.QueryRowContext(ctx, `INSERT INTO account (name,email,password_hash,phone,email_verified,phone_verified,cpf,cpf_claimed_at) VALUES (NULLIF($1,''),NULLIF($2,''),NULLIF($3,''),NULLIF($4,''),$5,$6,$7,NOW()) RETURNING id`, name, email, hash, phone, !phoneOnly, phoneOnly, cpf).Scan(&id)
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
	err = tx.QueryRowContext(ctx, `SELECT id, email IS NULL AND phone IS NULL AND password_hash IS NULL FROM account WHERE cpf=$1 FOR UPDATE`, cpf).Scan(&shadowID, &shadow)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil {
		if !shadow {
			return ErrConflict
		}
		if _, err = tx.ExecContext(ctx, `UPDATE loyalty_customer SET account_id=$1 WHERE account_id=$2`, id, shadowID); err != nil {
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
