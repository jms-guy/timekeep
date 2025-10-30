package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	WakaTime     WakaTimeConfig `json:"wakatime"`                // WakaTime integration variables
	Wakapi       WakapiConfig   `json:"wakapi"`                  // Wakapi integration variables
	PollInterval string         `json:"poll_interval,omitempty"` // Linux - monitor polling interval, default 1s
	PollGrace    int            `json:"poll_grace,omitempty"`    // Linux - number representing the grace period granted to PIDs accidently missed by polling, default 3
}

type WakaTimeConfig struct {
	Enabled       bool   `json:"enabled"`                  // WakaTime integration enabling value
	APIKey        string `json:"api_key,omitempty"`        // WakaTime account API key
	CLIPath       string `json:"cli_path,omitempty"`       // wakatime-cli path on local machine
	GlobalProject string `json:"global_project,omitempty"` // Default project to associate all tracked programs with
}

type WakapiConfig struct {
	Enabled       bool   `json:"enabled"`                  // Wakapi integration enabling value
	Server        string `json:"server,omitempty"`         // Wakapi server address
	APIKey        string `json:"api_key,omitempty"`        // Wakapi API key
	GlobalProject string `json:"global_project,omitempty"` // Default project to associate all tracked programs with
}

const defaultConfig = `{
  "wakatime": {
    "enabled": false
  },
  "wakapi": {
  "enabled": false
  }
}`

func Load() (*Config, error) {
	configFile, err := getConfigLocation()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(configFile), 0o750); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		err := os.WriteFile(configFile, []byte(defaultConfig), 0o600)
		if err != nil {
			return nil, fmt.Errorf("error generating default config: %w", err)
		}
	}

	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return &config, nil
}

func (c *Config) Save() error {
	configFile, err := getConfigLocation()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
