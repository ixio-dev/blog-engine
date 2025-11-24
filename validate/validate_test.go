package validate

import (
	"os"
	"path/filepath"
	"testing"
	"tech-blog/config"
)

func TestValidateConfig(t *testing.T) {
	// Test valid config
	validConfig := &config.Config{
		Site: config.SiteConfig{
			Title:   "Test Blog",
			BaseURL: "https://example.com",
			Author:  "Test Author",
		},
	}

	if err := Config(validConfig); err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}

	// Test config with missing title
	invalidConfig := &config.Config{
		Site: config.SiteConfig{
			Title:   "", // Missing title
			BaseURL: "https://example.com",
			Author:  "Test Author",
		},
	}

	if err := Config(invalidConfig); err == nil {
		t.Error("Expected error for missing title, got none")
	}

	// Test config with missing baseUrl
	invalidConfig2 := &config.Config{
		Site: config.SiteConfig{
			Title:   "Test Blog",
			BaseURL: "", // Missing baseUrl
			Author:  "Test Author",
		},
	}

	if err := Config(invalidConfig2); err == nil {
		t.Error("Expected error for missing baseUrl, got none")
	}

	// Test config with missing author
	invalidConfig3 := &config.Config{
		Site: config.SiteConfig{
			Title:   "Test Blog",
			BaseURL: "https://example.com",
			Author:  "", // Missing author
		},
	}

	if err := Config(invalidConfig3); err == nil {
		t.Error("Expected error for missing author, got none")
	}
}

func TestValidateContent(t *testing.T) {
	tempDir := t.TempDir()

	// Create a valid test post
	postPath := filepath.Join(tempDir, "test-post.md")
	content := `---
title: "Test Post"
date: "2024-01-15"
tags: ["test"]
draft: false
---

# Test Post

This is a test post content.
`

	err := os.WriteFile(postPath, []byte(content), 0644)
	if err != nil {
		t.Fatal("Failed to create test post")
	}

	if err := Content(tempDir); err != nil {
		t.Errorf("Expected valid content, got error: %v", err)
	}
}

func TestValidateContentWithInvalidPost(t *testing.T) {
	tempDir := t.TempDir()

	// Create an invalid test post (missing title)
	postPath := filepath.Join(tempDir, "test-post.md")
	content := `---
title: ""  # Empty title
date: "2024-01-15"
tags: ["test"]
draft: false
---

# Test Post

This is a test post content.
`

	err := os.WriteFile(postPath, []byte(content), 0644)
	if err != nil {
		t.Fatal("Failed to create test post")
	}

	if err := Content(tempDir); err == nil {
		t.Error("Expected error for invalid content, got none")
	}
}

func TestValidateNonExistentContentDir(t *testing.T) {
	if err := Content("/non/existent/directory"); err == nil {
		t.Error("Expected error for non-existent directory, got none")
	}
}