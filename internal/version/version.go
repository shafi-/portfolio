// Package version holds the build version metadata for the Portfolio Engine.
// It is the single source of truth consumed by the CLI version flag, the startup
// log, and the MCP server handshake, so all three always agree on one version.
package version

var (
	// version is the engine release version.
	version = "v0.3.11"
	// commit is the VCS commit the binary was built from ("dev" when unset).
	commit = "dev"
	// date is the build timestamp ("unknown" when unset).
	date = "unknown"
)

// Version returns the engine release version (e.g. "0.2.0").
func Version() string { return version }

// Commit returns the VCS commit the binary was built from.
func Commit() string { return commit }

// Date returns the build timestamp.
func Date() string { return date }

// Full returns a human-readable "version (commit: …, built: …)" string suitable
// for the CLI --version flag.
func Full() string {
	return version + " (commit: " + commit + ", built: " + date + ")"
}
