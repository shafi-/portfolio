package integration

type Status string

const (
	StatusInstalled    Status = "installed"
	StatusNotInstalled Status = "not_installed"
)

type IntegrationMeta struct {
	Name             string `json:"name"`
	AgentType        string `json:"agent_type"`
	Version          string `json:"version"`
	InstallPath      string `json:"install_path"`
	Status           Status `json:"status"`
	MinEngineVersion string `json:"min_engine_version"`
	InstalledAt      string `json:"installed_at"`
	UpdatedAt        string `json:"updated_at"`
}

type InstallOptions struct {
	InstallPath string
	Force       bool
}

type InstallResult struct {
	Meta     IntegrationMeta
	Warnings []string
}

type UpgradeOptions struct {
	TargetVersion string
	EngineVersion string
}

type UpgradeResult struct {
	PreviousVersion string
	NewVersion      string
	RolledBack      bool
	NoOp            bool
}

type ValidationResult struct {
	Passed bool              `json:"passed"`
	Checks []ValidationCheck `json:"checks"`
}

type ValidationCheck struct {
	Name         string `json:"name"`
	Passed       bool   `json:"passed"`
	Message      string `json:"message,omitempty"`
	Remediation  string `json:"remediation,omitempty"`
	SelfHealable bool   `json:"self_healable,omitempty"`
}
