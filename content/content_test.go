package content

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParsePost(t *testing.T) {
	tempDir := t.TempDir()
	contentPath := filepath.Join(tempDir, "test-post.md")

	content := `---
title: "Test Post"
date: "2024-01-15"
tags: ["test", "example"]
draft: false
slug: "test-post"
---

# Test Post

This is a test post content.
`

	err := os.WriteFile(contentPath, []byte(content), 0644)
	if err != nil {
		t.Fatal("Failed to create test post")
	}

	post, err := ParsePost(contentPath)
	if err != nil {
		t.Fatalf("Failed to parse post: %v", err)
	}

	if post.Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", post.Title)
	}

	expectedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if post.Date != expectedDate {
		t.Errorf("Expected date %v, got %v", expectedDate, post.Date)
	}

	if len(post.Tags) != 2 || post.Tags[0] != "test" || post.Tags[1] != "example" {
		t.Errorf("Expected tags [test, example], got %v", post.Tags)
	}

	if post.Draft != false {
		t.Errorf("Expected draft to be false, got %v", post.Draft)
	}

	if post.Slug != "test-post" {
		t.Errorf("Expected slug 'test-post', got '%s'", post.Slug)
	}
}

func TestParsePostWithInvalidDate(t *testing.T) {
	tempDir := t.TempDir()
	contentPath := filepath.Join(tempDir, "test-post.md");

	// Invalid date format - this should cause an error in ParsePost
	content := `---
title: "Test Post"
date: "invalid-date-format"
---

# Test Post

This is a test post content.
`

	err := os.WriteFile(contentPath, []byte(content), 0644);
	if err != nil {
		t.Fatal("Failed to create test post")
	}

	// This should fail because date format is invalid
	_, err = ParsePost(contentPath)
	if err == nil {
		t.Error("Expected error when date format is invalid")
	}
}

func TestLoadPosts(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a test post
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

	site, err := LoadPosts(tempDir, false) // Don't include drafts
	if err != nil {
		t.Fatalf("Failed to load posts: %v", err)
	}

	if len(site.Posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(site.Posts))
	}

	if site.Posts[0].Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", site.Posts[0].Title)
	}

	// Check tag indexing
	if len(site.PostsByTag) != 1 {
		t.Errorf("Expected 1 tag index, got %d", len(site.PostsByTag))
	}

	if posts, exists := site.PostsByTag["test"]; !exists || len(posts) != 1 {
		t.Errorf("Expected 1 post for tag 'test', got %d", len(posts))
	}
}

func TestExtractImageSources(t *testing.T) {
	tt := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name:     "no images",
			html:     "<p>Just some text</p>",
			expected: []string{},
		},
		{
			name:     "single image",
			html:     `<p>Image: <img src="/images/test.jpg" alt="test" /></p>`,
			expected: []string{"/images/test.jpg"},
		},
		{
			name:     "multiple images",
			html:     `<p><img src="/images/test1.jpg" alt="test1" /><img src="relative/image.png" /></p>`,
			expected: []string{"/images/test1.jpg", "relative/image.png"},
		},
		{
			name:     "external images are ignored",
			html:     `<img src="http://example.com/image.jpg" /><img src="/images/local.jpg" />`,
			expected: []string{"/images/local.jpg"},
		},
		{
			name:     "mixed image types",
			html:     `<p>External: <img src="https://example.com/img.jpg" /> Internal: <img src="/assets/img.png" /> Relative: <img src="./local.jpg" /></p>`,
			expected: []string{"/assets/img.png", "./local.jpg"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ExtractImageSources(template.HTML(tc.html))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d images, got %d", len(tc.expected), len(result))
				t.Errorf("Expected: %v, got: %v", tc.expected, result)
				return
			}

			for i, exp := range tc.expected {
				if result[i] != exp {
					t.Errorf("Expected image %d to be '%s', got '%s'", i, exp, result[i])
				}
			}
		})
	}
}

func TestParsePostWithImages(t *testing.T) {
	tempDir := t.TempDir()
	contentPath := filepath.Join(tempDir, "test-post.md")

	content := `---
title: "Test Post with Images"
date: "2024-01-15"
tags: ["test"]
draft: false
---

# Test Post

This post includes an image:

![Alt text](/images/test.jpg)

And another image:
![Another image](./relative/path.png)

External images should not be included:
![External](https://example.com/image.jpg)
`

	err := os.WriteFile(contentPath, []byte(content), 0644)
	if err != nil {
		t.Fatal("Failed to create test post")
	}

	// Create the image files that should be referenced
	imgDir := filepath.Join(tempDir, "images")
	err = os.MkdirAll(imgDir, 0755)
	if err != nil {
		t.Fatal("Failed to create image directory")
	}

	err = os.WriteFile(filepath.Join(imgDir, "test.jpg"), []byte("fake image content"), 0644)
	if err != nil {
		t.Fatal("Failed to create image file")
	}

	relDir := filepath.Join(tempDir, "relative", "path")
	err = os.MkdirAll(relDir, 0755)
	if err != nil {
		t.Fatal("Failed to create relative path directory")
	}

	err = os.WriteFile(filepath.Join(relDir, "png"), []byte("fake png content"), 0644)
	if err != nil {
		t.Fatal("Failed to create relative image file")
	}

	post, err := ParsePost(contentPath)
	if err != nil {
		t.Fatalf("Failed to parse post with images: %v", err)
	}

	// Check that the post has the correct image sources
	expectedImages := []string{"/images/test.jpg", "./relative/path.png"}
	if len(post.ImageSources) != len(expectedImages) {
		t.Errorf("Expected %d image sources, got %d", len(expectedImages), len(post.ImageSources))
		t.Errorf("Expected: %v, got: %v", expectedImages, post.ImageSources)
	}

	for i, exp := range expectedImages {
		if post.ImageSources[i] != exp {
			t.Errorf("Expected image source %d to be '%s', got '%s'", i, exp, post.ImageSources[i])
		}
	}
}

func TestSyntaxHighlighting(t *testing.T) {
	tempDir := t.TempDir()
	contentPath := filepath.Join(tempDir, "test-post.md")

	content := `---
title: "Test Post with Code"
date: "2024-01-15"
tags: ["test"]
draft: false
---

# Test Post

Here's some code:

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, world!")
}
` + "```" + `
`

	err := os.WriteFile(contentPath, []byte(content), 0644)
	if err != nil {
		t.Fatal("Failed to create test post with code")
	}

	post, err := ParsePost(contentPath)
	if err != nil {
		t.Fatalf("Failed to parse post with code: %v", err)
	}

	// Check that the HTML contains code block
	htmlContent := string(post.HTML)

	// The output shows that syntax highlighting is working - we see <pre> with styled spans
	// Check for the presence of syntax highlighting elements in the output
	if !strings.Contains(htmlContent, "<pre") || !strings.Contains(htmlContent, "style=") {
		t.Errorf("Expected code block with syntax highlighting in HTML output, got: %s", htmlContent)
	} else {
		t.Logf("Code block with syntax highlighting found in HTML output")
	}
}

// Helper function to check if HTML contains basic code block structure (used by TestExtractImageSources)
func containsCodeBlock(html string) bool {
	return strings.Contains(html, "<pre>") && strings.Contains(html, "<code")
}