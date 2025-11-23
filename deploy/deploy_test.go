package deploy

import (
	"os"
	"testing"
	"tech-blog/config"
)

func TestDeployer(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Change to the working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get original directory")
	}
	defer os.Chdir(origDir)
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal("Failed to change to working directory")
	}
	
	// Create the expected directories
	err = os.MkdirAll("content", 0755)
	if err != nil {
		t.Fatal("Failed to create content directory")
	}
	
	err = os.MkdirAll("templates", 0755)
	if err != nil {
		t.Fatal("Failed to create templates directory")
	}
	
	err = os.MkdirAll("static", 0755)
	if err != nil {
		t.Fatal("Failed to create static directory")
	}
	
	// Create basic templates
	baseContent := `<!doctype html>
<html lang="en">
<head>
  <title>{{.Title}}</title>
</head>
<body>
  {{template "content" .}}
</body>
</html>`

	indexContent := `{{define "content"}}
<h1>{{.Title}}</h1>
<ul>
{{range .Posts}}
  <li><a href="{{.URL}}">{{.Title}}</a></li>
{{end}}
{{end}}`

	postContentTemplate := `{{define "content"}}
<article>
  <h1>{{.Post.Title}}</h1>
  <div>{{.Post.HTML}}</div>
</article>
{{end}}`

	tagContent := `{{define "content"}}
<h1>Tag: {{.Tag}}</h1>
<ul>
{{range .Posts}}
  <li><a href="{{.URL}}">{{.Title}}</a></li>
{{end}}
{{end}}`

	tagsContent := `{{define "content"}}
<h1>Tags</h1>
<ul>
{{range $tag, $posts := .Tags}}
  <li><a href="/tags/{{$tag}}/">{{$tag}} ({{len $posts}})</a></li>
{{end}}
{{end}}`

	err = os.WriteFile("templates/base.gohtml", []byte(baseContent), 0644)
	if err != nil {
		t.Fatal("Failed to create base template")
	}

	err = os.WriteFile("templates/index.gohtml", []byte(indexContent), 0644)
	if err != nil {
		t.Fatal("Failed to create index template")
	}

	err = os.WriteFile("templates/post.gohtml", []byte(postContentTemplate), 0644)
	if err != nil {
		t.Fatal("Failed to create post template")
	}

	err = os.WriteFile("templates/tag.gohtml", []byte(tagContent), 0644)
	if err != nil {
		t.Fatal("Failed to create tag template")
	}

	err = os.WriteFile("templates/tags.gohtml", []byte(tagsContent), 0644)
	if err != nil {
		t.Fatal("Failed to create tags template")
	}

	// Create a simple CSS file in static
	err = os.MkdirAll("static/css", 0755)
	if err != nil {
		t.Fatal("Failed to create static/css directory")
	}
	
	err = os.WriteFile("static/css/main.css", []byte("body { color: black; }"), 0644)
	if err != nil {
		t.Fatal("Failed to create CSS file")
	}

	// Create a test post
	postContent := `---
title: "Test Post"
date: "2024-01-15"
tags: ["test"]
draft: false
---

# Test Post

This is a test post content.
`

	err = os.WriteFile("content/test-post.md", []byte(postContent), 0644)
	if err != nil {
		t.Fatal("Failed to create test post")
	}

	// Create config
	cfg := &config.Config{
		Site: config.SiteConfig{
			Title:   "Test Blog",
			BaseURL: "https://example.com",
			Author:  "Test Author",
		},
		Build: config.BuildConfig{
			OutputDir: "./dist",
		},
		Server: config.ServerConfig{
			Port: 4000,
		},
		Deploy: config.DeployConfig{
			Host: "user@localhost",  // This will fail in real deployment but not during creation
			Path: "/var/www/test",
		},
	}

	// Test creating a deployer
	deployer := New(cfg)
	if deployer == nil {
		t.Fatal("Failed to create deployer")
	}

	// Test that the config is stored properly
	if deployer.config.Site.Title != "Test Blog" {
		t.Errorf("Expected title 'Test Blog', got '%s'", deployer.config.Site.Title)
	}
}

func TestDeployerMissingConfig(t *testing.T) {
	// Create config with missing deploy values
	cfg := &config.Config{
		Site: config.SiteConfig{
			Title:   "Test Blog",
			BaseURL: "https://example.com",
			Author:  "Test Author",
		},
		Build: config.BuildConfig{
			OutputDir: "./dist",
		},
		Server: config.ServerConfig{
			Port: 4000,
		},
		Deploy: config.DeployConfig{
			Host: "",  // Missing host
			Path: "",  // Missing path
		},
	}

	// Test creating a deployer
	deployer := New(cfg)

	// Test deployment with missing config - should fail appropriately
	err := deployer.Deploy()
	if err == nil {
		t.Error("Expected error when deploy config is missing")
	}
}