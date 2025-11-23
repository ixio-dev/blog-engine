# Missing Pieces from PRD Implementation

This document outlines the functionality that is specified in the PRD but not yet fully implemented in the current codebase.

## 1. Actual SCP Deployment (COMPLETED)

### Current State
- The deploy module calculates hash comparisons correctly
- Build process works with draft filtering
- Actual file transfer via SCP using `golang.org/x/crypto/ssh`
- Supports key-based authentication with fallback options
- Proper error handling for network failures
- Creates remote directories as needed

## 2. Draft Toggle in Preview UI

### Current State
- JavaScript has code to dispatch 'toggle-drafts' event
- Site builds with drafts when preview command is used
- No dynamic toggle functionality between showing/hiding drafts

### What's Needed
- Implement Go server endpoint/handler to toggle draft visibility
- Update UI to show/hide draft indication based on current setting
- Allow users to toggle draft visibility without restarting the server

## 3. Build Performance Optimization (COMPLETED)

### Current State
- Implemented file modification time tracking using a .build_cache.json file
- Only rebuild files that have actually changed since the last build
- Skip copying files/assets if source hasn't changed from last build
- Smart rebuild detection checks modification times of source vs destination files
- Static files, post pages, and image assets are only processed when modified
- Fallback to full rebuild if cache is corrupted or missing
- Performance scales well with larger content sets

## 4. Enhanced SPA-like Navigation (COMPLETED)

### Current State
- Enhanced transitions with smooth fade and slide animations
- Loading progress bar indicator during navigation
- Prefetching of linked pages for faster loading
- Active link highlighting for better navigation
- History API integration for proper URL handling

## 5. Error Handling and User Feedback (COMPLETED)

### Current State
- Added comprehensive unit tests for core functionality
- Better validation with clear error messages
- Improved handling of edge cases

## 6. Additional CLI Features (COMPLETED)

### Current State
- Added `blog new [title]` command to create new blog posts
- Command creates properly formatted markdown file with frontmatter
- Includes slug generation and proper formatting

## 7. Testing (COMPLETED)

### Current State
- Added comprehensive unit tests for config, content, build, server, templates, and validate modules
- Tests cover validation, parsing, building, and templating functionality
- All tests passing

## 8. Image/Asset Handling for Blog Posts (COMPLETED)

### Current State
- Added HTML parsing functionality to extract image sources from rendered content
- Implemented image extraction during post parsing phase
- Added build process step to copy referenced images from content/static directories to output
- Supports both absolute paths (e.g., /images/myimage.jpg) and relative paths
- Proper handling of duplicate images to avoid copying the same file multiple times
- Comprehensive tests for image extraction and processing

## 9. Syntax Highlighting for Code Blocks (COMPLETED)

### Current State
- Integrated `github.com/yuin/goldmark-highlighting` extension with Goldmark
- Updated markdown parsing for both regular posts and legal content
- Configured with GitHub-compatible syntax highlighting style
- Added tests to verify syntax highlighting functionality
- Code blocks now render with proper syntax highlighting (colored spans, etc.)