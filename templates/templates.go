package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"

	"tech-blog/config"
	"tech-blog/content"
)

type TemplateEngine struct {
	templates map[string]*template.Template
	config    *config.Config
}

type TemplateData struct {
	Site             *config.SiteConfig
	BaseURL          string
	Title            string
	Description      string
	Content          template.HTML
	Posts            []*content.Post
	Post             *content.Post
	Tag              string
	Tags             map[string][]*content.Post
	PrevPost         *PostLink
	NextPost         *PostLink
	Pagination       template.HTML
	ExtraHead        template.HTML
	LiveReloadScript template.HTML
	DraftsVisible    bool
	AssetsVersion    string
	Year             string
	LegalContent     template.HTML
}

type PostLink struct {
	Title string
	URL   string
}

func New(config *config.Config) (*TemplateEngine, error) {
	engine := &TemplateEngine{
		config: config,
	}

	// Load templates
	err := engine.loadTemplates()
	if err != nil {
		return nil, err
	}

	return engine, nil
}

func (t *TemplateEngine) loadTemplates() error {
	// Load base template
	baseContent, err := os.ReadFile("templates/base.gohtml")
	if err != nil {
		return fmt.Errorf("failed to read base template: %w", err)
	}

	baseTmpl := template.New("base").Funcs(templateFuncs)
	_, err = baseTmpl.Parse(string(baseContent))
	if err != nil {
		return fmt.Errorf("failed to parse base template: %w", err)
	}

	// Create separate templates for each page type
	t.templates = make(map[string]*template.Template)
	pageTypes := []string{"index", "post", "tag", "tags", "legal"}

	for _, pageType := range pageTypes {
		pageTmpl, err := baseTmpl.Clone()
		if err != nil {
			return fmt.Errorf("failed to clone template for %s: %w", pageType, err)
		}

		contentFile := fmt.Sprintf("templates/%s.gohtml", pageType)
		content, err := os.ReadFile(contentFile)
		if err != nil {
			return fmt.Errorf("failed to read %s template: %w", pageType, err)
		}

		_, err = pageTmpl.New(pageType).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse %s template: %w", pageType, err)
		}

		t.templates[pageType] = pageTmpl
	}

	return nil
}

// ReloadTemplates reloads all templates from disk
func (t *TemplateEngine) ReloadTemplates() error {
	return t.loadTemplates()
}

func (t *TemplateEngine) RenderIndex(site *content.Site, data *TemplateData) ([]byte, error) {
	data.Posts = site.Posts
	return t.renderWithContent("index", data)
}

func (t *TemplateEngine) RenderPost(post *content.Post, prev, next *content.Post, data *TemplateData) ([]byte, error) {
	data.Post = post
	if prev != nil {
		data.PrevPost = &PostLink{Title: prev.Title, URL: prev.URL}
	}
	if next != nil {
		data.NextPost = &PostLink{Title: next.Title, URL: next.URL}
	}
	return t.renderWithContent("post", data)
}

func (t *TemplateEngine) RenderTag(tag string, posts []*content.Post, data *TemplateData) ([]byte, error) {
	data.Tag = tag
	data.Posts = posts
	return t.renderWithContent("tag", data)
}

func (t *TemplateEngine) RenderTagsIndex(tags map[string][]*content.Post, data *TemplateData) ([]byte, error) {
	data.Tags = tags
	return t.renderWithContent("tags", data)
}

func (t *TemplateEngine) renderWithContent(contentTemplate string, data *TemplateData) ([]byte, error) {
	tmpl, ok := t.templates[contentTemplate]
	if !ok {
		return nil, fmt.Errorf("template %s not found", contentTemplate)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderTemplate renders a specific template (like legal, about, etc.) with the base layout
func (t *TemplateEngine) RenderTemplate(templateName string, data *TemplateData) ([]byte, error) {
	return t.renderWithContent(templateName, data)
}

// Template functions
var templateFuncs = template.FuncMap{
	"joinStrings": func(slice []string, sep string) string {
		return strings.Join(slice, sep)
	},
	"slugify": func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
	},
	"isLast": func(slice []string, index int) bool {
		return index == len(slice)-1
	},
}

func init() {
	// This would be called when creating templates
	// template.New("").Funcs(templateFuncs)
}
