package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/auth"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/delivery"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/handlers"
	"github.com/Low-Stack-Technologies/message-delivery-service/pkg/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 1. Load Config
	cfgPath := "config.yaml"
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Start Hot-Reload
	if err := config.Watch(cfgPath); err != nil {
		log.Printf("Warning: Failed to start config watcher: %v", err)
	}

	// 3. Initialize Backends
	emailProvider := delivery.NewEmailProvider(cfg)
	smsProvider := delivery.NewSmsProvider(cfg)
	h := handlers.NewHandler(emailProvider, smsProvider)

	// 4. Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Get("/health", h.GetHealth)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.NewMiddleware())
		api.HandlerWithOptions(h, api.ChiServerOptions{
			BaseRouter: r,
		})
	})

	// 5. Start Server
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
