package mcp

import (
	"context"
	"database/sql"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"project-dash/internal/fs"
	"project-dash/internal/logging"
	"project-dash/internal/store"
	"project-dash/internal/version"
)

type serverTool struct {
	Tool    mcp.Tool
	Handler server.ToolHandlerFunc
}

type Server struct {
	projects      *store.ProjectStore
	metadata      *store.MetadataStore
	documents     *store.DocumentStore
	analyses      *store.AnalysisStore
	features      *store.FeatureStore
	technologies  *store.TechnologyStore
	relationships *store.RelationshipStore
	dependencies  *store.DependencyStore
	configuration *store.ConfigurationStore
	db            *sql.DB
	logger        *logging.Logger
	osFS          fs.Filesystem
	roots         []string
	mcp           *server.MCPServer
}

func New(cfg *Config) *Server {
	s := cfg.buildServer()

	s.mcp = server.NewMCPServer(
		"portfolio",
		version.Version(),
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	s.registerTools()

	return s
}

func (s *Server) Serve(ctx context.Context) error {
	s.logger.Info("starting MCP server on stdio")
	return server.ServeStdio(s.mcp)
}

func (s *Server) registerTools() {
	discoveryTools := s.discoveryTools()
	for _, t := range discoveryTools {
		s.mcp.AddTool(t.Tool, t.Handler)
	}

	searchTools := s.searchTools()
	for _, t := range searchTools {
		s.mcp.AddTool(t.Tool, t.Handler)
	}

	analysisTools := s.analysisTools()
	for _, t := range analysisTools {
		s.mcp.AddTool(t.Tool, t.Handler)
	}

	configTools := s.configTools()
	for _, t := range configTools {
		s.mcp.AddTool(t.Tool, t.Handler)
	}

	relationshipTools := s.relationshipTools()
	for _, t := range relationshipTools {
		s.mcp.AddTool(t.Tool, t.Handler)
	}

	codeTools := s.codeContentTools()
	for _, t := range codeTools {
		s.mcp.AddTool(t.Tool, t.Handler)
	}

	featureTools := s.featureTools()
	for _, t := range featureTools {
		s.mcp.AddTool(t.Tool, t.Handler)
	}

	technologyTools := s.technologyTools()
	for _, t := range technologyTools {
		s.mcp.AddTool(t.Tool, t.Handler)
	}
}
