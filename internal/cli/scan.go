package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"project-dash/internal/config"
	"project-dash/internal/database"
	"project-dash/internal/indexer"
	"project-dash/internal/logging"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

var scanProjectID string

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Extract metadata and index documentation for projects",
	Long: `Run the deterministic engine over discovered projects.

Extracts git, language, framework, dependency, capability, and maturity facts,
indexes documentation, and rebuilds full-text search. Performs no AI analysis —
it only populates the deterministic facts that downstream importance ranking and
agents build on.

By default scans all discovered projects. Use --project <id> to scan a single one.`,
	Run: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().StringVar(&scanProjectID, "project", "", "scan a single project by ID")
}

func runScan(cmd *cobra.Command, args []string) {
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

	// Ensure schema: runs pending migrations (including the metadata-extras v8
	// migration) before extraction writes to the new columns.
	if err := db.Initialize(); err != nil {
		logger.Error("Failed to initialize database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	projectsStore := store.NewProjectStore(db.DB(), logger.Zap())
	metaStore := store.NewMetadataStore(db.DB(), logger.Zap())

	idx := indexer.NewIndexer(db.DB(), logger.Zap()).WithProjectLister(projectsStore)

	ctx := context.Background()

	if scanProjectID != "" {
		project, err := projectsStore.GetProject(scanProjectID)
		if err != nil {
			logger.Error("Failed to get project", models.Field{Key: "error", Value: err})
			os.Exit(1)
		}
		if project == nil {
			fmt.Printf("Project not found: %s\n", scanProjectID)
			os.Exit(1)
		}
		res, err := idx.IndexProject(ctx, project.ID, project.RootPath)
		if err != nil {
			logger.Error("Scan failed",
				models.Field{Key: "project", Value: project.ID},
				models.Field{Key: "error", Value: err})
			os.Exit(1)
		}
		printScanSummary(metaStore, res)
		return
	}

	start := time.Now()
	results, err := idx.IndexAll(ctx)
	if err != nil {
		logger.Error("Scan failed", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	// Deterministic ordering for stable output.
	ids := make([]string, 0, len(results))
	for id := range results {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	fmt.Printf("Scanned %d project(s) in %s\n\n", len(ids), time.Since(start).Round(time.Millisecond))
	for _, id := range ids {
		printScanSummary(metaStore, results[id])
	}
}

// printScanSummary prints a per-project summary: indexer stats plus the
// deterministic metadata facts now persisted for that project.
func printScanSummary(metaStore *store.MetadataStore, res *indexer.IndexResult) {
	fmt.Printf("%s\n", res.ProjectID)
	fmt.Printf("  Documents indexed: %d (changed: %v, skipped: %d)\n", res.Documents, res.DocsChanged, res.Skipped)
	if res.FTSRebuilt {
		fmt.Printf("  FTS: rebuilt\n")
	}

	if meta, err := metaStore.GetMetadata(res.ProjectID); err == nil && meta != nil {
		fmt.Printf("  Maturity score: %d\n", meta.MaturityScore)
		if meta.MaturityIndicators != "" {
			fmt.Printf("  Maturity indicators: %s\n", meta.MaturityIndicators)
		}
		if meta.CapabilitiesSummary != "" {
			fmt.Printf("  Capabilities: %s\n", meta.CapabilitiesSummary)
		}
		fmt.Printf("  Commits (90d): %d  Contributors: %d  Tags: %d\n",
			meta.CommitVelocity90d, meta.ContributorCount, meta.TagCount)
		if meta.RemoteURL != "" {
			fmt.Printf("  Remote: %s (published: %v)\n", meta.RemoteURL, meta.IsPublished)
		}
	}

	if len(res.Errors) > 0 {
		fmt.Printf("  Errors: %d\n", len(res.Errors))
		for _, e := range res.Errors {
			fmt.Printf("    - [%s] %s\n", e.Code, e.Message)
		}
	}
	fmt.Println()
}
