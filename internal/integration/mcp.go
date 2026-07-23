package integration

import (
	"context"
	"fmt"
)

type MCPClient interface {
	Health(ctx context.Context) error
	ListTools(ctx context.Context) ([]string, error)
	RegisterTools(ctx context.Context, tools []ToolDef) error
}

type ToolDef struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type FakeMCPClient struct {
	Healthy     bool
	Tools       []string
	HealthError error
	ListError   error
	RegisterErr error
}

func NewFakeMCPClient() *FakeMCPClient {
	return &FakeMCPClient{
		Healthy: true,
		Tools: []string{
			"health",
			"discoverProjects",
			"listProjects",
			"getProject",
			"searchProjects",
			"searchDocumentation",
			"getAnalysis",
			"storeAnalysis",
			"listProjectsNeedingAnalysis",
			"getConfiguration",
			"updateConfiguration",
			"listRelationships",
		},
	}
}

func (f *FakeMCPClient) Health(ctx context.Context) error {
	if f.HealthError != nil {
		return f.HealthError
	}
	if !f.Healthy {
		return fmt.Errorf("MCP server not responding")
	}
	return nil
}

func (f *FakeMCPClient) ListTools(ctx context.Context) ([]string, error) {
	if f.ListError != nil {
		return nil, f.ListError
	}
	return f.Tools, nil
}

func (f *FakeMCPClient) RegisterTools(ctx context.Context, tools []ToolDef) error {
	if f.RegisterErr != nil {
		return f.RegisterErr
	}
	for _, tool := range tools {
		f.Tools = append(f.Tools, tool.Name)
	}
	return nil
}
