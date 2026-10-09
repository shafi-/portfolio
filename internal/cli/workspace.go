package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"project-dash/internal/config"
	"project-dash/internal/database"
	"project-dash/internal/logging"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

// workspaceCmd is the parent of the workspace management commands. Workspaces
// are grouping-only (ADR-023): they never affect scan or discovery behavior.
var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspaces — named groups of projects",
	Long: `Manage workspaces: named groups of projects, e.g. the microservices
that together form one product.

Workspaces are grouping-only. They scope queries and reporting (such as the
analysis freshness report in 'workspace show') but never change scan,
discovery, or any other engine behavior.

Available Subcommands:
  • create - Create a new workspace
  • add    - Add a project to a workspace
  • remove - Remove a project from a workspace
  • list   - List all workspaces with member counts
  • show   - Show a workspace's members and their analysis freshness

Examples:
  # Group the services of one product
  portfolio workspace create commerce
  portfolio workspace add commerce orders-service
  portfolio workspace add commerce payments-service
  portfolio workspace show commerce`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var workspaceCreateCmd = &cobra.Command{
	Use:   "create <name> [-d description]",
	Short: "Create a new workspace",
	Args:  cobra.ExactArgs(1),
	Run:   runWorkspaceCreate,
}

var workspaceAddCmd = &cobra.Command{
	Use:   "add <workspace> <project-id-or-name>",
	Short: "Add a project to a workspace",
	Args:  cobra.ExactArgs(2),
	Run:   runWorkspaceAdd,
}

var workspaceRemoveCmd = &cobra.Command{
	Use:   "remove <workspace> <project-id-or-name>",
	Short: "Remove a project from a workspace",
	Args:  cobra.ExactArgs(2),
	Run:   runWorkspaceRemove,
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces with member counts",
	Args:  cobra.NoArgs,
	Run:   runWorkspaceList,
}

var workspaceShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show a workspace's members and their analysis freshness",
	Args:  cobra.ExactArgs(1),
	Run:   runWorkspaceShow,
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
	workspaceCreateCmd.Flags().StringP("description", "d", "", "Workspace description")
	workspaceCmd.AddCommand(
		workspaceCreateCmd, workspaceAddCmd, workspaceRemoveCmd,
		workspaceListCmd, workspaceShowCmd,
	)
}

// openWorkspaceStores loads config, connects to the knowledge store, and
// returns the stores the workspace commands need. The caller must defer
// db.Close().
func openWorkspaceStores() (*database.Database, *store.WorkspaceStore, *store.ProjectStore, *store.AnalysisStore, *store.MetadataStore, error) {
	logger := logging.GetGlobalLogger()

	provider := config.NewProvider(cfgFile)
	cfg, err := provider.Load()
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create database: %w", err)
	}

	if err := db.Connect(); err != nil {
		db.Close()
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Apply pending migrations (idempotent) so a database created before
	// workspaces existed is brought up to date on first use.
	if err := db.Initialize(); err != nil {
		db.Close()
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	zapLogger := logger.Zap()
	return db,
		store.NewWorkspaceStore(db.DB(), zapLogger),
		store.NewProjectStore(db.DB(), zapLogger),
		store.NewAnalysisStore(db.DB(), zapLogger),
		store.NewMetadataStore(db.DB(), zapLogger),
		nil
}

// resolveWorkspace finds a workspace by name (the human-facing handle).
func resolveWorkspace(ws *store.WorkspaceStore, nameOrID string) (*models.Workspace, error) {
	w, err := ws.GetWorkspaceByName(nameOrID)
	if err != nil {
		return nil, err
	}
	if w != nil {
		return w, nil
	}
	// Fall back to ID so scripts can address workspaces robustly.
	return ws.GetWorkspace(nameOrID)
}

// resolveProjectArg finds a project by exact ID, then by exact name, then by
// unique name substring. Returns (project, nil) or a helpful error listing
// ambiguous matches.
func resolveProjectArg(ps *store.ProjectStore, nameOrID string) (*models.Project, error) {
	p, err := ps.GetProject(nameOrID)
	if err != nil {
		return nil, err
	}
	if p != nil {
		return p, nil
	}

	projects, err := ps.ListProjects()
	if err != nil {
		return nil, err
	}
	var exact []*models.Project
	var partial []*models.Project
	for _, cand := range projects {
		if cand.Name == nameOrID {
			exact = append(exact, cand)
		}
		if containsFold(cand.Name, nameOrID) {
			partial = append(partial, cand)
		}
	}
	switch {
	case len(exact) == 1:
		return exact[0], nil
	case len(exact) > 1:
		names := make([]string, 0, len(exact))
		for _, cand := range exact {
			names = append(names, fmt.Sprintf("%s (%s)", cand.Name, cand.ID))
		}
		return nil, fmt.Errorf(
			"multiple projects are named %q — use a project ID: %v", nameOrID, names)
	case len(partial) == 1:
		return partial[0], nil
	case len(partial) > 1:
		names := make([]string, 0, len(partial))
		for _, cand := range partial {
			names = append(names, cand.Name)
		}
		return nil, fmt.Errorf(
			"%q matches several projects — be more specific: %v", nameOrID, names)
	}
	return nil, fmt.Errorf("project not found: %s", nameOrID)
}

func containsFold(s, sub string) bool {
	n := len(sub)
	if n == 0 {
		return true
	}
	for i := 0; i+n <= len(s); i++ {
		if equalFoldASCII(s[i:i+n], sub) {
			return true
		}
	}
	return false
}

func equalFoldASCII(a, b string) bool {
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func runWorkspaceCreate(cmd *cobra.Command, args []string) {
	name := args[0]
	if name == "" {
		fmt.Fprintln(os.Stderr, "Error: workspace name cannot be empty")
		os.Exit(1)
	}

	description, _ := cmd.Flags().GetString("description")

	db, ws, _, _, _, err := openWorkspaceStores()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	workspace := &models.Workspace{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := ws.CreateWorkspace(workspace); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created workspace: %s\n", workspace.Name)
	fmt.Printf("ID: %s\n", workspace.ID)
	fmt.Println("Add projects with: portfolio workspace add " + workspace.Name + " <project-id-or-name>")
}

func runWorkspaceAdd(cmd *cobra.Command, args []string) {
	workspaceRef, projectRef := args[0], args[1]

	db, ws, ps, _, _, err := openWorkspaceStores()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	workspace, err := resolveWorkspace(ws, workspaceRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if workspace == nil {
		fmt.Fprintf(os.Stderr, "Error: workspace not found: %s\n", workspaceRef)
		os.Exit(1)
	}

	project, err := resolveProjectArg(ps, projectRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := ws.AddProjectToWorkspace(workspace.ID, project.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Added %s to workspace %s\n", project.Name, workspace.Name)
}

func runWorkspaceRemove(cmd *cobra.Command, args []string) {
	workspaceRef, projectRef := args[0], args[1]

	db, ws, ps, _, _, err := openWorkspaceStores()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	workspace, err := resolveWorkspace(ws, workspaceRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if workspace == nil {
		fmt.Fprintf(os.Stderr, "Error: workspace not found: %s\n", workspaceRef)
		os.Exit(1)
	}

	project, err := resolveProjectArg(ps, projectRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := ws.RemoveProjectFromWorkspace(workspace.ID, project.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Removed %s from workspace %s\n", project.Name, workspace.Name)
}

func runWorkspaceList(cmd *cobra.Command, args []string) {
	db, ws, _, _, _, err := openWorkspaceStores()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	workspaces, err := ws.ListWorkspaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces defined.")
		fmt.Println("Use 'portfolio workspace create <name>' to create one.")
		return
	}

	fmt.Println("Portfolio Workspaces")
	fmt.Println("====================")
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tMEMBERS\tDESCRIPTION")
	for _, workspace := range workspaces {
		count, err := ws.CountWorkspaceProjects(workspace.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(w, "%s\t%d\t%s\n", workspace.Name, count, workspace.Description)
	}
	w.Flush()
	fmt.Println()
	fmt.Printf("Total: %d workspace(s)\n", len(workspaces))
}

func runWorkspaceShow(cmd *cobra.Command, args []string) {
	db, ws, _, analyses, metadata, err := openWorkspaceStores()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	workspace, err := resolveWorkspace(ws, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if workspace == nil {
		fmt.Fprintf(os.Stderr, "Error: workspace not found: %s\n", args[0])
		os.Exit(1)
	}

	projects, err := ws.ListWorkspaceProjects(workspace.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Workspace: %s\n", workspace.Name)
	if workspace.Description != "" {
		fmt.Printf("Description: %s\n", workspace.Description)
	}
	fmt.Println()

	if len(projects) == 0 {
		fmt.Println("No projects in this workspace yet.")
		fmt.Printf("Use 'portfolio workspace add %s <project-id-or-name>'.\n", workspace.Name)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "PROJECT\tANALYSIS FRESHNESS\tID")
	for _, project := range projects {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			project.Name,
			analysisFreshnessLabel(analyses, metadata, project),
			project.ID,
		)
	}
	w.Flush()
	fmt.Println()
	fmt.Printf("Total: %d project(s)\n", len(projects))
}
