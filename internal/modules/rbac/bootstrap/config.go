package bootstrap

// RBACConfig represents the RBAC configuration structure
type RBACConfig struct {
	Permissions []PermissionConfig `yaml:"permissions"`
	Roles       []RoleConfig       `yaml:"roles"`
}

// PermissionConfig represents a permission in the config
type PermissionConfig struct {
	Code        string `yaml:"code"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Resource    string `yaml:"resource"`
	Action      string `yaml:"action"`
}

// RoleConfig represents a role in the config
type RoleConfig struct {
	Code        string   `yaml:"code"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	IsSystem    bool     `yaml:"is_system"`
	IsDefault   bool     `yaml:"is_default"`
	MaxUsers    *int     `yaml:"max_users"`
	Permissions []string `yaml:"permissions"` // Permission codes or wildcards
}
