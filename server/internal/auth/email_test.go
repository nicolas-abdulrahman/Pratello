package auth

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestEmailChangeAndLastMethod(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	svc := NewService(store)
	svc.SetHashCost(bcrypt.MinCost)
	u, err := svc.Register(ctx, RegisterInput{Email: "a@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.RemoveEmail(ctx, u.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("removed last method: %v", err)
	}
	if _, err = svc.SetEmail(ctx, u.ID, "b@example.com", "newpassword", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong current password: %v", err)
	}
	if _, err = svc.SetEmail(ctx, u.ID, "b@example.com", "newpassword", "password123"); err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.Login(ctx, "b@example.com", "newpassword"); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.AddPhone(ctx, u.ID, "+5511999999999"); err != nil {
		t.Fatal(err)
	}
	if err = svc.RemoveEmail(ctx, u.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.Login(ctx, "b@example.com", "newpassword"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("removed email login: %v", err)
	}
	if err = svc.RemovePhone(ctx, u.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("removed last phone: %v", err)
	}
}
