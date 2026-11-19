package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func tokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *Service) IssueTokens(ctx context.Context, id int) (*TokenPair, error) {
	return s.issueTokens(ctx, id, nil)
}

func (s *Service) issueTokens(ctx context.Context, id int, expectedVersion *int) (*TokenPair, error) {
	store, ok := s.store.(sessionStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	a, err := s.store.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if expectedVersion != nil && a.TokenVersion != *expectedVersion {
		return nil, ErrUnauthorized
	}
	access, err := s.signAccess(a)
	if err != nil {
		return nil, err
	}
	var nonce [32]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	now := time.Now()
	expiry := now.Add(30 * 24 * time.Hour)
	claims := AccountClaims{AccountID: a.ID, TokenVersion: a.TokenVersion, Type: "refresh", RegisteredClaims: jwt.RegisteredClaims{ID: hex.EncodeToString(nonce[:]), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(expiry)}}
	refresh, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}
	if err = store.SaveRefresh(ctx, id, a.TokenVersion, tokenHash(refresh), expiry); err != nil {
		return nil, err
	}
	return &TokenPair{access, refresh, "Bearer", 900}, nil
}

func (s *Service) Refresh(ctx context.Context, raw string) (*TokenPair, error) {
	store, ok := s.store.(sessionStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	token, err := jwt.ParseWithClaims(raw, &AccountClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrUnauthorized
		}
		return s.jwtSecret, nil
	}, jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}
	claims, ok := token.Claims.(*AccountClaims)
	if !ok || claims.Type != "refresh" || claims.AccountID <= 0 {
		return nil, ErrUnauthorized
	}
	id, err := store.ConsumeRefresh(ctx, tokenHash(raw), claims.TokenVersion)
	if err != nil {
		return nil, err
	}
	if id != claims.AccountID {
		return nil, ErrUnauthorized
	}
	return s.issueTokens(ctx, id, &claims.TokenVersion)
}

func (s *Service) Authenticate(ctx context.Context, token string) (*AccountClaims, error) {
	claims, err := s.ParseToken(token)
	if err != nil {
		return nil, ErrUnauthorized
	}
	a, err := s.store.GetAccountByID(ctx, claims.AccountID)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrUnauthorized
		}
		return nil, err
	}
	if a.TokenVersion != claims.TokenVersion || (a.Email == "" && a.Phone == "" && a.Password == "") {
		return nil, ErrUnauthorized
	}
	return claims, nil
}

func (s *Service) Logout(ctx context.Context, id int, raw string) error {
	store, ok := s.store.(sessionStore)
	if !ok {
		return ErrUnauthorized
	}
	return store.RevokeRefresh(ctx, tokenHash(raw), id)
}
func (s *Service) LogoutAll(ctx context.Context, id int) error {
	store, ok := s.store.(sessionStore)
	if !ok {
		return ErrUnauthorized
	}
	return store.RevokeAll(ctx, id)
}
