package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Site   SiteConfig   `yaml:"site"`
	Build  BuildConfig  `yaml:"build"`
	Server ServerConfig `yaml:"server"`
	Deploy DeployConfig `yaml:"deploy"`
}

type SiteConfig struct {
	Title    string `yaml:"title"`
	BaseURL  string `yaml:"baseUrl"`
	Author   string `yaml:"author"`
}

type BuildConfig struct {
	OutputDir string `yaml:"outputDir"`
	Drafts    bool   `yaml:"drafts"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DeployConfig struct {
	Host string `yaml:"host"`
	Path string `yaml:"path"`
}

func Load() (*Config, error) {
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config.yaml not found")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
	}

	// Set defaults
	if config.Build.OutputDir == "" {
		config.Build.OutputDir = "./dist"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 4000
	}

	return &config, nil
}

func (c *Config) ContentDir() string {
	return "content"
}

func (c *Config) TemplatesDir() string {
	return "templates"
}

func (c *Config) StaticDir() string {
	return "static"
}