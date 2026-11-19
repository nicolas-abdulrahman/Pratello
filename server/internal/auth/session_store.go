package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

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
