package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/config"
	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/handlers"
	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/middleware"
	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/repository"
	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/service"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Server represents the HTTP server
type Server struct {
	cfg            *config.Config
	repo           repository.Repository
	accrualSvc     *service.AccrualService
	orderProcessor *service.OrderProcessor
	handler        *handlers.Handler
	httpServer     *http.Server
}

// NewServer creates a new server
func NewServer(cfg *config.Config) *Server {
	repo := repository.NewPostgresRepository(cfg.DatabaseURI)
	accrualSvc := service.NewAccrualService(cfg.AccrualSystemAddress)
	orderProcessor := service.NewOrderProcessor(repo, accrualSvc)
	handler := handlers.NewHandler(repo, accrualSvc, "your-secret-key") // In real app, use a secure random key

	return &Server{
		cfg:            cfg,
		repo:           repo,
		accrualSvc:     accrualSvc,
		orderProcessor: orderProcessor,
		handler:        handler,
	}
}

// Run starts the HTTP server
func (s *Server) Run() error {
	// Initialize repository
	if err := s.repo.InitDB(s.cfg.DatabaseURI); err != nil {
		return err
	}

	// Start order processor
	s.orderProcessor.Start()

	// Create router
	r := chi.NewRouter()

	// Basic middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// Public routes
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", s.handler.RegisterUser)
		r.Post("/login", s.handler.LoginUser)

		// Protected routes
		r.Group(func(r chi.Router) {
			jwtConfig := &middleware.JWTConfig{
				SecretKey: "your-secret-key", // In real app, use a secure random key
				Repo:      s.repo,
			}
			r.Use(middleware.AuthMiddleware(jwtConfig))

			r.Post("/orders", s.handler.UploadOrder)
			r.Get("/orders", s.handler.GetOrders)
			r.Get("/balance", s.handler.GetBalance)
			r.Post("/balance/withdraw", s.handler.WithdrawBalance)
			r.Get("/withdrawals", s.handler.GetWithdrawals)
		})
	})

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    s.cfg.RunAddress,
		Handler: r,
	}

	// Start server
	log.Printf("Starting server on %s", s.cfg.RunAddress)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	// Stop order processor
	if s.orderProcessor != nil {
		s.orderProcessor.Stop()
	}

	// Close repository
	if s.repo != nil {
		if err := s.repo.Close(); err != nil {
			return err
		}
	}

	return nil
}
