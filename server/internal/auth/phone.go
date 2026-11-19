package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

func validPhone(phone string) bool {
	if len(phone) < 8 || len(phone) > 20 {
		return false
	}
	for i, c := range phone {
		if c == '+' && i == 0 {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func (s *Service) notifyPhone(phone string) error {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return err
	}
	return s.notifier.SendOTP(phone, fmt.Sprintf("%06d", n.Int64()))
}

// RequestPhoneLogin always succeeds for unknown or unverified phone numbers.
func (s *Service) RequestPhoneLogin(ctx context.Context, phone string) error {
	phone = strings.TrimSpace(phone)
	if !validPhone(phone) {
		return ErrValidation
	}
	store, ok := s.store.(phoneStore)
	if !ok {
		return ErrUnauthorized
	}
	a, err := store.GetAccountByPhone(ctx, phone)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !a.PhoneVerified || a.Phone != phone {
		return nil
	}
	return s.notifyPhone(phone)
}

func (s *Service) RequestPhoneRegistration(ctx context.Context, phone string) error {
	phone = strings.TrimSpace(phone)
	if !validPhone(phone) {
		return ErrValidation
	}
	store, ok := s.store.(phoneStore)
	if !ok {
		return ErrUnauthorized
	}
	_, err := store.GetAccountByPhone(ctx, phone)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrNotFound) {
		return err
	}
	return s.notifyPhone(phone)
}

func (s *Service) RegisterPhone(ctx context.Context, phone, name, cpf string) (*TokenPair, error) {
	phone = strings.TrimSpace(phone)
	if !validPhone(phone) || (cpf != "" && !validCPF(cpf)) {
		return nil, ErrValidation
	}
	store, ok := s.store.(phoneStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	var a *Account
	var err error
	if cpf != "" {
		cpfStore, ok := s.store.(cpfStore)
		if !ok {
			return nil, ErrUnauthorized
		}
		a, err = cpfStore.RegisterWithCPF(ctx, strings.TrimSpace(name), "", "", phone, cpf, true)
	} else {
		a, err = store.CreatePhoneAccount(ctx, phone, strings.TrimSpace(name))
	}
	if err != nil {
		return nil, err
	}
	return s.IssueTokens(ctx, a.ID)
}

// LoginPhone completes simulated verification without requiring a user-entered PIN.
func (s *Service) LoginPhone(ctx context.Context, phone string) (*TokenPair, error) {
	phone = strings.TrimSpace(phone)
	if !validPhone(phone) {
		return nil, ErrInvalidCredentials
	}
	store, ok := s.store.(phoneStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	a, err := store.GetAccountByPhone(ctx, phone)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !a.PhoneVerified || a.Phone != phone {
		return nil, ErrInvalidCredentials
	}
	return s.IssueTokens(ctx, a.ID)
}

func (s *Service) AddPhone(ctx context.Context, id int, phone string) (*Account, error) {
	phone = strings.TrimSpace(phone)
	if !validPhone(phone) {
		return nil, ErrValidation
	}
	store, ok := s.store.(phoneStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	existing, err := store.GetAccountByPhone(ctx, phone)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if err == nil && existing.ID != id {
		return nil, ErrConflict
	}
	if err = s.notifyPhone(phone); err != nil {
		return nil, err
	}
	if err = store.SetVerifiedPhone(ctx, id, phone); err != nil {
		return nil, err
	}
	return s.store.GetAccountByID(ctx, id)
}

func (s *Service) RemovePhone(ctx context.Context, id int) error {
	store, ok := s.store.(phoneStore)
	if !ok {
		return ErrUnauthorized
	}
	return store.RemovePhone(ctx, id)
}
