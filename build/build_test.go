package build

import (
	"os"
	"path/filepath"
	"tech-blog/config"
	"testing"
)

func TestBuilder(t *testing.T) {
	// Create a temporary working directory
	workingDir := t.TempDir()

	// Change to the working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get original directory")
	}
	defer os.Chdir(origDir)

	err = os.Chdir(workingDir)
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

	// Create basic templates
	baseContent := `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>{{.Title}} — {{.Site.Title}}</title>
  <meta name="description" content="{{.Description}}" />
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
  <time>{{.Post.DatePretty}}</time>
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
  <li><a href="{{$.BaseURL}}/tags/{{$tag}}/">{{$tag}} ({{len $posts}})</a></li>
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

	// Create legal template
	legalContent := `{{define "content"}}
<div>
  <h1>Legal Information</h1>
  <div>{{.LegalContent}}</div>
</div>
{{end}}`

	err = os.WriteFile("templates/legal.gohtml", []byte(legalContent), 0644)
	if err != nil {
		t.Fatal("Failed to create legal template")
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

	// Create the config file in the working directory
	cfgFile := `site:
  title: "Test Blog"
  baseUrl: "https://example.com"
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

	err = os.WriteFile("config.yaml", []byte(cfgFile), 0644)
	if err != nil {
		t.Fatal("Failed to create config file")
	}

	// Load the config from the working directory - importing config package function
	testCfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test creating a builder
	builder, err := New(testCfg)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	// Test building
	err = builder.Build(false, false) // Don't include drafts, not preview mode
	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	// Check if output files were created
	outputIndex := filepath.Join(testCfg.Build.OutputDir, "index.html")
	if _, err := os.Stat(outputIndex); os.IsNotExist(err) {
		t.Error("Expected index.html to be created")
	}

	// Check if posts were created
	postDir := filepath.Join(testCfg.Build.OutputDir, "posts", "test-post")
	postIndex := filepath.Join(postDir, "index.html")
	if _, err := os.Stat(postIndex); os.IsNotExist(err) {
		t.Error("Expected post index.html to be created")
	}

	// Check if assets were copied
	assetsDir := filepath.Join(testCfg.Build.OutputDir, "assets")
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		t.Error("Expected assets directory to be created")
	}
}

func TestBuilderWithMissingTemplates(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a config
	cfg := &config.Config{
		Site: config.SiteConfig{
			Title:   "Test Blog",
			BaseURL: "https://example.com",
			Author:  "Test Author",
		},
		Build: config.BuildConfig{
			OutputDir: filepath.Join(tempDir, "dist"),
		},
		Server: config.ServerConfig{
			Port: 4000,
		},
	}

	// Test creating a builder with missing templates
	_, err := New(cfg)
	if err == nil {
		t.Error("Expected error when templates are missing")
	}
}
