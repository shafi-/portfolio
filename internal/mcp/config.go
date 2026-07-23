package mcp

import (
	"database/sql"

	"project-dash/internal/fs"
	"project-dash/internal/logging"
	"project-dash/internal/store"
)

type Config struct {
	DB     *sql.DB
	Logger *logging.Logger
	Roots  []string
}

func (c *Config) buildServer() *Server {
	zapLogger := c.Logger.Zap()
	return &Server{
		db:            c.DB,
		logger:        c.Logger,
		projects:      store.NewProjectStore(c.DB, zapLogger),
		metadata:      store.NewMetadataStore(c.DB, zapLogger),
		documents:     store.NewDocumentStore(c.DB, zapLogger),
		analyses:      store.NewAnalysisStore(c.DB, zapLogger),
		features:      store.NewFeatureStore(c.DB, zapLogger),
		technologies:  store.NewTechnologyStore(c.DB, zapLogger),
		relationships: store.NewRelationshipStore(c.DB, zapLogger),
		dependencies:  store.NewDependencyStore(c.DB, zapLogger),
		configuration: store.NewConfigurationStore(c.DB, zapLogger),
		osFS:          fs.NewOSFilesystem(),
		roots:         c.Roots,
	}
}
