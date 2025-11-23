# Tech Blog - Static Blog Generator

A minimal, high-performance static blog system built from Markdown files, optimized for developers and technical readers. The system produces static HTML, CSS, and JS files and deploys them via SCP. File names remain stable for maximum cache efficiency.

## Features

- **Zero server-side logic** beyond static file serving
- **Developer-friendly experience**: fast builds, predictable filenames, clean architecture
- **Technically appealing design** with typography-focused layout
- **Draft workflow** built into preview, excluded from production
- **File watching** with live reload for instant preview
- **Enhanced SPA-like navigation** with smooth transitions and loading indicators
- **Improved client-side search** with ESC to close, click-to-close, and better UI
- **Draft indicator toggle** in preview mode
- **SCP deployment** with key-based authentication and hash comparison
- **Comprehensive CLI** with build, preview, deploy, validate, clean, and new commands

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd tech-blog

# Build the binary
go build -o blog
```

## Quick Start

1. Copy the `my-tech-blog` directory as a starting point:
   ```bash
   cp -r my-tech-blog my-new-blog
   cd my-new-blog
   ```

2. Customize the configuration in `config.yaml`

3. Create a new post:
   ```bash
   ../blog new "My First Post"
   ```

4. Start the preview server:
   ```bash
   ../blog preview
   ```

## Commands

The blog system provides several CLI commands:

- `blog build` - Builds the static site for production (excludes drafts)
- `blog preview` - Starts local server with drafts enabled and live reload
- `blog deploy` - Validates, builds, and deploys site via SCP
- `blog validate` - Validates configuration and content without building
- `blog clean` - Removes build artifacts
- `blog new [title]` - Creates a new blog post with proper formatting

## Configuration

The `config.yaml` file contains all configuration options:

```yaml
site:
  title: "Tech Blog"
  baseUrl: "https://example.com"
  author: "Your Name"

build:
  outputDir: "./dist"
  drafts: false

server:
  port: 4000

deploy:
  host: "user@example.com"
  path: "/var/www/blog"
```

## Content Structure

Posts are written in Markdown with YAML frontmatter:

```markdown
---
title: "Welcome to My Tech Blog"
date: "2024-01-15"
tags: ["introduction", "blogging"]
draft: false
---

# Welcome to My Tech Blog

This is the first post on my new tech blog...
```

## Templates

The system uses Go templates with the following files expected:
- `base.gohtml` - Base layout template
- `index.gohtml` - Index page template
- `post.gohtml` - Individual post template
- `tag.gohtml` - Tag page template
- `tags.gohtml` - Tags index template

## Development

### Adding Tests

Tests are written using Go's standard `testing` package. New tests should be added to `<module>_test.go` files in their respective directories.

### SPA Navigation

The JavaScript provides enhanced SPA-like navigation with:
- Smooth transitions between pages
- Loading progress bar
- Prefetching of linked pages
- Active link highlighting
- History API integration

## Architecture

```
                   +-------------------------+
                   |       CLI Driver       |
                   +-----------+-------------+
                               |
                               v
                +-----------------------------+
                |       Config Validator      |
                +-----------------------------+
                               |
                               v
          +-------------------------------------------+
          |               Build Engine                |
          |-------------------------------------------|
          |  Markdown Parser | Template Renderer      |
          |  Tag Builder     | Search Index Builder   |
          +------------------+------------------------+
                               |
                               v
                      +----------------+
                      | Output Folder  |
                      +----------------+
                               |
                +-------------------------------+
                |      Change Detector          |
                +-------------------------------+
                 |                         |
                 | no changes               | changes
                 v                         v
         +----------------+        +---------------------+
         |   Exit         |        | Deployment (SCP)    |
         +----------------+        +---------------------+
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[Specify your license here]