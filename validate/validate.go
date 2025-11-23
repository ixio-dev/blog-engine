package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tech-blog/config"
	"tech-blog/content"
)

func Config(cfg *config.Config) error {
	if cfg.Site.Title == "" {
		return fmt.Errorf("site.title is required")
	}
	if cfg.Site.BaseURL == "" {
		return fmt.Errorf("site.baseUrl is required")
	}
	if cfg.Site.Author == "" {
		return fmt.Errorf("site.author is required")
	}
	return nil
}

func Content(contentDir string) error {
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return fmt.Errorf("content directory does not exist: %s", contentDir)
	}

	return filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		return validatePost(path)
	})
}

func validatePost(path string) error {
	post, err := content.ParsePost(path)
	if err != nil {
		return fmt.Errorf("failed to parse post %s: %w", path, err)
	}

	if post.Title == "" {
		return fmt.Errorf("post %s: title is required", path)
	}

	if post.Date.IsZero() {
		return fmt.Errorf("post %s: date is required", path)
	}

	if post.Date.After(time.Now()) {
		return fmt.Errorf("post %s: date cannot be in the future", path)
	}

	// Validate slug format
	if post.Slug != "" {
		if strings.Contains(post.Slug, " ") {
			return fmt.Errorf("post %s: slug cannot contain spaces", path)
		}
		if strings.ContainsAny(post.Slug, "/\\") {
			return fmt.Errorf("post %s: slug cannot contain path separators", path)
		}
	}

	return nil
}