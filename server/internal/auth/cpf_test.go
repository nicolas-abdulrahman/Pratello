package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestShadowGraduationAndClaim(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	svc := NewService(store)
	svc.SetHashCost(bcrypt.MinCost)
	store.accounts[1] = &Account{ID: 1, CPF: "12345678901", CreatedAt: time.Now()}
	store.nextAcctID = 2
	u, err := svc.Register(ctx, RegisterInput{Email: "a@example.com", Password: "password123", CPF: "12345678901"})
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != 1 {
		t.Fatalf("shadow not graduated: %d", u.ID)
	}
	a, err := store.GetAccountByID(ctx, 1)
	if err != nil || a.CPFClaimedAt == nil || !a.EmailVerified {
		t.Fatalf("graduation: %#v %v", a, err)
	}
	if _, err := svc.Register(ctx, RegisterInput{Email: "b@example.com", Password: "password123", CPF: "12345678901"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("real CPF reused: %v", err)
	}
	store.accounts[2] = &Account{ID: 2, CPF: "22222222222", CreatedAt: time.Now()}
	store.nextAcctID = 3
	b, err := svc.Register(ctx, RegisterInput{Email: "b@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := svc.ClaimCPF(ctx, b.ID, "22222222222")
	if err != nil || claimed.CPF != "22222222222" {
		t.Fatalf("claim: %#v %v", claimed, err)
	}
	if _, err := store.GetAccountByID(ctx, 2); !errors.Is(err, ErrNotFound) {
		t.Fatalf("shadow remains: %v", err)
	}
	if _, err := svc.ClaimCPF(ctx, b.ID, "12345678901"); !errors.Is(err, ErrConflict) {
		t.Fatalf("second claim: %v", err)
	}
	if err := svc.RemoveCPF(ctx, b.ID); err != nil {
		t.Fatal(err)
	}
	if a, _ := store.GetAccountByID(ctx, b.ID); a.CPF != "" {
		t.Fatal("CPF still attached")
	}
}

func TestConcurrentCPFRegistration(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	var wg sync.WaitGroup
	results := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := store.RegisterWithCPF(ctx, "", "", "", fmt.Sprintf("+551199999%04d", i), "12345678901", true)
			results <- err
		}(i)
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrConflict) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("registrations succeeded: %d", successes)
	}
}
