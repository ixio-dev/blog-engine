package build

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tech-blog/config"
	"tech-blog/content"
	"tech-blog/templates"
)

type Builder struct {
	config *config.Config
	engine *templates.TemplateEngine
}

type SearchIndexEntry struct {
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Text    string   `json:"text"`
	Tags    []string `json:"tags"`
	URL     string   `json:"url"`
}

func New(config *config.Config) (*Builder, error) {
	engine, err := templates.New(config)
	if err != nil {
		return nil, err
	}

	return &Builder{
		config: config,
		engine: engine,
	}, nil
}

// FileCache keeps track of file modification times to enable incremental builds
type FileCache struct {
	FileModTimes map[string]time.Time `json:"fileModTimes"`
}

func (b *Builder) loadFileCache() (*FileCache, error) {
	cachePath := filepath.Join(b.config.Build.OutputDir, ".build_cache.json")

	cache := &FileCache{
		FileModTimes: make(map[string]time.Time),
	}

	// Try to load existing cache
	data, err := os.ReadFile(cachePath)
	if err != nil {
		// If cache doesn't exist, return empty cache
		if os.IsNotExist(err) {
			return cache, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cache); err != nil {
		return nil, fmt.Errorf("failed to parse build cache: %w", err)
	}

	return cache, nil
}

func (b *Builder) saveFileCache(cache *FileCache) error {
	cachePath := filepath.Join(b.config.Build.OutputDir, ".build_cache.json")

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal build cache: %w", err)
	}

	return os.WriteFile(cachePath, data, 0644)
}

func (b *Builder) shouldRebuildFile(filePath string, outputPath string, cache *FileCache) (bool, error) {
	// Get source file modification time
	srcInfo, err := os.Stat(filePath)
	if err != nil {
		return true, err // If source doesn't exist, we might need to clean up output
	}

	// Get destination file modification time
	dstInfo, err := os.Stat(outputPath)
	if err != nil {
		// If output doesn't exist, definitely rebuild
		if os.IsNotExist(err) {
			return true, nil
		}
		return true, err
	}

	// Check if source is newer than destination
	if srcInfo.ModTime().After(dstInfo.ModTime()) {
		return true, nil
	}

	// Check if source is newer than our cached time
	if srcInfo.ModTime().After(cache.FileModTimes[filePath]) {
		return true, nil
	}

	return false, nil
}

func (b *Builder) Build(includeDrafts bool, preview bool) error {
	// Override base URL for preview mode
	if preview {
		b.config.Site.BaseURL = fmt.Sprintf("http://localhost:%d", b.config.Server.Port)
	}

	// Load content with base URL
	site, err := content.LoadPostsAtBaseURL(b.config.ContentDir(), includeDrafts, b.config.Site.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to load posts: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(b.config.Build.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	// Load file cache for incremental builds
	cache, err := b.loadFileCache()
	if err != nil {
		// If cache loading fails, proceed with full rebuild
		cache = &FileCache{FileModTimes: make(map[string]time.Time)}
		fmt.Printf("Warning: could not load build cache, performing full rebuild: %v\n", err)
	}

	// Remember the previous modification times for comparison
	previousModTimes := make(map[string]time.Time)
	for k, v := range cache.FileModTimes {
		previousModTimes[k] = v
	}

	// Copy static files incrementally
	if err := b.copyStaticFilesIncremental(cache); err != nil {
		return fmt.Errorf("failed to copy static files: %w", err)
	}

	// Generate assets version
	assetsVersion := b.generateAssetsVersion()

	// Build pages (incrementally)
	if err := b.buildPagesIncremental(site, assetsVersion, includeDrafts, preview, cache); err != nil {
		return fmt.Errorf("failed to build pages: %w", err)
	}

	// Copy image assets referenced in posts incrementally
	if err := b.copyImageAssets(site, cache); err != nil {
		return fmt.Errorf("failed to copy image assets: %w", err)
	}

	// Generate search index
	if err := b.buildSearchIndex(site); err != nil {
		return fmt.Errorf("failed to build search index: %w", err)
	}

	// Inject live reload script for preview mode
	if preview {
		if err := b.injectLiveReloadScript(); err != nil {
			return fmt.Errorf("failed to inject live reload script: %w", err)
		}
	}

	// Update cache with current modification times
	for _, post := range site.Posts {
		cache.FileModTimes[filepath.Join(b.config.ContentDir(), post.Slug+".md")] = post.Date
	}

	// Save the updated cache
	if err := b.saveFileCache(cache); err != nil {
		fmt.Printf("Warning: failed to save build cache: %v\n", err)
	}

	fmt.Printf("Built %d posts\n", len(site.Posts))
	return nil
}

func (b *Builder) copyStaticFiles() error {
	return filepath.Walk(b.config.StaticDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(b.config.StaticDir(), path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(b.config.Build.OutputDir, "assets", relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		return b.copyFile(path, destPath)
	})
}

func (b *Builder) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (b *Builder) copyStaticFilesIncremental(cache *FileCache) error {
	return filepath.Walk(b.config.StaticDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(b.config.StaticDir(), path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(b.config.Build.OutputDir, "assets", relPath)

		// Check if we need to rebuild this file
		shouldRebuild, err := b.shouldRebuildFile(path, destPath, cache)
		if err != nil {
			return err
		}

		if !shouldRebuild {
			return nil // Skip this file, it's up to date
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		return b.copyFile(path, destPath)
	})
}

func (b *Builder) generateAssetsVersion() string {
	// Simple hash of current time for now - in production you'd hash the actual assets
	h := sha256.Sum256([]byte(time.Now().String()))
	return fmt.Sprintf("?v=%x", h[:8])
}

// buildPages with incremental support
func (b *Builder) buildPagesIncremental(site *content.Site, assetsVersion string, includeDrafts bool, preview bool, cache *FileCache) error {
	baseData := &templates.TemplateData{
		Site:          &b.config.Site,
		BaseURL:       b.config.Site.BaseURL,
		DraftsVisible: includeDrafts,
		AssetsVersion: assetsVersion,
		Year:          strconv.Itoa(time.Now().Year()),
	}

	if preview {
		baseData.LiveReloadScript = template.HTML(`<script defer src="/assets/js/live-reload.js"></script>`)
	}

	// Build index page
	indexData := *baseData
	indexData.Title = b.config.Site.Title
	indexPath := filepath.Join(b.config.Build.OutputDir, "index.html")

	// Check if index needs rebuild (depends on all posts)
	indexNeedsRebuild := true // For now, always rebuild index since it depends on all posts
	if indexNeedsRebuild {
		html, err := b.engine.RenderIndex(site, &indexData)
		if err != nil {
			return fmt.Errorf("failed to render index: %w", err)
		}
		if err := os.WriteFile(indexPath, html, 0644); err != nil {
			return err
		}
	}

	// Build post pages
	for i, post := range site.Posts {
		var prev, next *content.Post
		if i > 0 {
			prev = site.Posts[i-1]
		}
		if i < len(site.Posts)-1 {
			next = site.Posts[i+1]
		}

		postDir := filepath.Join(b.config.Build.OutputDir, "posts", post.Slug)
		postIndexPath := filepath.Join(postDir, "index.html")

		// Check if post needs rebuild based on source file modification
		sourcePath := filepath.Join(b.config.ContentDir(), post.Slug+".md")
		shouldRebuild, err := b.shouldRebuildFile(sourcePath, postIndexPath, cache)
		if err != nil {
			// If there's an error checking, rebuild to be safe
			shouldRebuild = true
		}

		if shouldRebuild {
			// Create post data and render
			postData := *baseData
			postData.Title = post.Title
			postData.Description = post.Summary

			html, err := b.engine.RenderPost(post, prev, next, &postData)
			if err != nil {
				return fmt.Errorf("failed to render post %s: %w", post.Slug, err)
			}

			if err := os.MkdirAll(postDir, 0755); err != nil {
				return err
			}

			if err := os.WriteFile(postIndexPath, html, 0644); err != nil {
				return err
			}
		}
	}

	// Build tags index page
	tagsDir := filepath.Join(b.config.Build.OutputDir, "tags")
	tagsIndexPath := filepath.Join(tagsDir, "index.html")

	if len(site.PostsByTag) > 0 {
		// Always rebuild tags index since it depends on all posts
		tagsData := *baseData
		tagsData.Title = "Tags"

		html, err := b.engine.RenderTagsIndex(site.PostsByTag, &tagsData)
		if err != nil {
			return fmt.Errorf("failed to render tags index: %w", err)
		}

		if err := os.MkdirAll(tagsDir, 0755); err != nil {
			return err
		}

		if err := os.WriteFile(tagsIndexPath, html, 0644); err != nil {
			return err
		}
	}

	// Build tag pages
	for tag, posts := range site.PostsByTag {
		tagDir := filepath.Join(b.config.Build.OutputDir, "tags", strings.ToLower(tag))
		tagIndexPath := filepath.Join(tagDir, "index.html")

		// Check if tag page needs rebuild
		// For now, tag pages depend on all posts with that tag, so rebuild if any post changed
		tagNeedsRebuild := false
		for _, post := range posts {
			sourcePath := filepath.Join(b.config.ContentDir(), post.Slug+".md")
			if cache.FileModTimes[sourcePath].IsZero() ||
				post.Date.After(cache.FileModTimes[sourcePath]) {
				tagNeedsRebuild = true
				break
			}
		}

		if tagNeedsRebuild {
			tagData := *baseData
			tagData.Title = fmt.Sprintf("Tag: %s", tag)

			html, err := b.engine.RenderTag(tag, posts, &tagData)
			if err != nil {
				return fmt.Errorf("failed to render tag %s: %w", tag, err)
			}

			if err := os.MkdirAll(tagDir, 0755); err != nil {
				return err
			}

			if err := os.WriteFile(tagIndexPath, html, 0644); err != nil {
				return err
			}
		}
	}

	// Build legal page
	legalDir := filepath.Join(b.config.Build.OutputDir, "legal")
	legalIndexPath := filepath.Join(legalDir, "index.html")

	legalData := *baseData
	legalData.Title = "Legal Notice"
	legalData.LegalContent = site.LegalContent

	legalHTML, err := b.engine.RenderTemplate("legal", &legalData)
	if err != nil {
		return fmt.Errorf("failed to render legal page: %w", err)
	}

	if err := os.MkdirAll(legalDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(legalIndexPath, legalHTML, 0644)
}

// Deprecated: Old buildPages function kept for reference - will be removed
func (b *Builder) buildPages(site *content.Site, assetsVersion string, includeDrafts bool, preview bool) error {
	return b.buildPagesIncremental(site, assetsVersion, includeDrafts, preview, &FileCache{FileModTimes: make(map[string]time.Time)})
}

func (b *Builder) buildSearchIndex(site *content.Site) error {
	var entries []SearchIndexEntry

	for _, post := range site.Posts {
		// Extract plain text from HTML (simplified)
		text := strings.ReplaceAll(string(post.HTML), "<[^>]*>", "")
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.TrimSpace(text)

		entries = append(entries, SearchIndexEntry{
			Title:   post.Title,
			Summary: post.Summary,
			Text:    text,
			Tags:    post.Tags,
			URL:     post.URL,
		})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	searchDir := filepath.Join(b.config.Build.OutputDir, "search")
	if err := os.MkdirAll(searchDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(searchDir, "index.json"), data, 0644)
}

func (b *Builder) injectLiveReloadScript() error {
	liveReloadPath := filepath.Join(b.config.Build.OutputDir, "assets", "js", "live-reload.js")
	liveReloadScript := fmt.Sprintf(`
(function() {
  try {
    var ws = new WebSocket('ws://localhost:%d/ws');
    ws.onmessage = function(event) {
      if (event.data === 'reload') {
        window.location.reload();
      }
    };
    ws.onclose = function(event) {
      // Silently ignore WebSocket connection failures
    };
    ws.onerror = function(error) {
      // Silently ignore WebSocket errors in preview mode
    };
  } catch (e) {
    // Silently ignore WebSocket connection failures
  }
})();
`, b.config.Server.Port)
	return os.WriteFile(liveReloadPath, []byte(liveReloadScript), 0644)
}

// copyImageAssets copies images referenced in posts to the output directory
func (b *Builder) copyImageAssets(site *content.Site, cache *FileCache) error {
	processedImages := make(map[string]bool) // Track which images have been processed to avoid duplicates

	for _, post := range site.Posts {
		for _, imgSrc := range post.ImageSources {
			// Handle image paths by resolving them relative to content directory
			// If imgSrc is absolute (starts with /), we need to find it in content/ folder
			var sourcePath string
			if strings.HasPrefix(imgSrc, "/") {
				// Absolute path like /images/myimage.jpg - look in content/images/myimage.jpg
				sourcePath = filepath.Join(b.config.ContentDir(), strings.TrimPrefix(imgSrc, "/"))
			} else {
				// Relative path - first try relative to content directory
				sourcePath = filepath.Join(b.config.ContentDir(), imgSrc)

				// If not found, try looking for it in static directory
				if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
					// Try in static directory as fallback
					staticPath := filepath.Join(b.config.StaticDir(), imgSrc)
					if _, err := os.Stat(staticPath); err == nil {
						sourcePath = staticPath
					} else {
						// If not in static either, we might need special handling for relative paths
						fmt.Printf("Warning: Image file not found for path: %s\n", imgSrc)
						continue
					}
				}
			}

			// Verify the image file exists
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				fmt.Printf("Warning: Image file not found: %s\n", sourcePath)
				continue
			}

			// Skip if already processed
			if processedImages[sourcePath] {
				continue
			}
			processedImages[sourcePath] = true

			// Determine destination path
			var destPath string
			if strings.HasPrefix(imgSrc, "/") {
				// For absolute paths, copy to output root (so /images/myimage.jpg -> /images/myimage.jpg)
				destPath = filepath.Join(b.config.Build.OutputDir, strings.TrimPrefix(imgSrc, "/"))
			} else {
				// For relative paths, preserve the relative structure in output
				destPath = filepath.Join(b.config.Build.OutputDir, imgSrc)
			}

			// Check if image needs to be copied (incremental build)
			shouldCopy, err := b.shouldRebuildFile(sourcePath, destPath, cache)
			if err != nil {
				// If there's an error checking, copy to be safe
				shouldCopy = true
			}

			if !shouldCopy {
				continue // Skip this image, it's up to date
			}

			// Create destination directory
			destDir := filepath.Dir(destPath)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory for image %s: %w", sourcePath, err)
			}

			// Copy the image file
			if err := b.copyFile(sourcePath, destPath); err != nil {
				return fmt.Errorf("failed to copy image %s to %s: %w", sourcePath, destPath, err)
			}
		}
	}
	return nil
}

// ReloadTemplates reloads all templates from disk
func (b *Builder) ReloadTemplates() error {
	return b.engine.ReloadTemplates()
}
