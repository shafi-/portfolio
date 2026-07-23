package indexer

import "project-dash/pkg/models"

type DocFile struct {
	RelPath string
	AbsPath string
	Kind    models.DocumentKind
}

type IndexResult struct {
	ProjectID     string       `json:"project_id"`
	Documents     int          `json:"documents"`
	BytesIndexed  int64        `json:"bytes_indexed"`
	Duration      string       `json:"duration"`
	FTSRebuilt    bool         `json:"fts_rebuilt"`
	Skipped       int          `json:"skipped"`
	Errors        []IndexError `json:"errors,omitempty"`
	Documentation string       `json:"documentation_hash"`
}

type IndexStats struct {
	TotalFiles   int
	SkippedFiles int
	BytesRead    int64
	BytesStored  int64
	StartTime    int64
	EndTime      int64
}

type DedupAction string

const (
	DedupSkip   DedupAction = "skip"
	DedupInsert DedupAction = "insert"
	DedupUpdate DedupAction = "update"
)

type SearchResult struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"project_id"`
	Path        string  `json:"path"`
	Kind        string  `json:"kind"`
	Content     string  `json:"content"`
	ContentHash string  `json:"content_hash"`
	IndexedAt   string  `json:"indexed_at"`
	Rank        float64 `json:"rank"`
}

type IndexError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
