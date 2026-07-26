package metadata

import "testing"

// These are white-box unit tests for the per-ecosystem ManifestParser types.
// They call Parse(content) directly — no filesystem — to confirm each parser is
// focused and individually testable now that it is a standalone type rather than
// a private function behind a central switch. They also assert that the declared
// version is decomposed into value (version) and constraint kind (version_type).

func TestNpmParser_Scope(t *testing.T) {
	deps, err := npmParser{}.Parse([]byte(`{
		"dependencies": {"react": "^18.0.0"},
		"devDependencies": {"jest": "^29.0.0"}
	}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	scope := make(map[string]string, len(deps))
	ver := make(map[string]string, len(deps))
	kind := make(map[string]string, len(deps))
	for _, d := range deps {
		scope[d.Name] = d.Scope
		ver[d.Name] = d.Version
		kind[d.Name] = d.VersionType
	}
	if scope["react"] != "prod" {
		t.Errorf("react scope: got %q, want prod", scope["react"])
	}
	if scope["jest"] != "dev" {
		t.Errorf("jest scope: got %q, want dev", scope["jest"])
	}
	if ver["react"] != "18.0.0" || kind["react"] != "^" {
		t.Errorf("react version/type: got %q / %q, want 18.0.0 / ^", ver["react"], kind["react"])
	}
	if ver["jest"] != "29.0.0" || kind["jest"] != "^" {
		t.Errorf("jest version/type: got %q / %q, want 29.0.0 / ^", ver["jest"], kind["jest"])
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
	ver := make(map[string]string, len(deps))
	kind := make(map[string]string, len(deps))
	for _, d := range deps {
		if d.Manager != "go_mod" {
			t.Errorf("manager: got %q, want go_mod", d.Manager)
		}
		ver[d.Name] = d.Version
		kind[d.Name] = d.VersionType
	}
	if ver["github.com/spf13/cobra"] != "v1.8.0" || kind["github.com/spf13/cobra"] != "exact" {
		t.Errorf("cobra version/type: got %q / %q, want v1.8.0 / exact", ver["github.com/spf13/cobra"], kind["github.com/spf13/cobra"])
	}
	if ver["github.com/gin-gonic/gin"] != "v1.9.0" || kind["github.com/gin-gonic/gin"] != "exact" {
		t.Errorf("gin version/type: got %q / %q, want v1.9.0 / exact", ver["github.com/gin-gonic/gin"], kind["github.com/gin-gonic/gin"])
	}
}

func TestMavenParser_GroupArtifact(t *testing.T) {
	deps, err := mavenParser{}.Parse([]byte(`
<project>
  <dependencies>
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter</artifactId>
      <version>3.2.0</version>
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
	if deps[0].Version != "3.2.0" || deps[0].VersionType != "exact" {
		t.Errorf("version/type: got %q / %q, want 3.2.0 / exact", deps[0].Version, deps[0].VersionType)
	}
}

func TestCargoParser_DepsSection(t *testing.T) {
	// The parser must stop at the next TOML section, so [dev-dependencies] is
	// excluded and only [dependencies] entries are captured. Both the string
	// form ("1.0") and the inline-table form ({ version = "1" }) are captured.
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
	ver := make(map[string]string, len(deps))
	kind := make(map[string]string, len(deps))
	for _, d := range deps {
		ver[d.Name] = d.Version
		kind[d.Name] = d.VersionType
	}
	if ver["serde"] != "1.0" || kind["serde"] != "exact" {
		t.Errorf("serde version/type: got %q / %q, want 1.0 / exact", ver["serde"], kind["serde"])
	}
	if ver["tokio"] != "1" || kind["tokio"] != "exact" {
		t.Errorf("tokio (table form) version/type: got %q / %q, want 1 / exact", ver["tokio"], kind["tokio"])
	}
}

func TestPipRequirementsParser(t *testing.T) {
	// Comments, "-r"-style flags, and PEP 508 extras ("psycopg2[binary]") are
	// handled: extras are stripped to the bare distribution name. Version is
	// captured after "==" with kind "=="; a bare name (flask) has no version.
	deps, err := pipRequirementsParser{}.Parse([]byte("django==4.0\n" +
		"# a comment\n" +
		"psycopg2[binary]==2.9\n" +
		"-r other.txt\n" +
		"flask\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	ver := make(map[string]string, len(deps))
	kind := make(map[string]string, len(deps))
	names := make(map[string]bool, len(deps))
	for _, d := range deps {
		if d.Manager != "pip" {
			t.Errorf("manager: got %q, want pip", d.Manager)
		}
		ver[d.Name] = d.Version
		kind[d.Name] = d.VersionType
		names[d.Name] = true
	}
	for _, want := range []string{"django", "psycopg2", "flask"} {
		if !names[want] {
			t.Errorf("missing %q; got %v", want, names)
		}
	}
	if names["psycopg2[binary]"] {
		t.Errorf("extras marker should be stripped from psycopg2")
	}
	if len(deps) != 3 {
		t.Errorf("expected 3 deps (comment/-r skipped, extras stripped), got %d: %v", len(deps), names)
	}
	if ver["django"] != "4.0" || kind["django"] != "==" {
		t.Errorf("django version/type: got %q / %q, want 4.0 / ==", ver["django"], kind["django"])
	}
	if ver["flask"] != "" || kind["flask"] != "" {
		t.Errorf("flask version/type: got %q / %q, want empty (no version declared)", ver["flask"], kind["flask"])
	}
}

func TestPipPyprojectParser(t *testing.T) {
	// Only the [tool.poetry.dependencies] (or [project.dependencies]) map is
	// captured; "python" is excluded and parsing stops at the next section, so
	// the dev-dependencies block below is ignored.
	deps, err := pipPyprojectParser{}.Parse([]byte(`[tool.poetry.dependencies]
python = "^3.11"
django = "^4.0"
requests = "^2.31"

[tool.poetry.dev-dependencies]
pytest = "^7.0"
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	ver := make(map[string]string, len(deps))
	kind := make(map[string]string, len(deps))
	names := make(map[string]bool, len(deps))
	for _, d := range deps {
		if d.Manager != "pip" {
			t.Errorf("manager: got %q, want pip", d.Manager)
		}
		ver[d.Name] = d.Version
		kind[d.Name] = d.VersionType
		names[d.Name] = true
	}
	for _, want := range []string{"django", "requests"} {
		if !names[want] {
			t.Errorf("missing %q; got %v", want, names)
		}
	}
	if names["python"] {
		t.Errorf("python should be excluded")
	}
	if names["pytest"] {
		t.Errorf("dev-dependencies section should not be captured (parser breaks at next [)")
	}
	if ver["django"] != "4.0" || kind["django"] != "^" {
		t.Errorf("django version/type: got %q / %q, want 4.0 / ^", ver["django"], kind["django"])
	}
	if ver["requests"] != "2.31" || kind["requests"] != "^" {
		t.Errorf("requests version/type: got %q / %q, want 2.31 / ^", ver["requests"], kind["requests"])
	}
}

func TestBundlerParser(t *testing.T) {
	// Only lines starting with "gem " are parsed; source/ruby/comment lines
	// are ignored. The version token after the name is captured when present
	// (rails "~> 7.0"); a gem with no version (sidekiq) has empty version/type.
	deps, err := bundlerParser{}.Parse([]byte(`source "https://rubygems.org"
gem "rails", "~> 7.0"
gem "sidekiq"
ruby "3.2.0"
# gem "commented"
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	ver := make(map[string]string, len(deps))
	kind := make(map[string]string, len(deps))
	names := make(map[string]bool, len(deps))
	for _, d := range deps {
		if d.Manager != "bundler" {
			t.Errorf("manager: got %q, want bundler", d.Manager)
		}
		ver[d.Name] = d.Version
		kind[d.Name] = d.VersionType
		names[d.Name] = true
	}
	for _, want := range []string{"rails", "sidekiq"} {
		if !names[want] {
			t.Errorf("missing %q; got %v", want, names)
		}
	}
	if len(deps) != 2 {
		t.Errorf("expected 2 gems, got %d: %v", len(deps), names)
	}
	if ver["rails"] != "7.0" || kind["rails"] != "~>" {
		t.Errorf("rails version/type: got %q / %q, want 7.0 / ~>", ver["rails"], kind["rails"])
	}
	if ver["sidekiq"] != "" || kind["sidekiq"] != "" {
		t.Errorf("sidekiq version/type: got %q / %q, want empty (no version declared)", ver["sidekiq"], kind["sidekiq"])
	}
}

func TestGradleParser(t *testing.T) {
	// Only implementation/api/compile configurations are captured; configurations
	// like testImplementation are intentionally not matched. The final ':' segment
	// of the coordinate is the version.
	deps, err := gradleParser{}.Parse([]byte(`plugins { id 'java' }
dependencies {
    implementation 'com.google.guava:guava:32.0'
    api 'org.slf4j:slf4j-api:2.0'
    compile 'commons-io:commons-io:2.11'
    testImplementation 'junit:junit:4.13'
}
`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	ver := make(map[string]string, len(deps))
	kind := make(map[string]string, len(deps))
	names := make(map[string]bool, len(deps))
	for _, d := range deps {
		if d.Manager != "gradle" {
			t.Errorf("manager: got %q, want gradle", d.Manager)
		}
		ver[d.Name] = d.Version
		kind[d.Name] = d.VersionType
		names[d.Name] = true
	}
	for _, want := range []string{
		"com.google.guava:guava:32.0",
		"org.slf4j:slf4j-api:2.0",
		"commons-io:commons-io:2.11",
	} {
		if !names[want] {
			t.Errorf("missing %q; got %v", want, names)
		}
	}
	if names["junit:junit:4.13"] {
		t.Errorf("testImplementation should not be captured (only implementation/api/compile)")
	}
	if ver["com.google.guava:guava:32.0"] != "32.0" || kind["com.google.guava:guava:32.0"] != "exact" {
		t.Errorf("guava version/type: got %q / %q, want 32.0 / exact", ver["com.google.guava:guava:32.0"], kind["com.google.guava:guava:32.0"])
	}
	if ver["org.slf4j:slf4j-api:2.0"] != "2.0" || kind["org.slf4j:slf4j-api:2.0"] != "exact" {
		t.Errorf("slf4j version/type: got %q / %q, want 2.0 / exact", ver["org.slf4j:slf4j-api:2.0"], kind["org.slf4j:slf4j-api:2.0"])
	}
}

func TestParseVersionSpec(t *testing.T) {
	cases := []struct {
		raw         string
		wantVersion string
		wantKind    string
	}{
		{"^4.0.0", "4.0.0", "^"},
		{"~4.2.1", "4.2.1", "~"},
		{"~> 7.0", "7.0", "~>"},
		{">=1.0", "1.0", ">="},
		{"<=2.0", "2.0", "<="},
		{"<3.0", "3.0", "<"},
		{">5.0", "5.0", ">"},
		{"==2.28.0", "2.28.0", "=="},
		{"=1.5", "1.5", "="},
		{"4.0.0", "4.0.0", "exact"},           // bare pin
		{"v1.2.3", "v1.2.3", "exact"},         // go.mod keeps the v prefix
		{">=1.0,<2.0", ">=1.0,<2.0", "range"}, // compound range
		{"1.0 - 2.0", "1.0 - 2.0", "range"},   // hyphen range
		{"*", "*", "any"},
		{"latest", "latest", "any"},
		{"1.x", "1.x", "any"}, // x-range
		{"3.+", "3.+", "any"}, // gradle dynamic
		{"", "", ""},          // unknown
		{"   ", "", ""},       // whitespace only
	}
	for _, c := range cases {
		gotVersion, gotKind := parseVersionSpec(c.raw)
		if gotVersion != c.wantVersion || gotKind != c.wantKind {
			t.Errorf("parseVersionSpec(%q): got %q / %q, want %q / %q",
				c.raw, gotVersion, gotKind, c.wantVersion, c.wantKind)
		}
	}
}
