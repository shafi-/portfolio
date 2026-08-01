package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"project-dash/internal/config"
	"project-dash/internal/database"
	"project-dash/internal/discovery"
	"project-dash/internal/fs"
	"project-dash/internal/logging"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage and query discovered projects",
	Long:  `List, search, and get details about discovered projects in your portfolio.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all discovered projects",
	Run:   runListProjects,
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search projects by name",
	Args:  cobra.ExactArgs(1),
	Run:   runSearchProjects,
}

var getCmd = &cobra.Command{
	Use:   "get <project-id>",
	Short: "Get detailed information about a project",
	Args:  cobra.ExactArgs(1),
	Run:   runGetProject,
}

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Run project discovery on configured roots",
	Run:   runDiscoverProjects,
}

// discoverRootCmd is a top-level shortcut for "portfolio projects discover".
// It reuses the same handler so behavior stays identical.
var discoverRootCmd = &cobra.Command{
	Use:   "discover",
	Short: "Run project discovery on configured roots (shortcut for 'projects discover')",
	Long: `Run project discovery on configured root directories.

This is a shortcut for 'portfolio projects discover'.`,
	Run: runDiscoverProjects,
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(discoverRootCmd)
	projectsCmd.AddCommand(listCmd)
	projectsCmd.AddCommand(searchCmd)
	projectsCmd.AddCommand(getCmd)
	projectsCmd.AddCommand(discoverCmd)
}

func runListProjects(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()

	provider := config.NewProvider(cfgFile)
	cfg, err := provider.Load()
	if err != nil {
		logger.Error("Failed to load config", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		logger.Error("Failed to create database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("Failed to connect to database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer db.Close()

	projectsStore := store.NewProjectStore(db.DB(), logger.Zap())
	projects, err := projectsStore.ListProjects()
	if err != nil {
		logger.Error("Failed to list projects", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println("No projects discovered yet.")
		fmt.Println("Run 'portfolio projects discover' to discover projects.")
		return
	}

	// Display in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPATH\tTYPE")
	for _, p := range projects {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.Name, p.RootPath, p.RepositoryType)
	}
	w.Flush()

	fmt.Printf("\nTotal: %d projects\n", len(projects))
}

func runSearchProjects(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()
	query := args[0]

	provider := config.NewProvider(cfgFile)
	cfg, err := provider.Load()
	if err != nil {
		logger.Error("Failed to load config", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		logger.Error("Failed to create database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("Failed to connect to database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer db.Close()

	rows, err := db.DB().Query(
		"SELECT id, name, root_path, repository_type FROM projects WHERE name LIKE ? ORDER BY name LIMIT 50",
		"%"+query+"%",
	)
	if err != nil {
		logger.Error("Search failed", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer rows.Close()

	var results []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.RootPath, &p.RepositoryType); err != nil {
			logger.Warn("Failed to scan project row", models.Field{Key: "error", Value: err})
			continue
		}
		results = append(results, p)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error iterating project rows", models.Field{Key: "error", Value: err})
	}

	if len(results) == 0 {
		fmt.Printf("No projects found matching '%s'\n", query)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPATH\tTYPE")
	for _, p := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.Name, p.RootPath, p.RepositoryType)
	}
	w.Flush()

	fmt.Printf("\nFound: %d projects\n", len(results))
}

func runGetProject(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()
	projectID := args[0]

	provider := config.NewProvider(cfgFile)
	cfg, err := provider.Load()
	if err != nil {
		logger.Error("Failed to load config", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		logger.Error("Failed to create database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("Failed to connect to database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer db.Close()

	projectsStore := store.NewProjectStore(db.DB(), logger.Zap())
	project, err := projectsStore.GetProject(projectID)
	if err != nil {
		logger.Error("Failed to get project", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if project == nil {
		fmt.Printf("Project not found: %s\n", projectID)
		os.Exit(1)
	}

	metadataStore := store.NewMetadataStore(db.DB(), logger.Zap())
	metadata, err := metadataStore.GetMetadata(projectID)
	if err != nil {
		logger.Warn("Failed to get metadata", models.Field{Key: "error", Value: err})
	}

	// Display project details
	fmt.Printf("Project: %s\n", project.Name)
	fmt.Printf("ID: %s\n", project.ID)
	fmt.Printf("Path: %s\n", project.RootPath)
	fmt.Printf("Type: %s\n", project.RepositoryType)
	fmt.Printf("Discovered: %s\n", project.DiscoveredAt)
	fmt.Printf("Updated: %s\n", project.UpdatedAt)

	if metadata != nil {
		fmt.Printf("\nMetadata:\n")
		fmt.Printf("  Git HEAD: %s\n", metadata.GitHead)
		if metadata.LanguageSummary != "" {
			fmt.Printf("  Languages: %s\n", metadata.LanguageSummary)
		}
		if metadata.FrameworkSummary != "" {
			fmt.Printf("  Frameworks: %s\n", metadata.FrameworkSummary)
		}
		if metadata.CapabilitiesSummary != "" {
			fmt.Printf("  Capabilities: %s\n", metadata.CapabilitiesSummary)
		}
		fmt.Printf("  Maturity score: %d\n", metadata.MaturityScore)
		if metadata.MaturityIndicators != "" {
			fmt.Printf("  Maturity indicators: %s\n", metadata.MaturityIndicators)
		}
		fmt.Printf("  Commits (total): %d  (last 90d): %d\n", metadata.CommitCount, metadata.CommitVelocity90d)
		fmt.Printf("  Contributors: %d  Tags: %d\n", metadata.ContributorCount, metadata.TagCount)
		if metadata.FirstCommitAt != "" {
			fmt.Printf("  First commit: %s\n", metadata.FirstCommitAt)
		}
		if metadata.RemoteURL != "" {
			fmt.Printf("  Remote: %s (published: %v)\n", metadata.RemoteURL, metadata.IsPublished)
		}
		if metadata.LastScanAt != "" {
			fmt.Printf("  Last scan: %s\n", metadata.LastScanAt)
		}
	}
}

func runDiscoverProjects(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()

	provider := config.NewProvider(cfgFile)
	cfg, err := provider.Load()
	if err != nil {
		logger.Error("Failed to load config", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if len(cfg.Discovery.ProjectRoots) == 0 {
		fmt.Println("No project roots configured.")
		fmt.Println("Configure roots with: portfolio config set-root <path>")
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		logger.Error("Failed to create database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("Failed to connect to database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		logger.Error("Failed to initialize database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	fmt.Println("Discovering projects...")
	fmt.Printf("Roots: %v\n\n", cfg.Discovery.ProjectRoots)

	// Create discoverer
	projectsStore := store.NewProjectStore(db.DB(), logger.Zap())
	osFS := fs.NewOSFilesystem()
	discLogger := logger.With("discovery")

	adapter := &discoveryStoreAdapter{store: projectsStore}
	provider2 := &rootsConfigProvider{roots: cfg.Discovery.ProjectRoots}

	discoverer := discovery.NewDiscoverer(osFS, provider2, adapter, discLogger, 10)
	result, err := discoverer.DiscoverProjects(cmd.Context())
	if err != nil {
		logger.Error("Discovery failed", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	fmt.Printf("Discovery completed:\n")
	fmt.Printf("  Discovered: %d new projects\n", result.Discovered)
	fmt.Printf("  Errors: %d\n", len(result.Errors))

	if result.Discovered > 0 {
		fmt.Printf("\nRun 'portfolio projects list' to see discovered projects.\n")
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, e := range result.Errors {
			fmt.Printf("  - %s: %v\n", e.DirPath, e.Err)
		}
	}
}

type discoveryStoreAdapter struct {
	store *store.ProjectStore
}

func (a *discoveryStoreAdapter) UpsertProject(p *discovery.Project) error {
	return a.store.UpsertProject(&models.Project{
		ID:             p.ID,
		Name:           p.Name,
		RootPath:       p.RootPath,
		RepositoryType: p.RepositoryType,
		DiscoveredAt:   p.DiscoveredAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      p.DiscoveredAt.Format("2006-01-02T15:04:05Z"),
	})
}

type rootsConfigProvider struct {
	roots []string
}

func (r *rootsConfigProvider) GetProjectRoots() ([]string, error) {
	return r.roots, nil
}

func (r *rootsConfigProvider) GetIgnoredPaths() []string {
	return []string{
		"node_modules", ".git", "vendor", "build", "dist", "target", "bin",
	}
}
