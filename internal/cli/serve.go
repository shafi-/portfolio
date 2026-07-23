package cli

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"project-dash/internal/api"
	"project-dash/internal/config"
	"project-dash/internal/database"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	Long: `Start the Portfolio HTTP API server for the dashboard.

The server provides RESTful endpoints for projects, search, configuration,
statistics, and relationships. Useful for the dashboard frontend.`,
	Run: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "HTTP server port")
}

func runServe(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()

	loader := config.NewLoader(cfgFile)
	cfg, err := loader.Load()
	if err != nil {
		logger.Error("failed to load config", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		logger.Error("failed to create database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("failed to connect to database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		logger.Error("failed to initialize database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	srv := api.NewServer(db.DB(), logger)
	addr := fmt.Sprintf(":%d", servePort)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv.Handler(),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("HTTP API server starting", models.Field{Key: "addr", Value: addr})
		fmt.Printf("Portfolio API server listening on %s\n", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", models.Field{Key: "error", Value: err})
			os.Exit(1)
		}
	}()

	<-quit
	logger.Info("shutting down server")
}
