package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Repositories []Repository `json:"repositories"`
}

type Repository struct {
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	URL            string      `json:"url"`
	CurrentVersion string      `json:"currentVersion"`
	Versioning     *Versioning `json:"versioning,omitempty"`
	Auth           *Auth       `json:"auth,omitempty"`
}

type Versioning struct {
	Scheme       string `json:"scheme,omitempty"`
	IgnorePrefix string `json:"ignorePrefix,omitempty"`
}

type Auth struct {
	Type        string `json:"type"`
	EnvVariable string `json:"envVariable"`
}

func LoadConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set default values
	for i := range config.Repositories {
		if config.Repositories[i].Versioning == nil {
			config.Repositories[i].Versioning = &Versioning{
				Scheme: "semver",
			}
		}
		if config.Repositories[i].Versioning.Scheme == "" {
			config.Repositories[i].Versioning.Scheme = "semver"
		}
	}

	return &config, nil
}

func (c *Config) FindRepository(name string) *Repository {
	for i := range c.Repositories {
		if c.Repositories[i].Name == name {
			return &c.Repositories[i]
		}
	}
	return nil
}