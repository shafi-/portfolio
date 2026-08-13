package dashboard

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

// ServingMode defines how dashboard assets are served
type ServingMode string

const (
	ModeEmbedded ServingMode = "embedded"
	ModeExternal ServingMode = "external"
)

// AssetServer handles serving dashboard assets
type AssetServer struct {
	embedded http.Handler
	external http.Handler
	mode     ServingMode
}

// NewAssetServer creates a new asset server
func NewAssetServer(mode ServingMode, externalPath string) (*AssetServer, error) {
	server := &AssetServer{mode: mode}

	if mode == ModeEmbedded {
		// Embedded mode - handler will be set by SetEmbeddedFS or use FileServer
		fs, err := FileServer()
		if err != nil {
			return nil, fmt.Errorf("failed to create embedded file server: %w", err)
		}
		server.embedded = fs
	} else if mode == ModeExternal {
		if externalPath == "" {
			return nil, errors.New("external path required for external mode")
		}
		if _, err := os.Stat(externalPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("external path does not exist: %s", externalPath)
		}
		server.external = http.FileServer(http.Dir(externalPath))
	}

	return server, nil
}

// SetEmbeddedFS sets the embedded filesystem handler
func (s *AssetServer) SetEmbeddedFS(fs interface{}) {
	if handler, ok := fs.(http.Handler); ok {
		s.embedded = handler
	}
}

// ServeHTTP implements http.Handler
func (s *AssetServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.mode == ModeEmbedded {
		if s.embedded == nil {
			http.Error(w, "Embedded assets not available", http.StatusServiceUnavailable)
			return
		}
		s.embedded.ServeHTTP(w, r)
	} else {
		if s.external == nil {
			http.Error(w, "External assets not available", http.StatusServiceUnavailable)
			return
		}
		s.external.ServeHTTP(w, r)
	}
}
