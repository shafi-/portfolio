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

type IndexRunner struct {
	docStore   *store.DocumentStore
	metaStore  *store.MetadataStore
	discoverer *DocDiscoverer
	reader     *DocReader
	dedup      *DedupEngine
	cleaner    *OrphanCleaner
	fts        *FTSManager
	logger     *zap.Logger
	mu         *ProjectMutex
}

func NewIndexRunner(
	docStore *store.DocumentStore,
	metaStore *store.MetadataStore,
	logger *zap.Logger,
) *IndexRunner {
	return &IndexRunner{
		docStore:   docStore,
		metaStore:  metaStore,
		discoverer: NewDocDiscoverer(),
		reader:     NewDocReader(1 << 20),
		dedup:      NewDedupEngine(docStore.StoreDB()),
		cleaner:    NewOrphanCleaner(docStore),
		fts:        NewFTSManager(docStore.StoreDB()),
		logger:     logger,
		mu:         NewProjectMutex(),
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
	result := &IndexResult{
		ProjectID: projectID,
	}
	stats := &IndexStats{
		StartTime: start.Unix(),
	}

	var allDocs []DocFile

	allDocs = append(allDocs, r.discoverer.FindREADME(rootPath)...)
	allDocs = append(allDocs, r.discoverer.FindDocs(rootPath)...)
	allDocs = append(allDocs, r.discoverer.FindADRs(rootPath)...)
	allDocs = append(allDocs, r.discoverer.FindCHANGELOG(rootPath)...)

	if len(allDocs) == 0 {
		r.logger.Debug("no documentation files found", zap.String("project", projectID))
	}

	var validPaths []string

	for _, doc := range allDocs {
		content, hash, err := r.reader.Read(doc.AbsPath)
		if err != nil {
			result.Errors = append(result.Errors, IndexError{
				Code:    "READ_FAILED",
				Message: fmt.Sprintf("%s: %v", doc.RelPath, err),
			})
			stats.SkippedFiles++
			continue
		}

		action, err := r.dedup.Resolve(projectID, doc.RelPath, hash)
		if err != nil {
			result.Errors = append(result.Errors, IndexError{
				Code:    "DEDUP_FAILED",
				Message: fmt.Sprintf("%s: %v", doc.RelPath, err),
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

		if err := r.docStore.UpsertDocument(docModel); err != nil {
			result.Errors = append(result.Errors, IndexError{
				Code:    "UPSERT_FAILED",
				Message: fmt.Sprintf("%s: %v", doc.RelPath, err),
			})
			continue
		}

		stats.BytesRead += int64(len(content))
		result.Documents++
		validPaths = append(validPaths, doc.RelPath)
	}

	if err := r.cleaner.Clean(ctx, projectID, validPaths); err != nil {
		result.Errors = append(result.Errors, IndexError{
			Code:    "ORPHAN_CLEAN_FAILED",
			Message: err.Error(),
		})
	}

	if r.fts.HasFTS5() {
		if err := r.fts.Rebuild(ctx); err != nil {
			result.Errors = append(result.Errors, IndexError{
				Code:    "FTS_BUILD_FAILED",
				Message: err.Error(),
			})
		} else {
			result.FTSRebuilt = true
		}
	}

	docHash, err := r.computeDocumentationHash(projectID)
	if err != nil {
		r.logger.Warn("failed to compute documentation hash", zap.String("project", projectID), zap.Error(err))
	} else {
		result.Documentation = docHash
		meta := &models.Metadata{
			ProjectID:         projectID,
			DocumentationHash: docHash,
			LastScanAt:        time.Now().UTC().Format(time.RFC3339),
		}
		if err := r.metaStore.UpsertMetadata(meta); err != nil {
			r.logger.Warn("failed to update metadata", zap.String("project", projectID), zap.Error(err))
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
	if joined == "" {
		joined = ""
	}

	combined := sha256.Sum256([]byte(joined))
	return fmt.Sprintf("%x", combined), nil
}
