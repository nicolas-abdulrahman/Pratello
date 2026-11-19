package auth

import "context"

func validCPF(cpf string) bool {
	if len(cpf) != 11 {
		return false
	}
	for _, c := range cpf {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func (s *Service) ClaimCPF(ctx context.Context, id int, cpf string) (*Account, error) {
	if !validCPF(cpf) {
		return nil, ErrValidation
	}
	store, ok := s.store.(cpfStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	if err := store.ClaimCPF(ctx, id, cpf); err != nil {
		return nil, err
	}
	return s.store.GetAccountByID(ctx, id)
}

func (s *Service) RemoveCPF(ctx context.Context, id int) error {
	store, ok := s.store.(cpfStore)
	if !ok {
		return ErrUnauthorized
	}
	return store.RemoveCPF(ctx, id)
}
