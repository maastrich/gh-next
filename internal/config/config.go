package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	IgnoreFailingJobs []string `json:"ignoreFailingJobs"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gh-next", "config.json")
}

func Load() *Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return &Config{}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}
	}
	return &cfg
}

func (c *Config) Save() error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
