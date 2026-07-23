package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type ConfigurationStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewConfigurationStore(db *sql.DB, logger *zap.Logger) *ConfigurationStore {
	return &ConfigurationStore{db: db, logger: logger}
}

func (s *ConfigurationStore) Set(key, value string) error {
	return s.set(s.db, key, value)
}

func (s *ConfigurationStore) SetTx(tx *sql.Tx, key, value string) error {
	return s.set(tx, key, value)
}

func (s *ConfigurationStore) set(q Querier, key, value string) error {
	query := `
		INSERT OR REPLACE INTO configuration (key, value, updated_at)
		VALUES (?, ?, ?)
	`
	_, err := q.ExecContext(context.Background(), query, key, value, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to set configuration: %w", err)
	}
	return nil
}

func (s *ConfigurationStore) Get(key string) (*models.Configuration, error) {
	query := `SELECT key, value, updated_at FROM configuration WHERE key = ?`
	c := &models.Configuration{}
	err := s.db.QueryRow(query, key).Scan(&c.Key, &c.Value, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}
	return c, nil
}

func (s *ConfigurationStore) Delete(key string) error {
	_, err := s.db.Exec("DELETE FROM configuration WHERE key = ?", key)
	if err != nil {
		return fmt.Errorf("failed to delete configuration: %w", err)
	}
	return nil
}

func (s *ConfigurationStore) List() ([]*models.Configuration, error) {
	query := `SELECT key, value, updated_at FROM configuration ORDER BY key`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list configuration: %w", err)
	}
	defer rows.Close()

	var configs []*models.Configuration
	for rows.Next() {
		c := &models.Configuration{}
		if err := rows.Scan(&c.Key, &c.Value, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan configuration row: %w", err)
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}
