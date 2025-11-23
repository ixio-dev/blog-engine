package content

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	highlighting "github.com/yuin/goldmark-highlighting"
	xhtml "golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

type Post struct {
	Title       string    `yaml:"title"`
	Date        time.Time `yaml:"-"`
	DateStr     string    `yaml:"date"`
	Tags        []string  `yaml:"tags"`
	Draft       bool      `yaml:"draft"`
	Slug        string
	Content     string
	HTML        template.HTML
	Summary     string
	DatePretty  string
	URL         string
	ImageSources []string
}

type Site struct {
	Posts        []*Post
	PostsByTag   map[string][]*Post
	LegalContent template.HTML
}

func LoadPosts(contentDir string, includeDrafts bool) (*Site, error) {
	return LoadPostsAtBaseURL(contentDir, includeDrafts, "")
}

func LoadPostsAtBaseURL(contentDir string, includeDrafts bool, baseURL string) (*Site, error) {
	posts := []*Post{}

	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		post, err := ParsePost(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		if post.Draft && !includeDrafts {
			return nil
		}

		// Generate slug from filename if not specified
		if post.Slug == "" {
			base := strings.TrimSuffix(filepath.Base(path), ".md")
			post.Slug = slugify(base)
		}

		// Generate URL based on baseURL
		if baseURL != "" {
			// Ensure baseURL doesn't end with a slash to properly concatenate
			baseURLForPath := strings.TrimSuffix(baseURL, "/")
			post.URL = fmt.Sprintf("%s/posts/%s/", baseURLForPath, post.Slug)
		} else {
			post.URL = fmt.Sprintf("/posts/%s/", post.Slug)
		}
		post.DatePretty = post.Date.Format("January 2, 2006")

		posts = append(posts, post)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort posts by date (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	// Build tag index
	postsByTag := make(map[string][]*Post)
	for _, post := range posts {
		for _, tag := range post.Tags {
			postsByTag[tag] = append(postsByTag[tag], post)
		}
	}

	// Load legal notice content
	legalContent, err := LoadLegalContent()
	if err != nil {
		// If legal.md doesn't exist, use default placeholder content
		legalContent = "<p>Legal notice content will be displayed here.</p>"
	}

	return &Site{
		Posts:        posts,
		PostsByTag:   postsByTag,
		LegalContent: legalContent,
	}, nil
}

// LoadLegalContent loads and converts the legal notice markdown file to HTML
func LoadLegalContent() (template.HTML, error) {
	content, err := os.ReadFile("legal.md")
	if err != nil {
		return "", err
	}

	// Convert markdown to HTML with syntax highlighting
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return "", fmt.Errorf("failed to convert legal notice markdown: %w", err)
	}

	return template.HTML(buf.String()), nil
}

func ParsePost(path string) (*Post, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Split frontmatter and content
	parts := bytes.SplitN(content, []byte("---\n"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid frontmatter format")
	}

	var post Post
	if err := yaml.Unmarshal(parts[1], &post); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Parse date
	if post.DateStr != "" {
		parsed, err := time.Parse("2006-01-02", post.DateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}
		post.Date = parsed
	}

	// Convert markdown to HTML with syntax highlighting
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(parts[2], &buf); err != nil {
		return nil, fmt.Errorf("failed to convert markdown: %w", err)
	}

	post.Content = string(parts[2])
	post.HTML = template.HTML(buf.String())

	// Extract image sources from HTML
	imageSources, err := ExtractImageSources(post.HTML)
	if err != nil {
		return nil, fmt.Errorf("failed to extract image sources: %w", err)
	}
	post.ImageSources = imageSources

	// Generate summary (first paragraph)
	lines := strings.Split(strings.TrimSpace(post.Content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			post.Summary = line
			if len(post.Summary) > 200 {
				post.Summary = post.Summary[:200] + "..."
			}
			break
		}
	}

	return &post, nil
}

func slugify(s string) string {
	// Simple slugification - convert to lowercase and replace spaces with hyphens
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}

// ExtractImageSources extracts image sources from HTML content
func ExtractImageSources(htmlContent template.HTML) ([]string, error) {
	doc, err := xhtml.Parse(strings.NewReader(string(htmlContent)))
	if err != nil {
		return nil, err
	}

	var imageSources []string
	var f func(*xhtml.Node)
	f = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					// Only include relative paths (not absolute URLs)
					if !strings.HasPrefix(attr.Val, "http://") && !strings.HasPrefix(attr.Val, "https://") {
						imageSources = append(imageSources, attr.Val)
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return imageSources, nil
}