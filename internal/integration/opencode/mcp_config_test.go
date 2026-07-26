package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeFile is a small helper for seeding test config files.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// parseTop reads the config file into a generic map for assertions.
func parseTop(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return m
}

func TestEnsureMCPConfigCreatesEntry(t *testing.T) {
	o := newTestIntegration(t)

	if err := o.ensureMCPConfig(); err != nil {
		t.Fatalf("ensureMCPConfig: %v", err)
	}

	top := parseTop(t, o.config.ConfigPath)

	// $schema present.
	var schema string
	if raw, ok := top["$schema"]; ok {
		_ = json.Unmarshal(raw, &schema)
	} else {
		t.Fatal("$schema missing from config")
	}
	if schema != schemaURL {
		t.Errorf("$schema = %q, want %q", schema, schemaURL)
	}

	// mcp.portfolio entry with correct shape.
	servers := readMCPServers(top)
	srv, ok := servers[mcpEntryName]
	if !ok {
		t.Fatal("mcp.portfolio entry missing")
	}
	if srv.Type != "local" {
		t.Errorf("type = %q, want local", srv.Type)
	}
	if !srv.Enabled {
		t.Error("enabled = false, want true")
	}
	if len(srv.Command) != 2 || srv.Command[0] != o.config.BinaryPath || srv.Command[1] != "mcp" {
		t.Errorf("command = %v, want [%s mcp]", srv.Command, o.config.BinaryPath)
	}
}

func TestEnsureMCPConfigPreservesUnrelatedKeys(t *testing.T) {
	o := newTestIntegration(t)

	// Pre-existing config with unrelated top-level keys, a different $schema,
	// and a sibling MCP server that must survive.
	seed := `{
  "$schema": "https://example.com/kept.json",
  "theme": "dark",
  "mcp": {
    "other-server": { "type": "local", "command": ["foo"], "enabled": false }
  }
}`
	writeFile(t, o.config.ConfigPath, seed)

	if err := o.ensureMCPConfig(); err != nil {
		t.Fatalf("ensureMCPConfig: %v", err)
	}

	top := parseTop(t, o.config.ConfigPath)

	// Existing $schema must be preserved, not overwritten.
	var schema string
	_ = json.Unmarshal(top["$schema"], &schema)
	if schema != "https://example.com/kept.json" {
		t.Errorf("$schema overwritten: got %q", schema)
	}

	// Unrelated top-level key preserved.
	var theme string
	if raw, ok := top["theme"]; ok {
		_ = json.Unmarshal(raw, &theme)
	} else {
		t.Fatal("unrelated key 'theme' was dropped")
	}
	if theme != "dark" {
		t.Errorf("theme = %q, want dark", theme)
	}

	// Sibling server preserved + portfolio added.
	servers := readMCPServers(top)
	if _, ok := servers["other-server"]; !ok {
		t.Error("sibling MCP server 'other-server' was dropped")
	}
	if _, ok := servers[mcpEntryName]; !ok {
		t.Error("portfolio entry was not added")
	}
}

func TestEnsureMCPConfigIdempotent(t *testing.T) {
	o := newTestIntegration(t)

	if err := o.ensureMCPConfig(); err != nil {
		t.Fatalf("first ensure: %v", err)
	}
	first, _ := os.ReadFile(o.config.ConfigPath)

	if err := o.ensureMCPConfig(); err != nil {
		t.Fatalf("second ensure: %v", err)
	}
	second, _ := os.ReadFile(o.config.ConfigPath)

	// No sibling servers, so the file should be byte-identical after re-run
	// (deterministic, stable key ordering from json.Marshal on a map).
	if string(first) != string(second) {
		t.Errorf("ensureMCPConfig is not idempotent\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	// And only one portfolio entry exists.
	servers := readMCPServers(parseTop(t, o.config.ConfigPath))
	count := 0
	for name := range servers {
		if name == mcpEntryName {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 portfolio entry, got %d", count)
	}
}

func TestIsMCPRegistered(t *testing.T) {
	o := newTestIntegration(t)

	if o.isMCPRegistered() {
		t.Error("isMCPRegistered true before any install")
	}

	if err := o.ensureMCPConfig(); err != nil {
		t.Fatalf("ensureMCPConfig: %v", err)
	}
	if !o.isMCPRegistered() {
		t.Error("isMCPRegistered false after install")
	}
}

func TestRemoveMCPConfigPreservesOthers(t *testing.T) {
	o := newTestIntegration(t)

	if err := o.ensureMCPConfig(); err != nil {
		t.Fatalf("ensureMCPConfig: %v", err)
	}

	// Add a sibling server alongside portfolio.
	top, _ := readConfigMap(o.config.ConfigPath)
	servers := readMCPServers(top)
	servers["other-server"] = opencodeServer{Type: "local", Command: []string{"foo"}, Enabled: true}
	raw, _ := json.Marshal(servers)
	top["mcp"] = raw
	if err := writeConfigMap(o.config.ConfigPath, top); err != nil {
		t.Fatalf("writeConfigMap: %v", err)
	}

	if err := o.removeMCPConfig(); err != nil {
		t.Fatalf("removeMCPConfig: %v", err)
	}

	servers = readMCPServers(parseTop(t, o.config.ConfigPath))
	if _, ok := servers[mcpEntryName]; ok {
		t.Error("portfolio entry still present after remove")
	}
	if _, ok := servers["other-server"]; !ok {
		t.Error("sibling server 'other-server' was removed")
	}
}

func TestRemoveMCPConfigDropsEmptyMCPMap(t *testing.T) {
	o := newTestIntegration(t)

	if err := o.ensureMCPConfig(); err != nil {
		t.Fatalf("ensureMCPConfig: %v", err)
	}
	if err := o.removeMCPConfig(); err != nil {
		t.Fatalf("removeMCPConfig: %v", err)
	}

	top := parseTop(t, o.config.ConfigPath)
	if _, ok := top["mcp"]; ok {
		t.Error("empty 'mcp' object was left behind after removing the only entry")
	}
}

func TestRemoveMCPConfigMissingIsNoOp(t *testing.T) {
	o := newTestIntegration(t)

	// No file at all.
	if err := o.removeMCPConfig(); err != nil {
		t.Fatalf("removeMCPConfig on missing file: %v", err)
	}

	// File with no portfolio entry.
	writeFile(t, o.config.ConfigPath, `{"theme": "dark"}`)
	if err := o.removeMCPConfig(); err != nil {
		t.Fatalf("removeMCPConfig with no entry: %v", err)
	}
	if _, err := os.Stat(o.config.ConfigPath); err != nil {
		t.Fatalf("config file should be untouched: %v", err)
	}
}

func TestReadConfigMapMalformedErrors(t *testing.T) {
	o := newTestIntegration(t)
	writeFile(t, o.config.ConfigPath, `{not valid json`)

	if _, err := readConfigMap(o.config.ConfigPath); err == nil {
		t.Fatal("expected error parsing malformed config, got nil")
	}

	// And ensureMCPConfig must refuse to clobber a malformed file.
	if err := o.ensureMCPConfig(); err == nil {
		t.Fatal("ensureMCPConfig must not overwrite a malformed config")
	}
}
