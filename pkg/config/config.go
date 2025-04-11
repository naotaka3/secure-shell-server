package config

// Default execution timeout in seconds.
const DefaultExecutionTimeout = 30

// ShellConfig holds configuration for the secure shell command execution.
type ShellConfig struct {
	// AllowedCommands is a map of command names that are allowed to be executed
	AllowedCommands map[string]bool

	// RestrictedEnv is a map of environment variables to be set
	RestrictedEnv map[string]string

	// WorkingDir is the working directory for command execution
	WorkingDir string

	// MaxExecutionTime is the maximum execution time in seconds
	MaxExecutionTime int
}

// NewDefaultConfig returns a default configuration.
func NewDefaultConfig() *ShellConfig {
	return &ShellConfig{
		AllowedCommands: map[string]bool{
			"ls":   true,
			"echo": true,
			"cat":  true,
		},
		RestrictedEnv: map[string]string{
			"PATH": "/usr/bin:/bin",
		},
		WorkingDir:       "",
		MaxExecutionTime: DefaultExecutionTimeout,
	}
}

// AddAllowedCommand adds a command to the allowed commands list.
func (c *ShellConfig) AddAllowedCommand(cmd string) {
	c.AllowedCommands[cmd] = true
}

// RemoveAllowedCommand removes a command from the allowed commands list.
func (c *ShellConfig) RemoveAllowedCommand(cmd string) {
	delete(c.AllowedCommands, cmd)
}

// SetEnv sets an environment variable.
func (c *ShellConfig) SetEnv(key, value string) {
	c.RestrictedEnv[key] = value
}

// IsCommandAllowed checks if a command is allowed.
func (c *ShellConfig) IsCommandAllowed(cmd string) bool {
	return c.AllowedCommands[cmd]
}
