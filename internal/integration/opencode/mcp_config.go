package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	schemaURL    = "https://opencode.ai/config.json"
	mcpEntryName = "portfolio"
)

// opencodeServer is one entry under the config's "mcp" object. OpenCode's
// local-stdio server shape is:
//
//	mcp.<name> = { "type": "local", "command": ["<binary>", "mcp"], "enabled": true }
type opencodeServer struct {
	Type    string   `json:"type"`
	Command []string `json:"command"`
	Enabled bool     `json:"enabled"`
}

// readConfigMap reads opencode.json into a top-level map of raw JSON values.
// Using json.RawMessage preserves every unrelated key (theme, permissions,
// other MCP servers, ...) across our read-modify-write. A missing file yields
// an empty map; a malformed file yields an error so we never clobber it.
func readConfigMap(path string) (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]json.RawMessage{}, nil
		}
		return nil, err
	}
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return m, nil
}

// writeConfigMap atomically writes the config map, creating its directory.
// The temp-file-then-rename guarantees the config is never left half-written.
func writeConfigMap(path string, m map[string]json.RawMessage) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// readMCPServers parses the "mcp" object from the config, returning an empty
// (non-nil) map when absent or empty. A malformed mcp object is treated as
// empty rather than fatal, so a corrupt sibling entry never blocks install.
func readMCPServers(m map[string]json.RawMessage) map[string]opencodeServer {
	servers := map[string]opencodeServer{}
	if raw, ok := m["mcp"]; ok {
		_ = json.Unmarshal(raw, &servers)
	}
	return servers
}

// ensureMCPConfig registers the Portfolio MCP server in opencode.json. It is an
// idempotent read-merge-write: existing $schema, unrelated top-level keys, and
// other MCP servers are all preserved.
func (o *OpenCodeIntegration) ensureMCPConfig() error {
	path := o.config.ConfigPath
	m, err := readConfigMap(path)
	if err != nil {
		return err
	}

	// Preserve an existing $schema; set one if absent so the file is self-describing.
	if _, ok := m["$schema"]; !ok {
		schemaRaw, _ := json.Marshal(schemaURL)
		m["$schema"] = schemaRaw
	}

	servers := readMCPServers(m)
	servers[mcpEntryName] = opencodeServer{
		Type:    "local",
		Command: []string{o.config.BinaryPath, "mcp"},
		Enabled: true,
	}
	mcpRaw, _ := json.Marshal(servers)
	m["mcp"] = mcpRaw

	return writeConfigMap(path, m)
}

// removeMCPConfig removes the Portfolio entry from opencode.json. If the "mcp"
// object is left empty it is dropped entirely; unrelated keys are preserved.
// Missing file or entry is a no-op.
func (o *OpenCodeIntegration) removeMCPConfig() error {
	path := o.config.ConfigPath
	m, err := readConfigMap(path)
	if err != nil {
		return err
	}

	servers := readMCPServers(m)
	if _, exists := servers[mcpEntryName]; !exists {
		return nil
	}
	delete(servers, mcpEntryName)

	if len(servers) == 0 {
		delete(m, "mcp")
	} else {
		mcpRaw, _ := json.Marshal(servers)
		m["mcp"] = mcpRaw
	}

	return writeConfigMap(path, m)
}

// isMCPRegistered reports whether a Portfolio MCP entry exists in opencode.json.
func (o *OpenCodeIntegration) isMCPRegistered() bool {
	m, err := readConfigMap(o.config.ConfigPath)
	if err != nil {
		return false
	}
	servers := readMCPServers(m)
	_, ok := servers[mcpEntryName]
	return ok
}
