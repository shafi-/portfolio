package database

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"project-dash/pkg/models"
)

// GetDatabaseKey returns the database access key
// Checks multiple sources in order of priority:
// 1. Environment variable PORTFOLIO_DB_KEY
// 2. Secure key file (~/.config/portfolio/db_key or equivalent)
// 3. Config file (for backward compatibility)
// 4. Generate new key if none exists
func GetDatabaseKey() (string, error) {
	// Priority 1: Environment variable (for testing/deployment)
	if key := os.Getenv("PORTFOLIO_DB_KEY"); key != "" {
		return key, nil
	}

	// Priority 2: Secure key file
	keyFile := models.GetDatabaseKeyPath()
	if key, err := readKeyFile(keyFile); err == nil {
		return key, nil
	}

	// Priority 3: Generate new key (first run)
	key, err := generateNewKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate database key: %w", err)
	}
	if err := writeKeyFile(keyFile, key); err != nil {
		return "", fmt.Errorf("failed to write database key: %w", err)
	}

	return key, nil
}

// readKeyFile reads the database key from secure storage
func readKeyFile(keyFile string) (string, error) {
	keyBytes, err := os.ReadFile(keyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("key file does not exist: %s", keyFile)
		}
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	key := string(keyBytes)
	if key == "" {
		return "", fmt.Errorf("empty key in key file")
	}

	return key, nil
}

// writeKeyFile writes the database key to secure storage
func writeKeyFile(keyFile, key string) error {
	// Ensure directory exists
	keyDir := filepath.Dir(keyFile)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Write key with secure permissions (only owner can read/write)
	if err := os.WriteFile(keyFile, []byte(key), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// generateNewKey generates a new secure database key
func generateNewKey() (string, error) {
	// Generate 32 bytes (256 bits) for AES-256
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(keyBytes), nil
}

// ValidateKeyAccess checks if the database key is accessible
func ValidateKeyAccess() error {
	_, err := GetDatabaseKey()
	return err
}

// RotateKey rotates the database key (for future implementation)
// This would allow users to change their database password
func RotateKey(newKey string) error {
	if newKey == "" {
		return fmt.Errorf("new key cannot be empty")
	}

	keyFile := models.GetDatabaseKeyPath()
	return writeKeyFile(keyFile, newKey)
}
