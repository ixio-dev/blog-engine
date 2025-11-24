# Tech Blog - Static Blog Generator

A minimal, high-performance static blog generator built in Go, designed for developers and technical writers. The system converts Markdown files into static HTML, CSS, and JavaScript files optimized for performance and SEO. Deployments are handled via SCP with file hash comparison to minimize transfer times.

## Features

- **Static Site Generation**: Zero server-side logic beyond static file serving
- **Developer Experience**: Fast builds, predictable filenames, clean architecture
- **Typographic Design**: Clean, readable layout optimized for technical content
- **Draft Workflow**: Built-in draft support for preview mode, excluded from production
- **Live Preview**: File watching with live reload for instant preview
- **SPA-like Navigation**: Smooth transitions, loading indicators, and prefetching
- **Client-side Search**: Full-text search with keyboard shortcuts (ESC to close)
- **Simple Deployment**: SCP deployment with smart file comparison to reduce transfers
- **Comprehensive CLI**: Commands for build, preview, deploy, validate, clean, and new posts

## Prerequisites

- Go 1.21 or higher
- Git
- SSH access for deployments (if using deploy feature)

## Installation

```bash
# Clone the repository
git clone https://github.com/ixio-dev/blog-engine.git
cd blog-engine

# Build the binary
go build -o blog

# Or install directly
go install .
```

## Quick Start

1. Copy the example blog directory to start your own:
   ```bash
   cp -r my-tech-blog my-new-blog
   cd my-new-blog
   ```

2. Customize the configuration in `config.yaml` (see Configuration section below)

3. Create your first post:
   ```bash
   ../blog new "Hello World"
   ```

4. Start the preview server to see changes in real-time:
   ```bash
   ../blog preview
   ```
   Visit `http://localhost:4000` to view your blog

5. Build your site for production:
   ```bash
   ../blog build
   ```

6. Deploy to your server:
   ```bash
   ../blog deploy
   ```

## CLI Commands

The blog system provides several commands for different workflows:

- `blog build` - Compiles the static site for production (excludes drafts)
- `blog preview` - Starts local server with drafts enabled and live reload
- `blog deploy` - Validates, builds, and deploys site via SCP
- `blog validate` - Checks configuration and content without building
- `blog clean` - Removes build artifacts from output directory
- `blog new <title>` - Creates a new blog post with proper frontmatter
- `blog --help` - Shows available commands and options

## Configuration

The `config.yaml` file controls your blog's behavior:

```yaml
site:
  title: "My Tech Blog"
  baseUrl: "https://example.com"
  author: "Your Name"
  description: "A blog about technology and development"
  language: "en"

build:
  outputDir: "./dist"
  drafts: false  # Include drafts in production builds (usually false)
  prettyUrls: true  # Use /post/ instead of /post.html

server:
  port: 4000
  host: "localhost"

deploy:
  host: "user@example.com:22"
  path: "/var/www/blog"
  keyFile: "~/.ssh/id_rsa"  # SSH key for authentication
  dryRun: false  # Test mode - show what would be deployed
```

## Content Structure

Posts are written in Markdown with YAML frontmatter at the top:

```markdown
---
title: "Getting Started with Tech Blog"
date: "2024-01-15T10:30:00Z"
tags: ["introduction", "setup", "tutorial"]
draft: false
summary: "Learn how to set up and use the Tech Blog static site generator."
---

# Getting Started with Tech Blog

This is a tutorial about how to use the Tech Blog static site generator...

## Code Example

Here's how you can write code in your posts:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Tech Blog!")
}
```

## Lists and More

- You can include lists
- Tables
- Images
- And other Markdown features
```

### Frontmatter Fields

- `title` (required): The title of your post
- `date` (required): Publication date in RFC3339 format
- `tags` (optional): Array of tags for the post
- `draft` (optional): Whether this is a draft (default: false)
- `summary` (optional): Short summary for index pages

## Templates

The system uses Go's html/template system. Required template files:

- `base.gohtml` - Main layout template (HTML structure, CSS, JS includes)
- `index.gohtml` - Homepage/post list template
- `post.gohtml` - Individual post page template
- `tag.gohtml` - Individual tag page template
- `tags.gohtml` - Tag index page template

Templates have access to structured data including site metadata, posts, and current page context.

## Development

### Running Tests

Run the test suite to ensure everything works:

```bash
go test ./...
```

### Adding Features

1. Fork the repository
2. Create a feature branch
3. Add your changes
4. Include tests when applicable
5. Update documentation if needed
6. Submit a pull request

### Project Structure

```
tech-blog/
├── main.go                 # CLI entry point
├── build/                  # Build system components
├── server/                 # Preview server
├── deploy/                 # Deployment logic
├── config/                 # Configuration handling
├── templates/              # Template parsing and rendering
├── validate/               # Content and config validation
└── content/                # Content processing (Markdown, metadata)
```

## Deployment

Deployment uses SCP with file hash comparison to transfer only changed files:

1. Configure SSH access to your server
2. Update deploy configuration in `config.yaml`
3. Run `blog deploy`

The system compares file hashes to avoid unnecessary transfers, making incremental deployments fast.

## Customization

### Styling

- CSS files are in the `assets/css/` directory
- Modify `main.css` or add additional files
- The default theme is clean and typographically focused

### JavaScript

- Client-side JavaScript is in `assets/js/`
- Includes SPA-like navigation and search functionality
- Modify or extend as needed

### Templates

- All site structure is controlled via Go templates
- Customize layout, navigation, and content presentation
- Templates support partials and template composition

## Troubleshooting

- **Preview server won't start**: Check that the port isn't already in use
- **Build fails**: Run `blog validate` to check for content or config issues
- **Deployment fails**: Verify SSH access and configuration in `config.yaml`
- **Content not showing**: Check frontmatter format and date formatting

## Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Update documentation as needed
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a pull request

### Development Dependencies

- Go 1.21+
- Git
- Basic Unix tools (cp, rm, etc.)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.