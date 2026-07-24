package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	epic6api "project-dash/internal/api"
	"project-dash/internal/dashboard/api"
	"project-dash/internal/dashboard/assets"
	"project-dash/internal/dashboard/middleware"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
	"syscall"
	"time"
)

// Server represents the dashboard HTTP server
type Server struct {
	httpServer *http.Server
	epic6API   *epic6api.Server
	config     *models.Config
	store      *sql.DB
	logger     *logging.Logger
	startTime  time.Time
}

// NewServer creates a new dashboard server
func NewServer(db *sql.DB, config *models.Config, logger *logging.Logger) *Server {
	// Create Epic 6 API server
	epic6API := epic6api.NewServer(db, logger)

	return &Server{
		epic6API:  epic6API,
		config:    config,
		store:     db,
		logger:    logger,
		startTime: time.Now(),
	}
}

// Start starts the dashboard server
func (s *Server) Start() error {
	// Create HTTP multiplexer
	mux := http.NewServeMux()

	// Register Epic 6 API handlers (wrapped)
	s.registerAPIHandlers(mux)

	// Register dashboard-specific handlers
	s.registerDashboardHandlers(mux)

	// Register asset handler
	s.registerAssetHandlers(mux)

	// Create middleware chain
	handler := s.createMiddlewareChain(mux)

	// Configure HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Dashboard.Host, s.config.Dashboard.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("dashboard server starting",
			models.Field{Key: "address", Value: addr},
			models.Field{Key: "asset_path", Value: s.config.Dashboard.AssetPath})
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for startup error or successful bind
	select {
	case err := <-errChan:
		return fmt.Errorf("failed to start dashboard server: %w", err)
	case <-time.After(100 * time.Millisecond):
		// Successfully started
		s.logger.Info("dashboard server started",
			models.Field{Key: "address", Value: addr})
		return nil
	}
}

// registerAPIHandlers registers Epic 6 API handlers
func (s *Server) registerAPIHandlers(mux *http.ServeMux) {
	// Wrap Epic 6 handlers with dashboard error handling
	epic6Handler := s.epic6API.Handler()

	// Register the Epic 6 handler directly since it already has routing
	mux.Handle("/projects", epic6Handler)
	mux.Handle("/projects/", epic6Handler)
	mux.Handle("/search", epic6Handler)
	mux.Handle("/relationships/", epic6Handler)
	mux.Handle("/statistics", epic6Handler)
}

// registerDashboardHandlers registers dashboard-specific handlers
func (s *Server) registerDashboardHandlers(mux *http.ServeMux) {
	// Health handler (dashboard-specific)
	healthHandler := api.NewHealthHandler(s.store, &apiLoggerAdapter{s.logger})
	mux.Handle("/health", healthHandler)

	// Configuration handler (dashboard-specific)
	configHandler := api.NewConfigHandler(s.config, nil, &apiLoggerAdapter{s.logger})
	mux.Handle("/configuration", configHandler)
}

// registerAssetHandlers registers asset handlers
func (s *Server) registerAssetHandlers(mux *http.ServeMux) {
	// Create asset handler
	assetHandler := assets.NewHandler(s.config.Dashboard.AssetPath, &assetsLoggerAdapter{s.logger})

	// Register asset handler for all paths
	mux.Handle("/", assetHandler)
}

// createMiddlewareChain creates the middleware chain
func (s *Server) createMiddlewareChain(handler http.Handler) http.Handler {
	// Build CORS configuration
	corsConfig := middleware.DefaultCORSConfig()
	if len(s.config.Dashboard.AllowedOrigins) > 0 {
		corsConfig.AllowedOrigins = s.config.Dashboard.AllowedOrigins
	}

	// Apply middleware in reverse order (last one is outermost)
	chain := handler

	// CORS middleware
	chain = middleware.CORSMiddleware(corsConfig)(chain)

	// Body limit middleware (1MB)
	chain = middleware.BodyLimitMiddleware(1 << 20)(chain)

	// Request logging middleware
	chain = middleware.RequestLoggingMiddleware(&middlewareLoggerAdapter{s.logger})(chain)

	return chain
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	s.logger.Info("dashboard server shutting down")

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown dashboard server: %w", err)
	}

	s.logger.Info("dashboard server stopped")
	return nil
}

// WaitForShutdown waits for shutdown signal and gracefully shuts down
func (s *Server) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	s.logger.Info("shutdown signal received")

	// Create shutdown context with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		s.logger.Error("shutdown failed", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	os.Exit(0)
}

// API logger adapter for dashboard API handlers
type apiLoggerAdapter struct {
	logger *logging.Logger
}

func (z *apiLoggerAdapter) Info(msg string, fields ...api.Field) {
	logFields := make([]models.Field, len(fields))
	for i, f := range fields {
		logFields[i] = models.Field{Key: f.Key, Value: f.Value}
	}
	z.logger.Info(msg, logFields...)
}

func (z *apiLoggerAdapter) Error(msg string, fields ...api.Field) {
	logFields := make([]models.Field, len(fields))
	for i, f := range fields {
		logFields[i] = models.Field{Key: f.Key, Value: f.Value}
	}
	z.logger.Error(msg, logFields...)
}

func (z *apiLoggerAdapter) Warn(msg string, fields ...api.Field) {
	logFields := make([]models.Field, len(fields))
	for i, f := range fields {
		logFields[i] = models.Field{Key: f.Key, Value: f.Value}
	}
	z.logger.Warn(msg, logFields...)
}

// Assets logger adapter for asset handlers
type assetsLoggerAdapter struct {
	logger *logging.Logger
}

func (z *assetsLoggerAdapter) Info(msg string, fields ...assets.Field) {
	logFields := make([]models.Field, len(fields))
	for i, f := range fields {
		logFields[i] = models.Field{Key: f.Key, Value: f.Value}
	}
	z.logger.Info(msg, logFields...)
}

func (z *assetsLoggerAdapter) Error(msg string, fields ...assets.Field) {
	logFields := make([]models.Field, len(fields))
	for i, f := range fields {
		logFields[i] = models.Field{Key: f.Key, Value: f.Value}
	}
	z.logger.Error(msg, logFields...)
}

func (z *assetsLoggerAdapter) Warn(msg string, fields ...assets.Field) {
	logFields := make([]models.Field, len(fields))
	for i, f := range fields {
		logFields[i] = models.Field{Key: f.Key, Value: f.Value}
	}
	z.logger.Warn(msg, logFields...)
}

// Middleware logger adapter for middleware
type middlewareLoggerAdapter struct {
	logger *logging.Logger
}

func (z *middlewareLoggerAdapter) Info(msg string, fields ...middleware.Field) {
	logFields := make([]models.Field, len(fields))
	for i, f := range fields {
		logFields[i] = models.Field{Key: f.Key, Value: f.Value}
	}
	z.logger.Info(msg, logFields...)
}
