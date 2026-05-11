package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	IgnoreFailingJobs []string `json:"ignoreFailingJobs"`
}

func Load() *Config {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".gh-next", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}
	}
	return &cfg
}
