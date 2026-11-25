package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"pratello/server/internal/auth"
	"pratello/server/internal/db"
	"pratello/server/internal/menu"
	"pratello/server/internal/order"
	"pratello/server/internal/restaurants"
	"pratello/server/internal/review"
	"pratello/server/internal/sales"
	"pratello/server/internal/transaction"
	"pratello/server/internal/user"
)

func isAuthFixedCode(val string) bool {
	v := strings.TrimSpace(val)
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}

func main() {
	reset := flag.Bool("reset-demo", false, "wipe the development database and media, then seed João Burger and Tanoshii Sushi")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	fixedAuth := isAuthFixedCode(os.Getenv("AUTH_FIXED_CODE"))
	if fixedAuth {
		os.Setenv("AUTH_FIXED_CODE", "1")
		if os.Getenv("REDIS_URL") == "" {
			os.Setenv("REDIS_URL", "redis://localhost:6379/0")
		}
	}
	shouldReset := *reset || fixedAuth

	dbURL := os.Getenv("DATABASE_URL")
	if shouldReset && (os.Getenv("AUTH_FIXED_CODE") != "1" || os.Getenv("REDIS_URL") == "") {
		log.Fatal("database reset requires AUTH_FIXED_CODE=1 and REDIS_URL")
	}
	if shouldReset && dbURL == "" {
		log.Fatal("DATABASE_URL is required to reset and seed the database")
	}
	var database *sql.DB
	if dbURL != "" {
		var err error
		database, err = db.Open(dbURL)
		if err != nil {
			if shouldReset {
				log.Fatalf("Cannot connect to development database: %v", err)
			}
			log.Printf("Warning: failed to connect to database (%v); only in-memory authentication is available", err)
			database = nil
		} else {
			defer database.Close()
			if !shouldReset {
				if err := db.Migrate(database); err != nil {
					log.Fatalf("Database migration failed: %v", err)
				}
				log.Println("Database migrated successfully")
			}
		}
	} else {
		log.Println("DATABASE_URL not set; only in-memory authentication is available")
	}
	if shouldReset {
		log.Println("Resetting the active PostgreSQL schema and /media for the demo (João Burger & Tanoshii Sushi)")
		if err := resetDemo(ctx, database, mediaDirectory()); err != nil {
			log.Fatalf("Development reset failed: %v", err)
		}
	}

	authSvc := auth.NewService(auth.NewStore(database))
	if os.Getenv("AUTH_FIXED_CODE") == "1" {
		closeVerification, err := authSvc.EnableFixedCodeVerification(os.Getenv("REDIS_URL"))
		if err != nil {
			log.Fatalf("Fixed-code verification requires a valid REDIS_URL: %v", err)
		}
		defer closeVerification()
	}
	authHandler := auth.NewHandler(authSvc)
	restHandler := restaurants.NewHandler(database, authSvc)

	// Build mux and route tree
	mux := http.NewServeMux()
	user.NewHandler(database, authSvc.Identify).Routes(mux)
	menu.NewHandler(menu.NewService(database), authSvc.Identify).Routes(mux)
	order.NewHandler(order.NewService(database), authSvc.Identify, os.Getenv("ORDER_PUBLIC_BASE_URL")).Routes(mux)
	review.NewHandler(review.NewService(database), authSvc.Identify).Routes(mux)
	sales.NewHandler(sales.NewService(database), authSvc.Identify).Routes(mux)
	var loyaltyCache sales.CompletionCache
	if rawURL := os.Getenv("REDIS_URL"); rawURL != "" {
		cache, err := db.OpenRedis(rawURL)
		if err != nil {
			log.Print("Invalid REDIS_URL; using PostgreSQL loyalty idempotency")
		} else {
			defer cache.Close()
			loyaltyCache = cache
		}
	}
	transactionSvc := transaction.NewServiceWithLoyaltyCache(database, transaction.MockProvider{}, loyaltyCache)
	if os.Getenv("AUTH_FIXED_CODE") == "1" {
		transactionSvc.EnableMockOrderReady(database, 5*time.Second)
	}
	transaction.NewHandler(transactionSvc, authSvc.Identify).Routes(mux)
	handler := registerRoutes(mux, authHandler, restHandler)
	if shouldReset {
		fixture, err := seedDemo(ctx, database, handler)
		if err != nil {
			log.Fatalf("Demo seed failed: %v", err)
		}
		log.Printf("Demo data ready (João Burger & Tanoshii Sushi). Owner: %s; Administrator: %s; shared demo password: %s", fixture.Owner.Email, fixture.Administrator.Email, fixture.Password)
	}

	// Create and launch server
	srv := NewServer(addr, handler)

	go func() {
		log.Printf("Pratello API server running at http://localhost%s", addr)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Block on termination signal
	<-ctx.Done()
	log.Println("Shutting down Pratello API server gracefully...")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("Server gracefully stopped")
	}
}
