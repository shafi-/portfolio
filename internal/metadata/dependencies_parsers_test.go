package metadata

import "testing"

// These are white-box unit tests for the per-ecosystem ManifestParser types.
// They call Parse(content) directly — no filesystem — to confirm each parser is
// focused and individually testable now that it is a standalone type rather than
// a private function behind a central switch.

func TestNpmParser_Scope(t *testing.T) {
	deps, err := npmParser{}.Parse([]byte(`{
		"dependencies": {"react": "^18.0.0"},
		"devDependencies": {"jest": "^29.0.0"}
	}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	scope := make(map[string]string, len(deps))
	for _, d := range deps {
		scope[d.Name] = d.Scope
	}
	if scope["react"] != "prod" {
		t.Errorf("react scope: got %q, want prod", scope["react"])
	}
	if scope["jest"] != "dev" {
		t.Errorf("jest scope: got %q, want dev", scope["jest"])
	}
}

func TestNpmParser_Malformed(t *testing.T) {
	// Malformed JSON is tolerated (returns nil, nil) so the dispatcher can skip
	// the file rather than aborting the whole scan.
	deps, err := npmParser{}.Parse([]byte("not json"))
	if err != nil {
		t.Errorf("malformed manifest should be tolerated, got err: %v", err)
	}
	if deps != nil {
		t.Errorf("expected nil deps for malformed manifest, got %d", len(deps))
	}
}

func TestGoModParser_BlockAndSingle(t *testing.T) {
	deps, err := goModParser{}.Parse([]byte(`module example
go 1.21
require github.com/spf13/cobra v1.8.0
require (
	github.com/gin-gonic/gin v1.9.0
)
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	names := make(map[string]bool, len(deps))
	for _, d := range deps {
		if d.Manager != "go_mod" {
			t.Errorf("manager: got %q, want go_mod", d.Manager)
		}
		names[d.Name] = true
	}
	for _, want := range []string{"github.com/spf13/cobra", "github.com/gin-gonic/gin"} {
		if !names[want] {
			t.Errorf("missing dependency %q (block+single require handling); got %v", want, names)
		}
	}
}

func TestMavenParser_GroupArtifact(t *testing.T) {
	deps, err := mavenParser{}.Parse([]byte(`
<project>
  <dependencies>
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter</artifactId>
    </dependency>
  </dependencies>
</project>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dep, got %d", len(deps))
	}
	want := "org.springframework.boot:spring-boot-starter"
	if deps[0].Name != want {
		t.Errorf("name: got %q, want %q", deps[0].Name, want)
	}
	if deps[0].Manager != "maven" {
		t.Errorf("manager: got %q, want maven", deps[0].Manager)
	}
}

func TestCargoParser_DepsSection(t *testing.T) {
	// The parser must stop at the next TOML section, so [dev-dependencies] is
	// excluded and only [dependencies] entries are captured.
	deps, err := cargoParser{}.Parse([]byte(`
[package]
name = "demo"
[dependencies]
serde = "1.0"
tokio = { version = "1" }
[dev-dependencies]
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	names := make(map[string]bool, len(deps))
	for _, d := range deps {
		names[d.Name] = true
	}
	if !names["serde"] || !names["tokio"] {
		t.Errorf("expected serde + tokio from [dependencies], got %v", names)
	}
}
