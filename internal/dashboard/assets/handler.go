package assets

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Handler handles static asset serving
type Handler struct {
	externalPath string
	useEmbedded  bool
	logger       Logger
}

// Logger interface for asset handler logging
type Logger interface {
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// NewHandler creates a new asset handler
func NewHandler(externalPath string, logger Logger) *Handler {
	h := &Handler{
		externalPath: externalPath,
		logger:       logger,
	}

	// Determine whether to use embedded or external assets
	if externalPath != "" {
		if _, err := os.Stat(externalPath); err == nil {
			h.logger.Info("using external dashboard assets", Field{Key: "path", Value: externalPath})
			h.useEmbedded = false
			return h
		} else {
			h.logger.Warn("external dashboard asset path not found, checking embedded assets",
				Field{Key: "path", Value: externalPath}, Field{Key: "error", Value: err})
		}
	}

	// Fall back to embedded assets
	if HasEmbeddedAssets() {
		h.logger.Info("using embedded dashboard assets")
		h.useEmbedded = true
		return h
	}

	h.logger.Warn("no dashboard assets available (external path not found, no embedded assets)")
	return h
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Security: Reject directory traversal attempts
	if strings.Contains(r.URL.Path, "..") {
		h.writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	path := r.URL.Path

	// Root request serves index.html (SPA entry point)
	if path == "/" || path == "" {
		path = "/index.html"
	}

	// Try to serve the file
	if h.useEmbedded {
		h.serveEmbedded(w, r, path)
	} else if h.externalPath != "" {
		h.serveExternal(w, r, path)
	} else {
		h.writeError(w, http.StatusNotFound, "asset not found: no assets available")
	}
}

// serveEmbedded serves from embedded filesystem
func (h *Handler) serveEmbedded(w http.ResponseWriter, r *http.Request, path string) {
	embeddedFS, err := DashboardFS()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to access embedded assets")
		h.logger.Error("failed to access embedded filesystem", Field{Key: "error", Value: err})
		return
	}

	// Remove leading slash for filesystem access
	fsPath := strings.TrimPrefix(path, "/")

	file, err := embeddedFS.Open(fsPath)
	if err != nil {
		// File not found - try SPA fallback
		if h.trySPAFallback(w, r) {
			return
		}
		h.writeError(w, http.StatusNotFound, fmt.Sprintf("asset not found: %s", path))
		return
	}
	defer file.Close()

	// Get file info for MIME type and modification time
	stat, err := file.Stat()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to read file info")
		return
	}

	// Check if it's a directory - don't allow directory listings
	if stat.IsDir() {
		if h.trySPAFallback(w, r) {
			return
		}
		h.writeError(w, http.StatusNotFound, "directory listing not allowed")
		return
	}

	// Set content type based on file extension
	h.setMimeType(w, path)

	// Set cache headers
	h.setCacheHeaders(w, path, stat.ModTime())

	// Serve the file content
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, file); err != nil {
		h.logger.Error("failed to write asset response", Field{Key: "path", Value: path}, Field{Key: "error", Value: err})
	}
}

// serveExternal serves from external filesystem
func (h *Handler) serveExternal(w http.ResponseWriter, r *http.Request, path string) {
	// Construct the full file system path
	fsPath := filepath.Join(h.externalPath, strings.TrimPrefix(path, "/"))

	file, err := os.Open(fsPath)
	if err != nil {
		// File not found - try SPA fallback
		if h.trySPAFallback(w, r) {
			return
		}
		h.writeError(w, http.StatusNotFound, fmt.Sprintf("asset not found: %s", path))
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to read file info")
		return
	}

	// Check if it's a directory - don't allow directory listings
	if stat.IsDir() {
		if h.trySPAFallback(w, r) {
			return
		}
		h.writeError(w, http.StatusNotFound, "directory listing not allowed")
		return
	}

	// Set content type based on file extension
	h.setMimeType(w, path)

	// Set cache headers
	h.setCacheHeaders(w, path, stat.ModTime())

	// Serve the file content
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, file); err != nil {
		h.logger.Error("failed to write asset response", Field{Key: "path", Value: path}, Field{Key: "error", Value: err})
	}
}

// trySPAFallback attempts to serve index.html for SPA routing
func (h *Handler) trySPAFallback(w http.ResponseWriter, r *http.Request) bool {
	// Only apply SPA fallback to GET requests for non-API paths
	if r.Method != "GET" {
		return false
	}

	// Don't fallback for API routes
	if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
		return false
	}

	// Try to serve index.html
	if h.useEmbedded {
		return h.serveIndexHTMLEmbedded(w)
	}
	return h.serveIndexHTMLExternal(w)
}

// serveIndexHTMLEmbedded serves index.html from embedded assets
func (h *Handler) serveIndexHTMLEmbedded(w http.ResponseWriter) bool {
	embeddedFS, err := DashboardFS()
	if err != nil {
		return false
	}

	file, err := embeddedFS.Open("index.html")
	if err != nil {
		return false
	}
	defer file.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
	return true
}

// serveIndexHTMLExternal serves index.html from external assets
func (h *Handler) serveIndexHTMLExternal(w http.ResponseWriter) bool {
	indexPath := filepath.Join(h.externalPath, "index.html")
	file, err := os.Open(indexPath)
	if err != nil {
		return false
	}
	defer file.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
	return true
}

// setMimeType sets the Content-Type header based on file extension
func (h *Handler) setMimeType(w http.ResponseWriter, path string) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".html", ".htm":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".woff", ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
	case ".ttf":
		w.Header().Set("Content-Type", "font/ttf")
	case ".eot":
		w.Header().Set("Content-Type", "application/vnd.ms-fontobject")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}

// setCacheHeaders sets appropriate cache headers based on file type
func (h *Handler) setCacheHeaders(w http.ResponseWriter, path string, modTime time.Time) {
	// Check if file has fingerprint (hash) in name
	base := filepath.Base(path)
	if containsFingerprint(base) {
		// Fingerprinted assets - cache aggressively
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else if base == "index.html" {
		// index.html - never cache
		w.Header().Set("Cache-Control", "no-cache")
	} else {
		// Non-fingerprinted assets - use ETag
		etag := fmt.Sprintf("\"%x\"", modTime.Unix())
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "no-cache")

		// Check If-None-Match
		if match := w.Header().Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
}

// containsFingerprint checks if filename contains a hash fingerprint
func containsFingerprint(filename string) bool {
	// Common patterns: main.a1b2c3.js, style.d4e5f6.css
	// Look for sequences of hex characters that are likely fingerprints
	parts := strings.Split(filename, ".")
	for i, part := range parts {
		// Skip the actual extension and potential longer hex strings
		if i == 0 || i == len(parts)-1 {
			continue
		}
		// Check if this part looks like a hash (8+ hex chars)
		if len(part) >= 8 && isHex(part) {
			return true
		}
	}
	return false
}

// isHex checks if a string is all hexadecimal characters
func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// writeError writes a JSON error response
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s", "message": "%s"}`, errorCodes[status], message)))
}

// errorCodes maps HTTP status codes to error codes
var errorCodes = map[int]string{
	http.StatusBadRequest:          "bad_request",
	http.StatusNotFound:            "not_found",
	http.StatusMethodNotAllowed:    "method_not_allowed",
	http.StatusInternalServerError: "internal_error",
}
