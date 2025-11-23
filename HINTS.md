# Notes & integration tips
- The templates use simple placeholders. Replace {{ ... }}, {% ... %}, or {{ for ... }} constructs with whatever templating engine you use (Jinja2, Liquid, Mustache, Handlebars, Go templates, etc.). Keep variable names consistent with your renderer (examples: post.slug, post.html, post.title, posts).
- Use assets_version query param (set to ?v=<sha256> or empty) to preserve stable filenames while enabling cache invalidation when assets change. The build engine should compute assets_version once per build.
- search-index.json structure expected by search.js:

```
[
  {
    "title": "Post title",
    "summary": "Short summary",
    "text": "Plain text extract of the body",
    "tags": ["tag1","tag2"],
    "url": "/posts/post-slug/"
  }
]
```

- For deterministic builds, avoid inserting variable timestamps into HTML that are part of checksummed files. If you must expose a build time on the site, render it into a separate non-checksummed file or marker.
- live-reload.js must be injected only in preview mode — not in production builds.
- app.js includes a simple navigation interception which degrades to normal navigation if JS is disabled. Keep transitions small and unobtrusive for performance.
