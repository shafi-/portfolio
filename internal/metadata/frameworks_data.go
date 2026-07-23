package metadata

type FrameworkMarker struct {
	Name      string
	Manifest  string
	Pattern   string
	Ecosystem string
}

var defaultFrameworkMarkers = []FrameworkMarker{
	{Name: "React", Manifest: "package.json", Pattern: "react", Ecosystem: "npm"},
	{Name: "Vue", Manifest: "package.json", Pattern: "vue", Ecosystem: "npm"},
	{Name: "Angular", Manifest: "package.json", Pattern: "@angular/core", Ecosystem: "npm"},
	{Name: "Svelte", Manifest: "package.json", Pattern: "svelte", Ecosystem: "npm"},
	{Name: "Next.js", Manifest: "package.json", Pattern: "next", Ecosystem: "npm"},
	{Name: "Nuxt", Manifest: "package.json", Pattern: "nuxt", Ecosystem: "npm"},
	{Name: "Gatsby", Manifest: "package.json", Pattern: "gatsby", Ecosystem: "npm"},
	{Name: "Express", Manifest: "package.json", Pattern: "express", Ecosystem: "npm"},
	{Name: "NestJS", Manifest: "package.json", Pattern: "@nestjs/core", Ecosystem: "npm"},
	{Name: "Fastify", Manifest: "package.json", Pattern: "fastify", Ecosystem: "npm"},
	{Name: "Vite", Manifest: "package.json", Pattern: "vite", Ecosystem: "npm"},
	{Name: "Jest", Manifest: "package.json", Pattern: "jest", Ecosystem: "npm"},
	{Name: "Vitest", Manifest: "package.json", Pattern: "vitest", Ecosystem: "npm"},

	{Name: "Gin", Manifest: "go.mod", Pattern: "github.com/gin-gonic/gin", Ecosystem: "go_mod"},
	{Name: "Echo", Manifest: "go.mod", Pattern: "github.com/labstack/echo", Ecosystem: "go_mod"},
	{Name: "Fiber", Manifest: "go.mod", Pattern: "github.com/gofiber/fiber", Ecosystem: "go_mod"},
	{Name: "Chi", Manifest: "go.mod", Pattern: "github.com/go-chi/chi", Ecosystem: "go_mod"},
	{Name: "Cobra", Manifest: "go.mod", Pattern: "github.com/spf13/cobra", Ecosystem: "go_mod"},

	{Name: "Django", Manifest: "requirements.txt", Pattern: "django", Ecosystem: "pip"},
	{Name: "Flask", Manifest: "requirements.txt", Pattern: "flask", Ecosystem: "pip"},
	{Name: "FastAPI", Manifest: "requirements.txt", Pattern: "fastapi", Ecosystem: "pip"},
	{Name: "Django", Manifest: "pyproject.toml", Pattern: "django", Ecosystem: "pip"},
	{Name: "Flask", Manifest: "pyproject.toml", Pattern: "flask", Ecosystem: "pip"},
	{Name: "FastAPI", Manifest: "pyproject.toml", Pattern: "fastapi", Ecosystem: "pip"},

	{Name: "Rocket", Manifest: "Cargo.toml", Pattern: "rocket", Ecosystem: "cargo"},
	{Name: "Actix", Manifest: "Cargo.toml", Pattern: "actix-web", Ecosystem: "cargo"},
	{Name: "Axum", Manifest: "Cargo.toml", Pattern: "axum", Ecosystem: "cargo"},

	{Name: "Rails", Manifest: "Gemfile", Pattern: "rails", Ecosystem: "bundler"},
	{Name: "Sinatra", Manifest: "Gemfile", Pattern: "sinatra", Ecosystem: "bundler"},

	{Name: "Spring Boot", Manifest: "pom.xml", Pattern: "spring-boot", Ecosystem: "maven"},
	{Name: "Spring Boot", Manifest: "build.gradle", Pattern: "spring-boot", Ecosystem: "gradle"},
	{Name: "Quarkus", Manifest: "pom.xml", Pattern: "quarkus", Ecosystem: "maven"},
}

func DefaultFrameworkMarkers() []FrameworkMarker {
	markers := make([]FrameworkMarker, len(defaultFrameworkMarkers))
	copy(markers, defaultFrameworkMarkers)
	return markers
}
