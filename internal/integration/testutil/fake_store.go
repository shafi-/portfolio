package testutil

import (
	"context"
	"fmt"

	"project-dash/internal/integration"
)

type FakeStore struct {
	integrations map[string]integration.IntegrationMeta
	SaveError    error
	GetError     error
	ListError    error
	DeleteError  error
}

func NewFakeStore() *FakeStore {
	return &FakeStore{
		integrations: make(map[string]integration.IntegrationMeta),
	}
}

func (f *FakeStore) SaveIntegration(ctx context.Context, meta integration.IntegrationMeta) error {
	if f.SaveError != nil {
		return f.SaveError
	}
	f.integrations[meta.Name] = meta
	return nil
}

func (f *FakeStore) GetIntegration(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
	if f.GetError != nil {
		return nil, f.GetError
	}
	if meta, ok := f.integrations[name]; ok {
		return &meta, nil
	}
	return nil, integration.ErrNotFound
}

func (f *FakeStore) ListIntegrations(ctx context.Context) ([]integration.IntegrationMeta, error) {
	if f.ListError != nil {
		return nil, f.ListError
	}
	result := make([]integration.IntegrationMeta, 0, len(f.integrations))
	for _, meta := range f.integrations {
		result = append(result, meta)
	}
	return result, nil
}

func (f *FakeStore) DeleteIntegration(ctx context.Context, name string) error {
	if f.DeleteError != nil {
		return f.DeleteError
	}
	if _, ok := f.integrations[name]; !ok {
		return integration.ErrNotFound
	}
	delete(f.integrations, name)
	return nil
}

func (f *FakeStore) WithIntegration(name, agentType, version string) *FakeStore {
	f.integrations[name] = integration.IntegrationMeta{
		Name:      name,
		AgentType: agentType,
		Version:   version,
		Status:    integration.StatusInstalled,
	}
	return f
}

func (f *FakeStore) WithStoreError(err error) *FakeStore {
	f.SaveError = err
	f.GetError = err
	f.ListError = err
	f.DeleteError = err
	return f
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

func (f *FakeMCPClient) RegisterTools(ctx context.Context, tools []integration.ToolDef) error {
	if f.RegisterErr != nil {
		return f.RegisterErr
	}
	for _, tool := range tools {
		f.Tools = append(f.Tools, tool.Name)
	}
	return nil
}

func (f *FakeMCPClient) WithUnhealthy() *FakeMCPClient {
	f.Healthy = false
	f.HealthError = fmt.Errorf("MCP server not responding")
	return f
}

func (f *FakeMCPClient) WithError(err error) *FakeMCPClient {
	f.HealthError = err
	f.ListError = err
	f.RegisterErr = err
	return f
}

func (f *FakeMCPClient) WithToolListError(err error) *FakeMCPClient {
	f.ListError = err
	return f
}
