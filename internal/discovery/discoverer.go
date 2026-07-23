package discovery

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"project-dash/internal/fs"
	"project-dash/internal/logging"
	"project-dash/pkg/models"

	"github.com/google/uuid"
)

// Discoverer orchestrates project discovery across configured root directories
type Discoverer struct {
	osFS        fs.Filesystem
	config      ConfigProvider
	store       ProjectStore
	logger      *logging.Logger
	mutex       sync.Mutex
	running     atomic.Bool
	maxDepth    int
	ignorePaths []string
}

// ConfigProvider interface for accessing configuration
type ConfigProvider interface {
	GetProjectRoots() ([]string, error)
	GetIgnoredPaths() []string
}

// ProjectStore interface for storing discovered projects
type ProjectStore interface {
	UpsertProject(project *Project) error
}

// Project represents a discovered project
type Project struct {
	ID             string
	Name           string
	RootPath       string
	RepositoryType string
	DiscoveredAt   time.Time
}

// NewDiscoverer creates a new Discoverer instance
func NewDiscoverer(
	osFS fs.Filesystem,
	config ConfigProvider,
	store ProjectStore,
	logger *logging.Logger,
	maxDepth int,
) *Discoverer {
	return &Discoverer{
		osFS:     osFS,
		config:   config,
		store:    store,
		logger:   logger.With("discovery"),
		maxDepth: maxDepth,
	}
}

// DiscoverProjects walks all configured root directories and discovers Git repositories
func (d *Discoverer) DiscoverProjects(ctx context.Context) (*DiscoveryResult, error) {
	// Acquire mutex to prevent concurrent discovery operations
	if !d.mutex.TryLock() {
		return nil, &ConcurrentDiscoveryError{CurrentOperation: "DiscoverProjects"}
	}
	defer d.mutex.Unlock()

	// Mark as running
	d.running.Store(true)
	d.logger.Info("Starting project discovery")

	// Ensure we mark ourselves as not running when we exit
	defer func() { d.running.Store(false) }()

	// Get configuration
	roots, err := d.config.GetProjectRoots()
	if err != nil {
		return nil, fmt.Errorf("failed to get project roots: %w", err)
	}

	// Get ignored paths from config
	d.ignorePaths = d.config.GetIgnoredPaths()

	// Initialize result
	result := &DiscoveryResult{
		Errors:    []DiscoveryError{},
		RootStats: make(map[string]RootStat),
	}

	// Process each root directory
	for _, root := range roots {
		root = filepath.Clean(root)
		startTime := time.Now()

		d.logger.Info("Processing root directory",
			models.Field{Key: "root", Value: root},
		)

		rootStat, errors := d.discoverInRoot(ctx, root)

		// Record statistics
		duration := time.Since(startTime)
		rootStat.DurationMs = duration.Milliseconds()
		result.RootStats[root] = rootStat

		// Update totals
		result.Discovered += rootStat.Discovered
		result.Errors = append(result.Errors, errors...)

		d.logger.Info("Root directory processed",
			models.Field{Key: "root", Value: root},
			models.Field{Key: "discovered", Value: rootStat.Discovered},
			models.Field{Key: "errors", Value: rootStat.Errors},
			models.Field{Key: "duration_ms", Value: rootStat.DurationMs},
		)
	}

	d.logger.Info("Project discovery completed",
		models.Field{Key: "total_discovered", Value: result.Discovered},
		models.Field{Key: "total_errors", Value: len(result.Errors)},
	)

	return result, nil
}

// discoverInRoot discovers projects in a single root directory
func (d *Discoverer) discoverInRoot(ctx context.Context, root string) (RootStat, []DiscoveryError) {
	stat := RootStat{}
	errors := []DiscoveryError{}

	// Create detector and walker for this root
	detector := NewDetector(d.osFS)
	walker := NewWalker(d.osFS, detector, d.ignorePaths, d.maxDepth, d.logger)

	// Walk the directory tree
	err := walker.Walk(ctx, root, func(path string, event WalkEvent, err error) error {
		switch event {
		case EventFoundRepo:
			// Create project record
			project := d.createProject(path)
			if err := d.store.UpsertProject(project); err != nil {
				d.logger.Error("Failed to upsert project",
					models.Field{Key: "path", Value: path},
					models.Field{Key: "error", Value: err},
				)
				stat.Errors++
				errors = append(errors, DiscoveryError{
					RootPath: root,
					DirPath:  path,
					Err:      fmt.Errorf("failed to store project: %w", err),
				})
				return nil // Continue walking despite error
			}
			stat.Discovered++
			d.logger.Debug("Discovered project",
				models.Field{Key: "path", Value: path},
			)

		case EventError:
			// Log error but continue walking
			if err != nil {
				d.logger.Warn("Error during walk",
					models.Field{Key: "path", Value: path},
					models.Field{Key: "error", Value: err},
				)
				stat.Errors++
				errors = append(errors, DiscoveryError{
					RootPath: root,
					DirPath:  path,
					Err:      err,
				})
				return nil // Continue walking
			}

		case EventSkipped:
			d.logger.Debug("Skipped directory",
				models.Field{Key: "path", Value: path},
			)

		case EventEnterDir:
			d.logger.Debug("Entering directory",
				models.Field{Key: "path", Value: path},
			)
		}

		return nil
	})

	// Check if walk was cancelled by context
	if err == context.Canceled || err == context.DeadlineExceeded {
		d.logger.Warn("Discovery cancelled by context")
		return stat, errors
	}

	// Check for other errors
	if err != nil {
		d.logger.Error("Walk error",
			models.Field{Key: "root", Value: root},
			models.Field{Key: "error", Value: err},
		)
		// Add the walk error itself
		errors = append(errors, DiscoveryError{
			RootPath: root,
			Err:      err,
		})
	}

	return stat, errors
}

// createProject creates a Project record from a discovered path
func (d *Discoverer) createProject(path string) *Project {
	// Generate unique ID for the project
	id := uuid.New().String()

	// Extract project name from path
	name := filepath.Base(path)

	// Detect project type using marker detector
	markerDetector := NewMarkerDetector(d.osFS)
	projectType := markerDetector.DetectProjectType(path)

	return &Project{
		ID:             id,
		Name:           name,
		RootPath:       path,
		RepositoryType: projectType,
		DiscoveredAt:   time.Now().UTC(),
	}
}

// IsRunning returns true if a discovery operation is currently running
func (d *Discoverer) IsRunning() bool {
	return d.running.Load()
}
