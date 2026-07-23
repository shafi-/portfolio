package integration

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)

type DatabaseStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDatabaseStore(db *sql.DB, logger *zap.Logger) *DatabaseStore {
	return &DatabaseStore{
		db:     db,
		logger: logger,
	}
}

func (s *DatabaseStore) SaveIntegration(ctx context.Context, meta IntegrationMeta) error {
	data, err := MetaToJSON(meta)
	if err != nil {
		s.logger.Error("Failed to serialize integration metadata",
			zap.String("name", meta.Name),
			zap.Error(err))
		return err
	}

	key := GetMetaKey(meta.Name)
	query := `INSERT OR REPLACE INTO configuration (key, value) VALUES (?, ?)`

	_, err = s.db.ExecContext(ctx, query, key, string(data))
	if err != nil {
		s.logger.Error("Failed to save integration metadata",
			zap.String("name", meta.Name),
			zap.Error(err))
		return err
	}

	s.logger.Info("Integration metadata saved", zap.String("name", meta.Name))
	return nil
}

func (s *DatabaseStore) GetIntegration(ctx context.Context, name string) (*IntegrationMeta, error) {
	key := GetMetaKey(name)
	query := `SELECT value FROM configuration WHERE key = ?`

	var value string
	err := s.db.QueryRowContext(ctx, query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		s.logger.Error("Failed to get integration metadata",
			zap.String("name", name),
			zap.Error(err))
		return nil, err
	}

	meta, err := MetaFromJSON([]byte(value))
	if err != nil {
		s.logger.Error("Failed to deserialize integration metadata",
			zap.String("name", name),
			zap.Error(err))
		return nil, err
	}

	return &meta, nil
}

func (s *DatabaseStore) ListIntegrations(ctx context.Context) ([]IntegrationMeta, error) {
	prefix := "integration:"
	query := `SELECT value FROM configuration WHERE key LIKE ? || '%'`

	rows, err := s.db.QueryContext(ctx, query, prefix)
	if err != nil {
		s.logger.Error("Failed to list integrations", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var metas []IntegrationMeta
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			s.logger.Error("Failed to scan integration row", zap.Error(err))
			continue
		}

		meta, err := MetaFromJSON([]byte(value))
		if err != nil {
			s.logger.Error("Failed to deserialize integration metadata", zap.Error(err))
			continue
		}

		metas = append(metas, meta)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating integrations", zap.Error(err))
		return nil, err
	}

	return metas, nil
}

func (s *DatabaseStore) DeleteIntegration(ctx context.Context, name string) error {
	key := GetMetaKey(name)
	query := `DELETE FROM configuration WHERE key = ?`

	result, err := s.db.ExecContext(ctx, query, key)
	if err != nil {
		s.logger.Error("Failed to delete integration metadata",
			zap.String("name", name),
			zap.Error(err))
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	s.logger.Info("Integration metadata deleted", zap.String("name", name))
	return nil
}
