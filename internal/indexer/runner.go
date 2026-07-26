package indexer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

// MetaExtractor runs deterministic metadata extraction against a project root.
// Implemented by *metadata.Service; declared here so the indexer does not
// import the metadata package directly at call sites.
type MetaExtractor interface {
	Extract(rootPath string) (*models.Metadata, []models.Dependency, error)
}

type IndexRunner struct {
	docStore      *store.DocumentStore
	metaStore     *store.MetadataStore
	depStore      *store.DependencyStore
	metaExtractor MetaExtractor
	discoverer    *DocDiscoverer
	reader        *DocReader
	dedup         *DedupEngine
	cleaner       *OrphanCleaner
	fts           *FTSManager
	logger        *zap.Logger
	mu            *ProjectMutex
}

func NewIndexRunner(
	docStore *store.DocumentStore,
	metaStore *store.MetadataStore,
	depStore *store.DependencyStore,
	metaExtractor MetaExtractor,
	logger *zap.Logger,
) *IndexRunner {
	return &IndexRunner{
		docStore:      docStore,
		metaStore:     metaStore,
		depStore:      depStore,
		metaExtractor: metaExtractor,
		discoverer:    NewDocDiscoverer(),
		reader:        NewDocReader(1 << 20),
		dedup:         NewDedupEngine(docStore.StoreDB()),
		cleaner:       NewOrphanCleaner(docStore),
		fts:           NewFTSManager(docStore.StoreDB()),
		logger:        logger,
		mu:            NewProjectMutex(),
	}
}

type ProjectMutex struct {
	locks map[string]struct{}
}

func NewProjectMutex() *ProjectMutex {
	return &ProjectMutex{locks: make(map[string]struct{})}
}

func (pm *ProjectMutex) Lock(projectID string) bool {
	if _, exists := pm.locks[projectID]; exists {
		return false
	}
	pm.locks[projectID] = struct{}{}
	return true
}

func (pm *ProjectMutex) Unlock(projectID string) {
	delete(pm.locks, projectID)
}

func (r *IndexRunner) Run(ctx context.Context, projectID, rootPath string) (*IndexResult, error) {
	if !r.mu.Lock(projectID) {
		return nil, fmt.Errorf("already indexing project %s", projectID)
	}
	defer r.mu.Unlock(projectID)

	start := time.Now()
	result := &IndexResult{ProjectID: projectID}
	stats := &IndexStats{StartTime: start.Unix()}

	allDocs := r.discoverer.FindREADME(rootPath)
	allDocs = append(allDocs, r.discoverer.FindDocs(rootPath)...)
	allDocs = append(allDocs, r.discoverer.FindADRs(rootPath)...)
	allDocs = append(allDocs, r.discoverer.FindCHANGELOG(rootPath)...)

	if len(allDocs) == 0 {
		r.logger.Debug("no documentation files found", zap.String("project", projectID))
	}

	tx, err := r.docStore.StoreDB().Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var validPaths []string

	for _, doc := range allDocs {
		content, hash, err := r.reader.Read(doc.AbsPath)
		if err != nil {
			result.Errors = append(result.Errors, IndexError{
				Code: "READ_FAILED", Message: fmt.Sprintf("%s: %v", doc.RelPath, err),
			})
			stats.SkippedFiles++
			continue
		}

		action, err := r.dedup.ResolveTx(tx, projectID, doc.RelPath, hash)
		if err != nil {
			result.Errors = append(result.Errors, IndexError{
				Code: "DEDUP_FAILED", Message: fmt.Sprintf("%s: %v", doc.RelPath, err),
			})
			continue
		}

		if action == DedupSkip {
			stats.SkippedFiles++
			validPaths = append(validPaths, doc.RelPath)
			continue
		}

		docModel := &models.Document{
			ID:          uuid.New().String(),
			ProjectID:   projectID,
			Path:        doc.RelPath,
			Kind:        string(doc.Kind),
			Content:     content,
			ContentHash: hash,
			IndexedAt:   time.Now().UTC().Format(time.RFC3339),
		}

		if err := r.docStore.UpsertDocumentTx(tx, docModel); err != nil {
			result.Errors = append(result.Errors, IndexError{
				Code: "UPSERT_FAILED", Message: fmt.Sprintf("%s: %v", doc.RelPath, err),
			})
			continue
		}

		stats.BytesRead += int64(len(content))
		result.Documents++
		validPaths = append(validPaths, doc.RelPath)
	}

	if err := r.cleaner.CleanTx(tx, projectID, validPaths); err != nil {
		result.Errors = append(result.Errors, IndexError{
			Code: "ORPHAN_CLEAN_FAILED", Message: err.Error(),
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	if r.fts.HasFTS5() {
		if err := r.fts.Rebuild(ctx); err != nil {
			result.Errors = append(result.Errors, IndexError{
				Code: "FTS_BUILD_FAILED", Message: err.Error(),
			})
		} else {
			result.FTSRebuilt = true
		}
	}

	oldMeta, _ := r.metaStore.GetMetadata(projectID)
	oldHash := ""
	if oldMeta != nil {
		oldHash = oldMeta.DocumentationHash
	}

	docHash, err := r.computeDocumentationHash(projectID)
	if err != nil {
		r.logger.Warn("failed to compute documentation hash", zap.String("project", projectID), zap.Error(err))
	}

	// Extract all deterministic metadata (git, languages, frameworks,
	// dependencies, capabilities, maturity), attach indexer-owned fields
	// (DocumentationHash, LastScanAt), and persist in a single upsert. This is
	// the single correct metadata writer — it replaces the prior 2-field
	// INSERT OR REPLACE that clobbered every other column. Dependency rows are
	// replaced in the same pass.
	meta := &models.Metadata{ProjectID: projectID}
	var deps []models.Dependency
	if r.metaExtractor != nil {
		if extracted, extractedDeps, extractErr := r.metaExtractor.Extract(rootPath); extractErr != nil {
			r.logger.Warn("metadata extraction failed", zap.String("project", projectID), zap.Error(extractErr))
		} else if extracted != nil {
			meta = extracted
			deps = extractedDeps
		}
	}
	meta.ProjectID = projectID
	if docHash != "" {
		meta.DocumentationHash = docHash
		result.Documentation = docHash
		result.DocsChanged = docHash != oldHash
	}
	meta.LastScanAt = time.Now().UTC().Format(time.RFC3339)

	if err := r.metaStore.UpsertMetadata(meta); err != nil {
		r.logger.Warn("failed to update metadata", zap.String("project", projectID), zap.Error(err))
	}

	if deps != nil && r.depStore != nil {
		for i := range deps {
			deps[i].ProjectID = projectID
		}
		if err := r.depStore.ReplaceDependencies(projectID, deps); err != nil {
			r.logger.Warn("failed to replace dependencies", zap.String("project", projectID), zap.Error(err))
		}
	}

	result.Duration = time.Since(start).String()
	result.Skipped = stats.SkippedFiles

	return result, nil
}

func (r *IndexRunner) computeDocumentationHash(projectID string) (string, error) {
	docs, err := r.docStore.ListDocuments(projectID)
	if err != nil {
		return "", err
	}

	var hashes []string
	for _, d := range docs {
		hashes = append(hashes, d.ContentHash)
	}
	sort.Strings(hashes)

	var joined string
	for _, h := range hashes {
		joined += h
	}

	combined := sha256.Sum256([]byte(joined))
	return fmt.Sprintf("%x", combined), nil
}
