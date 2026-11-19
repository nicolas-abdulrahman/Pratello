package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type emailStore interface {
	SetVerifiedEmail(context.Context, int, string, string, string) error
	RemoveEmail(context.Context, int) error
}

func (s *Service) SetEmail(ctx context.Context, id int, email, password, currentPassword string) (*Account, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(password) < 8 {
		return nil, ErrValidation
	}
	a, err := s.store.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Email != "" && bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(currentPassword)) != nil {
		return nil, ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.hashCost)
	if err != nil {
		return nil, err
	}
	store, ok := s.store.(emailStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	if other, err := s.store.GetAccountByEmail(ctx, email); err == nil && other.ID != id {
		return nil, ErrConflict
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if err = s.notifier.SendEmail(email, ""); err != nil {
		return nil, err
	}
	if err = store.SetVerifiedEmail(ctx, id, email, string(hash), a.Password); err != nil {
		return nil, err
	}
	return s.store.GetAccountByID(ctx, id)
}

func (s *Service) RemoveEmail(ctx context.Context, id int) error {
	store, ok := s.store.(emailStore)
	if !ok {
		return ErrUnauthorized
	}
	return store.RemoveEmail(ctx, id)
}
