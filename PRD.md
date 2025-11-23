# Static Blog System — Product Requirements Document (PRD)

## 1. Overview
A minimal, high-performance static blog system built from Markdown files, optimized for developers and technical readers. The system produces static HTML, CSS, and JS files and deploys them via SCP. File names remain stable for maximum cache efficiency.

A local preview server enables fast iteration. Client-side features (search, navigation, drafts toggle) simulate a dynamic experience but are generated fully statically.

---

## 2. Goals
- Zero server-side logic beyond static file serving.
- Developer-friendly experience: fast builds, predictable filenames, clean architecture.
- Technically appealing design and interaction patterns.
- Draft workflow built into preview, excluded from production.
- Ability to detect changes, rebuild, and deploy in one command.

---

## 3. Target Users
- Software engineers.
- Technical bloggers.
- People who prefer writing in Markdown and deploying via CLI.

---

## 4. User Experience
### 4.1 Visual Design
- Minimalistic, high-contrast, typography-focused.
- Clean layout with strong alignment and consistent spacing.
- Images included in the blog posts are handled properly.
- Code-friendly styling (monospace headings, syntax highlighting, dark/light themes).

### 4.2 Page Layout
All pages share one consistent layout:
- Fixed header with site title and navigation.
- Content loaded via full-page transition animations to simulate a dynamic SPA.
- Images included in the blog posts are handled properly.
- Fixed footer with simple copyright and legal notice link.

### 4.3 Blog Features
- Blog index with chronological listing.
- Individual post pages.
- Tag pages generated automatically.
- Client-side search powered by a prebuilt JSON index.
- Draft indicator toggle in preview mode.
- Smooth navigation simulated using JS transitions.

---

## 5. Functional Requirements
### 5.1 Content Management
- Markdown → HTML conversion.
- Frontmatter parameters: title, date, tags, draft flag.
- Validation of metadata before build.
- Stable output file names based on slug.

### 5.2 Build System
- Incremental rebuild support.
- Deterministic output.
- Generates:
  - `/index.html`
  - `/posts/<slug>/index.html`
  - `/tags/<tag>/index.html`
  - `/assets/` (CSS/JS)
  - `/search/index.json`

### 5.3 Local Preview
- Watches for file changes.
- Provides draft toggle in UI.
- Serves pages exactly like production.

### 5.4 Deployment
- Single command:
  - Validates metadata.
  - Compares current built output to last deployed hash.
  - Rebuilds if changes exist.
  - Uploads via SCP.

---

## 6. Non‑Functional Requirements
- Build time < 2 seconds for 100 posts.
- File names must remain stable across builds.
- Works offline.
- All features functional in static hosting environments.

---

## 7. Technical Architecture
### 7.1 Components
- **Markdown Parser** → HTML generator.
- **Template Engine** → Consistent page layout.
- **Search Indexer** → Builds JSON index.
- **Draft Filter** → Included only in preview.
- **Change Detector** → Hash-based comparison.
- **Preview Server** → Local dev environment.
- **Deployer** → SCP uploader.

### 7.2 Interaction Flow
1. User runs a CLI command.
2. CLI parses config + content.
3. Validator checks structure and metadata.
4. Build system generates HTML + assets.
5. Diff engine compares output hashes.
6. Preview server serves files OR deployer uploads via SCP.

---

## 8. Architecture Diagram
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

---

## 9. CLI Specification
### Command: `blog build`
- Validates all metadata.
- Builds full static output.
- Excludes drafts.
- Generates search index.
- Reports build summary.

### Command: `blog preview`
- Starts local server at `http://localhost:4000`.
- Includes drafts.
- Offers UI toggle to show/hide drafts.
- Watches for file changes and rebuilds instantly.

### Command: `blog deploy`
- Runs validation.
- Builds output.
- Computes hash of final build.
- Compares to previous deployment hash.
- If different: uploads via SCP.
- Stores new hash.

### Command: `blog validate`
- Runs validations without building.
- Returns clear error messages.

### Command: `blog clean`
- Deletes build output.
- Deletes previous hash.

---

## 10. Configuration (config.yaml)
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

---

## 11. Success Metrics
- Build+deploy cycle under 3 seconds.
- No broken links in any build.
- No change in filenames unless content changes.
- Search index under 30 KB for 100 posts.

---

End of document.

