// Enhanced SPA-like navigation: dark-mode, advanced transitions, prefetch-on-hover, and keyboard handlers.
// Assumes search.js is loaded too.
(function () {
  'use strict';
  const root = document.documentElement;
  const body = document.body;
  const storageKey = 'site:theme';

  let loadingIndicator;
  let progressBar;

  function applyTheme(t) {
    if (t === 'light') {
      root.classList.add('light');
    } else {
      root.classList.remove('light');
    }
  }

  function toggleTheme() {
    const current = root.classList.contains('light') ? 'light' : 'dark';
    const next = current === 'light' ? 'dark' : 'light';
    applyTheme(next);
    try { localStorage.setItem(storageKey, next); } catch(e) {}
  }

  // Initialize theme from storage or system
  try {
    const saved = localStorage.getItem(storageKey);
    if (saved) {
      applyTheme(saved);
    } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
      applyTheme('light');
    } else {
      applyTheme('dark');
    }
  } catch (e) { /* ignore */ }

  document.addEventListener('click', (ev) => {
    const t = ev.target;
    if (t && t.id === 'theme-toggle') {
      toggleTheme();
    } else if (t && t.id === 'draft-toggle') {
      toggleDrafts();
    }
  });

  // Keyboard shortcuts
  document.addEventListener('keydown', (e) => {
    if (e.key === '/') { // focus search
      const inSearch = document.getElementById('search-input');
      if (inSearch) { e.preventDefault(); inSearch.focus(); showSearchUI(true); }
    } else if (e.key === 'd') { // draft toggle: preview server injects handler that toggles drafts; here, fire event
      toggleDrafts();
    }
  });

  // Establish WebSocket connection if in preview mode
  let ws = null;
  function initWebSocket() {
    // Check if we're in preview mode by seeing if we're on the expected preview port
    // or by checking if WebSocket endpoint is available
    try {
      // Only connect if we're likely in preview mode (localhost with expected port)
      const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
      // Use the same port as the current page for WebSocket connection
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = wsProtocol + '//' + window.location.host + '/ws';

      ws = new WebSocket(wsUrl);
      window.ws = ws; // Make available globally for draft toggle

      ws.onopen = function() {
        console.log('WebSocket connection established for draft toggling');
      };

      ws.onclose = function() {
        console.log('WebSocket connection closed');
        // Try to reconnect after a delay
        setTimeout(initWebSocket, 5000);
      };

      ws.onerror = function(error) {
        console.error('WebSocket error:', error);
      };
    } catch(e) {
      console.error('Failed to establish WebSocket connection:', e);
    }
  }

  // Initialize WebSocket if we're likely in preview mode
  if (location.hostname === 'localhost' || location.hostname === '127.0.0.1') {
    initWebSocket();
  }

  // Draft toggle functionality
  let draftsVisible = typeof window.DRAFTS_VISIBLE !== 'undefined' ? window.DRAFTS_VISIBLE : true;  // Use server-provided state or default to true

  function updateDraftToggleUI() {
    const draftToggleBtn = document.getElementById('draft-toggle');
    if (!draftToggleBtn) return;

    if (draftsVisible) {
      draftToggleBtn.textContent = '📝';  // Pen/edit icon when drafts are visible
      draftToggleBtn.title = 'Hide drafts';
      draftToggleBtn.setAttribute('aria-label', 'Hide drafts');
      draftToggleBtn.classList.remove('drafts-hidden');
      draftToggleBtn.classList.add('drafts-visible');
    } else {
      draftToggleBtn.textContent = '👁️';  // Eye icon when drafts are hidden
      draftToggleBtn.title = 'Show drafts';
      draftToggleBtn.setAttribute('aria-label', 'Show drafts');
      draftToggleBtn.classList.remove('drafts-visible');
      draftToggleBtn.classList.add('drafts-hidden');
    }
  }

  function toggleDrafts() {
    // Send message to WebSocket server to toggle drafts
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send('toggle-drafts');
      // Update local state immediately for responsive UI
      draftsVisible = !draftsVisible;
      updateDraftToggleUI();
    } else {
      console.warn('WebSocket not connected, cannot toggle drafts');
    }
  }

  // Initialize the draft toggle button UI when the DOM is loaded
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', updateDraftToggleUI);
  } else {
    // DOM is already loaded, run immediately
    updateDraftToggleUI();
  }

  // Enhanced prefetching
  let prefetchSet = new Set();
  function prefetch(url) {
    if (!url || prefetchSet.has(url) || url.indexOf(location.origin) !== 0) return;
    prefetchSet.add(url);

    // Create link element for prefetch
    const link = document.createElement('link');
    link.rel = 'prefetch';
    link.href = url;
    document.head.appendChild(link);

    // Also prefetch via fetch for HTML pages to improve loading speed
    if (url.endsWith('/') || url.endsWith('.html')) {
      fetch(url, {method: 'GET', cache: 'force-cache'})
        .catch(() => {}); // Ignore errors
    }
  }

  document.addEventListener('mouseover', (e) => {
    const a = e.target.closest && e.target.closest('a');
    if (a && a.href) {
      prefetch(a.href);
    }
  });

  // Create loading indicator
  function createLoadingIndicator() {
    if (loadingIndicator) return;

    const indicator = document.createElement('div');
    indicator.id = 'page-loading';
    indicator.innerHTML = `
      <div class="loading-bar">
        <div class="loading-progress"></div>
      </div>
    `;
    document.body.appendChild(indicator);

    loadingIndicator = indicator;
    progressBar = indicator.querySelector('.loading-progress');
  }

  // Show loading indicator
  function showLoading() {
    if (!loadingIndicator) createLoadingIndicator();
    loadingIndicator.style.display = 'block';
    if (progressBar) progressBar.style.width = '10%'; // Start progress

    // Simulate progress
    let progress = 10;
    const interval = setInterval(() => {
      progress += Math.random() * 15;
      if (progress >= 90) {
        clearInterval(interval);
        return;
      }
      if (progressBar) progressBar.style.width = progress + '%';
    }, 100);

    return interval;
  }

  // Complete loading
  function completeLoading() {
    if (!loadingIndicator) return;

    if (progressBar) progressBar.style.width = '100%';
    setTimeout(() => {
      loadingIndicator.style.display = 'none';
      if (progressBar) progressBar.style.width = '0';
    }, 200);
  }

  // Enhanced SPA navigation with better transitions
  document.addEventListener('click', (e) => {
    const a = e.target.closest && e.target.closest('a');
    if (!a || a.target === '_blank' || a.href.indexOf(location.origin) !== 0) return;
    // allow ctrl/meta click
    if (e.metaKey || e.ctrlKey || e.shiftKey) return;

    e.preventDefault();
    const url = a.href;

    // Show loading indicator
    const progressInterval = showLoading();

    // Fade out content with enhanced animation
    const content = document.getElementById('content');
    if (content) {
      content.style.transform = 'translateY(-10px)';
      content.style.opacity = '0';
      content.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
    }

    // Delay navigation slightly to allow for visual feedback
    setTimeout(() => {
      // Load the new page via fetch
      fetch(url)
        .then(response => response.text())
        .then(html => {
          // Parse the response to get the new content
          const parser = new DOMParser();
          const doc = parser.parseFromString(html, 'text/html');
          const newContent = doc.querySelector('#content');

          if (newContent) {
            // Update the content
            if (content) {
              content.innerHTML = newContent.innerHTML;

              // Update page title
              document.title = doc.title;

              // Update URL without page reload
              history.pushState({}, '', url);

              // Fade in the new content
              content.style.opacity = '1';
              content.style.transform = 'translateY(0)';

              // Update active nav links
              updateActiveNavLinks(url);
            }
          } else {
            // If content not found, fallback to regular navigation
            location.href = url;
          }

          // Complete loading
          clearInterval(progressInterval);
          completeLoading();
        })
        .catch(error => {
          console.error('Navigation error:', error);
          // Fallback to regular navigation on error
          location.href = url;
        });
    }, 100);
  });

  // Update active navigation links
  function updateActiveNavLinks(currentUrl) {
    // Remove active class from all links
    document.querySelectorAll('.site-nav a, .post-nav a').forEach(link => {
      link.classList.remove('active');
    });

    // Add active class to current page link
    document.querySelectorAll('.site-nav a, .post-nav a').forEach(link => {
      if (link.href === currentUrl) {
        link.classList.add('active');
      }
    });
  }

  // Handle browser back/forward buttons
  window.addEventListener('popstate', (e) => {
    location.reload(); // Simple approach for now - could be enhanced to use SPA navigation for history
  });

  // Initialize loading indicator if needed
  createLoadingIndicator();
  loadingIndicator.style.display = 'none';

  // Search UI helpers (used by search.js)
  window.showSearchUI = function (show) {
    const ui = document.getElementById('search-ui');
    if (!ui) return;
    if (show) {
      ui.classList.remove('hidden');
      ui.setAttribute('aria-hidden','false');
      const input = document.getElementById('search-input');
      if (input) input.focus();
    } else {
      ui.classList.add('hidden');
      ui.setAttribute('aria-hidden','true');
    }
  };

  // Add ESC key functionality to close search
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      const ui = document.getElementById('search-ui');
      if (ui && !ui.classList.contains('hidden')) {
        showSearchUI(false);
      }
    }
  });

  // Also close search when clicking outside of it
  document.addEventListener('click', (e) => {
    const ui = document.getElementById('search-ui');
    const searchToggle = document.getElementById('search-toggle');

    if (ui && !ui.classList.contains('hidden') &&
        !ui.contains(e.target) &&
        !searchToggle.contains(e.target)) {
      showSearchUI(false);
    }
  });

})();
