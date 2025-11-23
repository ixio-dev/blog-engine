package templates

import (
	"os"
	"testing"
	"tech-blog/config"
	"tech-blog/content"
)

func TestTemplateEngine(t *testing.T) {
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
	err = os.MkdirAll("templates", 0755)
	if err != nil {
		t.Fatal("Failed to create templates directory")
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

	// Create a config
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
		Deploy: config.DeployConfig{},
	}

	// Test creating a template engine
	engine, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// Test that templates are loaded
	if len(engine.templates) != 5 {
		t.Errorf("Expected 5 templates, got %d", len(engine.templates))
	}

	// Test rendering with mock data
	site := &content.Site{
		Posts: []*content.Post{
			{
				Title: "Test Post",
				Slug:  "test-post",
				URL:   "/posts/test-post/",
			},
		},
		PostsByTag: map[string][]*content.Post{
			"test": {
				{
					Title: "Test Post",
					Slug:  "test-post",
					URL:   "/posts/test-post/",
				},
			},
		},
	}

	// Test index rendering
	data := &TemplateData{
		Site:    &cfg.Site,
		BaseURL: cfg.Site.BaseURL,
		Title:   "Test Index",
	}

	_, err = engine.RenderIndex(site, data)
	if err != nil {
		t.Errorf("Failed to render index: %v", err)
	}

	// Test post rendering
	_, err = engine.RenderPost(site.Posts[0], nil, nil, data)
	if err != nil {
		t.Errorf("Failed to render post: %v", err)
	}

	// Test tag rendering
	_, err = engine.RenderTag("test", site.PostsByTag["test"], data)
	if err != nil {
		t.Errorf("Failed to render tag: %v", err)
	}

	// Test tags rendering
	_, err = engine.RenderTagsIndex(site.PostsByTag, data)
	if err != nil {
		t.Errorf("Failed to render tags index: %v", err)
	}
}