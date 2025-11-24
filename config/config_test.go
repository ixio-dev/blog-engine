package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempConfig := `site:
  title: "Test Blog"
  baseUrl: "https://test.com"
  author: "Test Author"

build:
  outputDir: "./dist"
  drafts: false

server:
  port: 4000

deploy:
  host: "user@test.com"
  path: "/var/www/test"
`

	// Write to the expected config.yaml file name
	err := os.WriteFile("config.yaml", []byte(tempConfig), 0644)
	if err != nil {
		t.Fatal("Failed to create config file")
	}
	defer os.Remove("config.yaml")

	// Test loading the config
	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the loaded values
	if config.Site.Title != "Test Blog" {
		t.Errorf("Expected title 'Test Blog', got '%s'", config.Site.Title)
	}

	if config.Site.BaseURL != "https://test.com" {
		t.Errorf("Expected baseUrl 'https://test.com', got '%s'", config.Site.BaseURL)
	}

	if config.Build.OutputDir != "./dist" {
		t.Errorf("Expected outputDir './dist', got '%s'", config.Build.OutputDir)
	}
}

func TestLoadMissingConfig(t *testing.T) {
	// Ensure config.yaml doesn't exist
	os.Remove("config.yaml") // Clean up if exists from other tests

	_, err := Load()
	if err == nil {
		t.Error("Expected error when config file is missing")
	}
}
