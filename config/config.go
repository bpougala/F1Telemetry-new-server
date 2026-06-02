package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed *.json
var configFS embed.FS

// Config holds the environment-specific SignalR endpoints. The base URLs include
// the scheme (e.g. https/wss for prod, http/ws for local) so the same connection
// code can target the real F1 service or a local mock server.
type Config struct {
	NegotiateBaseURL string `json:"negotiateBaseURL"`
	ConnectBaseURL   string `json:"connectBaseURL"`
}

// Load reads the config for the environment named by APP_ENV (defaulting to "prod").
// Config files are embedded into the binary so they are always available regardless
// of the working directory.
func Load() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "prod"
	}
	data, err := configFS.ReadFile(env + ".json")
	if err != nil {
		return nil, fmt.Errorf("loading config for env %q: %w", env, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config for env %q: %w", env, err)
	}
	return &cfg, nil
}
