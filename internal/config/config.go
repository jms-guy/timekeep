package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	WakaTime WakaTimeConfig `json:"wakatime"`
}

type WakaTimeConfig struct {
	Enabled       bool   `json:"enabled"`
	APIKey        string `json:"api_key,omitempty"`
	CLIPath       string `json:"cli_path,omitempty"`
	GlobalProject string `json:"global_project,omitempty"`
}

const defaultConfig = `{
  "wakatime": {
    "enabled": false,
	"api_key": "",
	"cli_path": "",
	"global_project": ""
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
