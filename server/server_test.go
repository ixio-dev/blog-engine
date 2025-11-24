package server

import (
	"os"
	"tech-blog/config"
	"testing"
)

func TestServer(t *testing.T) {
	// Basic server functionality tests would go here
	// For now, just verify the package is valid
	if true != true {
		t.Error("Basic test failed")
	}
}

func TestServerDraftToggle(t *testing.T) {
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

	// Create config directory structure
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

	// Create a draft post
	draftContent := `---
title: "Draft Post"
date: "2024-01-15"
tags: ["draft"]
draft: true
---

# Draft Post

This is a draft post that should only appear in preview when drafts are included.
`
	err = os.WriteFile("content/draft-post.md", []byte(draftContent), 0644)
	if err != nil {
		t.Fatal("Failed to create draft post")
	}

	// Create a regular post
	regularContent := `---
title: "Regular Post"
date: "2024-01-16"
tags: ["test"]
draft: false
---

# Regular Post

This is a regular post that should always appear.
`
	err = os.WriteFile("content/regular-post.md", []byte(regularContent), 0644)
	if err != nil {
		t.Fatal("Failed to create regular post")
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

	postContent := `{{define "content"}}
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

	legalContent := `{{define "content"}}
<div>
  <h1>Legal Information</h1>
  <div>{{.LegalContent}}</div>
</div>
{{end}}`

	err = os.WriteFile("templates/base.gohtml", []byte(baseContent), 0644)
	if err != nil {
		t.Fatal("Failed to create base template")
	}

	err = os.WriteFile("templates/index.gohtml", []byte(indexContent), 0644)
	if err != nil {
		t.Fatal("Failed to create index template")
	}

	err = os.WriteFile("templates/post.gohtml", []byte(postContent), 0644)
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

	err = os.WriteFile("templates/legal.gohtml", []byte(legalContent), 0644)
	if err != nil {
		t.Fatal("Failed to create legal template")
	}

	// Create basic config
	cfgFile := `site:
  title: "Test Blog"
  baseUrl: "https://example.com"
  author: "Test Author"

build:
  outputDir: "./dist"

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

	// Load the config
	testCfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create the server
	server := New(testCfg)

	// Test initial state - drafts should be included by default in preview mode
	if !server.includeDrafts {
		t.Error("Expected drafts to be included by default in preview server")
	}

	// Test toggle functionality
	server.toggleDrafts()
	if server.includeDrafts {
		t.Error("Expected drafts to be excluded after toggle")
	}

	// Toggle back
	server.toggleDrafts()
	if !server.includeDrafts {
		t.Error("Expected drafts to be included after second toggle")
	}
}