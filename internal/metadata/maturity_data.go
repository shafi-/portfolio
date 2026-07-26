package metadata

// MaturityArtifact describes a project-maturity signal detected by file
// presence. An artifact is "present" when any of its Paths exists at the
// project root (files via os.Stat; glob patterns via filepath.Glob). Each
// artifact contributes its Weight to the maturity score.
type MaturityArtifact struct {
	Paths  []string
	Kind   string
	Weight int
}

var defaultMaturityArtifacts = []MaturityArtifact{
	{Kind: "readme", Weight: 1, Paths: []string{"README", "README.md", "README.rst", "README.txt"}},
	{Kind: "license", Weight: 1, Paths: []string{"LICENSE", "LICENSE.md", "LICENSE.txt", "COPYING", "LICENSE-MIT", "LICENSE-APACHE"}},
	{Kind: "changelog", Weight: 1, Paths: []string{"CHANGELOG", "CHANGELOG.md", "CHANGES.md", "CHANGES"}},
	{Kind: "contributing", Weight: 1, Paths: []string{"CONTRIBUTING", "CONTRIBUTING.md", ".github/CONTRIBUTING.md"}},
	{Kind: "security", Weight: 1, Paths: []string{"SECURITY.md", ".github/SECURITY.md"}},
	{Kind: "env_example", Weight: 1, Paths: []string{".env.example", ".env.sample", ".env.template"}},
	{Kind: "dockerfile", Weight: 1, Paths: []string{"Dockerfile", "docker/Dockerfile", "containers/Dockerfile"}},
	{Kind: "docker_compose", Weight: 1, Paths: []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}},
	{Kind: "ci", Weight: 2, Paths: []string{
		".github/workflows/*.yml", ".github/workflows/*.yaml",
		".gitlab-ci.yml", ".circleci/config.yml", "azure-pipelines.yml",
	}},
	{Kind: "tests", Weight: 2, Paths: []string{
		"jest.config.js", "jest.config.ts", "jest.config.json", "jest.config.cjs", "jest.config.mjs",
		"vitest.config.ts", "vitest.config.js", "vitest.config.mjs",
		"playwright.config.ts", "playwright.config.js",
		"pytest.ini", "conftest.py", "tox.ini",
	}},
	{Kind: "linter", Weight: 1, Paths: []string{
		".eslintrc", ".eslintrc.js", ".eslintrc.json", ".eslintrc.yml", ".eslintrc.cjs",
		"eslint.config.js", "eslint.config.mjs",
		".prettierrc", ".prettierrc.js", ".prettierrc.json", "prettier.config.js",
		".golangci.yml", ".golangci.yaml",
		".rubocop.yml", ".stylelintrc",
	}},
	{Kind: "typescript", Weight: 1, Paths: []string{"tsconfig.json"}},
	{Kind: "docs", Weight: 2, Paths: []string{"docs", "doc", "documentation"}},
	{Kind: "makefile", Weight: 1, Paths: []string{"Makefile", "makefile", "GNUmakefile"}},
}

// DefaultMaturityArtifacts returns a defensive copy of the default maturity artifacts.
func DefaultMaturityArtifacts() []MaturityArtifact {
	out := make([]MaturityArtifact, len(defaultMaturityArtifacts))
	copy(out, defaultMaturityArtifacts)
	return out
}
