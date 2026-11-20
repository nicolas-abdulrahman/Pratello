package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	dbURL := os.Getenv("DATABASE_URL")
	var database *sql.DB
	if dbURL != "" {
		var err error
		database, err = db.Open(dbURL)
		if err != nil {
			log.Printf("Warning: failed to connect to database (%v); only in-memory authentication is available", err)
			database = nil
		} else {
			defer database.Close()
			if err := db.Migrate(database); err != nil {
				log.Fatalf("Database migration failed: %v", err)
			} else {
				log.Println("Database migrated successfully")
			}
		}
	} else {
		log.Println("DATABASE_URL not set; only in-memory authentication is available")
	}

	authSvc := auth.NewService(auth.NewStore(database))
	authHandler := auth.NewHandler(authSvc)
	restHandler := restaurants.NewHandler(database, authSvc)

	// Build mux and route tree
	mux := http.NewServeMux()
	user.NewHandler(database, authSvc.Identify).Routes(mux)
	menu.NewHandler(menu.NewService(database), authSvc.Identify).Routes(mux)
	order.NewHandler(order.NewService(database), authSvc.Identify, os.Getenv("ORDER_PUBLIC_BASE_URL")).Routes(mux)
	review.NewHandler(review.NewService(database), authSvc.Identify).Routes(mux)
	sales.NewHandler(sales.NewService(database), authSvc.Identify).Routes(mux)
	transaction.NewHandler(transaction.NewService(database, transaction.MockProvider{}), authSvc.Identify).Routes(mux)
	handler := registerRoutes(mux, authHandler, restHandler)

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
