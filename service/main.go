package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/auth"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/delivery"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/handlers"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/state"
	"github.com/Low-Stack-Technologies/message-delivery-service/pkg/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	resetDB := flag.Bool("reset-db", false, "reset the SQLite database and exit")
	seedDB := flag.Bool("seed-db", false, "reset and seed the SQLite database, then exit")
	flag.Parse()

	// 1. Open Database
	dbPath := os.Getenv("MDS_DB_PATH")
	if dbPath == "" {
		dbPath = "message-delivery-service.db"
	}

	if *resetDB || *seedDB {
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			log.Fatalf("Failed to remove database file: %v", err)
		}
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			log.Fatalf("Failed to prepare database directory: %v", err)
		}
		store, err := state.New(dbPath)
		if err != nil {
			log.Fatalf("Failed to initialize state store: %v", err)
		}
		if err := store.Close(); err != nil {
			log.Printf("Warning: failed to close database: %v", err)
		}
		log.Printf("Database initialized at %s", dbPath)
		return
	}

	store, err := state.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize state store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("Warning: failed to close database: %v", err)
		}
	}()

	cfg := config.Get()
	if cfg == nil {
		log.Fatal("configuration snapshot is not available")
	}

	// 2. Initialize Backends
	emailProvider := delivery.NewEmailProvider(cfg)
	smsProvider := delivery.NewSmsProvider(cfg)
	h := handlers.NewHandler(emailProvider, smsProvider, store)

	// 3. Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Get("/health", h.GetHealth)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.NewMiddleware(store))
		api.HandlerWithOptions(h, api.ChiServerOptions{
			BaseRouter: r,
		})
	})

	// 4. Start Server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
