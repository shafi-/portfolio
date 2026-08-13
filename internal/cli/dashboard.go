package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"project-dash/internal/api"
	"project-dash/internal/config"
	"project-dash/internal/dashboard"
	"project-dash/internal/database"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

var (
	dashboardPort     int
	dashboardDevMode  bool
	dashboardDistPath string
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Start the dashboard web server",
	Long: `Start the Portfolio dashboard web server.

The dashboard provides a web interface for exploring your projects,
viewing statistics, searching, and visualizing relationships.

In development mode (--dev), assets are served from the local filesystem
for hot reload. In production mode, assets are embedded in the binary.`,
	Run: runDashboard,
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().IntVarP(&dashboardPort, "port", "p", 3000, "Dashboard server port")
	dashboardCmd.Flags().BoolVar(&dashboardDevMode, "dev", false, "Development mode (serve from filesystem)")
	dashboardCmd.Flags().StringVar(&dashboardDistPath, "dist", models.DefaultDashboardExternalPath, "Path to dist folder in dev mode")
}

func runDashboard(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()

	provider := config.NewProvider(cfgFile)
	cfg, err := provider.Load()
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

	// Determine serving mode
	mode := dashboard.ModeEmbedded
	if dashboardDevMode {
		mode = dashboard.ModeExternal
	}

	// Create asset server
	assetServer, err := dashboard.NewAssetServer(mode, dashboardDistPath)
	if err != nil {
		logger.Error("failed to create asset server", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	// Create API server
	apiServer := api.NewServer(db.DB(), logger)

	// Create dashboard router
	corsEnabled := true // Enable CORS for development
	router := dashboard.NewDashboardRouter(assetServer, apiServer.Handler(), corsEnabled)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", dashboardPort)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		modeStr := "development"
		if mode == dashboard.ModeEmbedded {
			modeStr = "production"
		}

		logger.Info("Dashboard server starting",
			models.Field{Key: "addr", Value: addr},
			models.Field{Key: "mode", Value: modeStr})

		fmt.Printf("Portfolio Dashboard listening on %s\n", addr)
		fmt.Printf("Open http://localhost%s in your browser\n", addr)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", models.Field{Key: "error", Value: err})
			os.Exit(1)
		}
	}()

	<-quit
	logger.Info("shutting down dashboard server")
	if err := httpServer.Shutdown(context.Background()); err != nil {
		logger.Error("shutdown error", models.Field{Key: "error", Value: err})
	}
}
