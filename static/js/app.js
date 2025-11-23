// Minimal app shell: dark-mode, transitions, prefetch-on-hover, and keyboard handlers.
// Assumes search.js is loaded too.
(function () {
  'use strict';
  const root = document.documentElement;
  const body = document.body;
  const storageKey = 'site:theme';

  function applyTheme(t) {
    if (t === 'light') root.classList.add('light'); else root.classList.remove('light');
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
    if (saved) applyTheme(saved);
    else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) applyTheme('light');
  } catch (e) { /* ignore */ }

  document.addEventListener('click', (ev) => {
    const t = ev.target;
    if (t && t.id === 'theme-toggle') { toggleTheme(); }
  });

  // Keyboard shortcuts
  document.addEventListener('keydown', (e) => {
    if (e.key === '/') { // focus search
      const inSearch = document.getElementById('search-input');
      if (inSearch) { e.preventDefault(); inSearch.focus(); showSearchUI(true); }
    } else if (e.key === 'd') { // draft toggle: preview server injects handler that toggles drafts; here, fire event
      document.dispatchEvent(new CustomEvent('toggle-drafts'));
    }
  });

  // Prefetch links on hover
  let prefetchSet = new Set();
  function prefetch(url) {
    if (!url || prefetchSet.has(url) || url.indexOf(location.origin) !== 0) return;
    prefetchSet.add(url);
    const l = document.createElement('link');
    l.rel = 'prefetch';
    l.href = url;
    document.head.appendChild(l);
  }
  document.addEventListener('mouseover', (e) => {
    const a = e.target.closest && e.target.closest('a');
    if (a && a.href) prefetch(a.href);
  });

  // Basic transition: fade out/in when clicking internal links
  document.addEventListener('click', (e) => {
    const a = e.target.closest && e.target.closest('a');
    if (!a || a.target === '_blank' || a.href.indexOf(location.origin) !== 0) return;
    // allow ctrl/meta click
    if (e.metaKey || e.ctrlKey || e.shiftKey) return;
    e.preventDefault();
    const url = a.href;
    document.getElementById('content').classList.add('fading-out');
    setTimeout(() => { location.href = url; }, 220);
  });

  // Simple CSS class for fade-out (needs CSS class .fading-out defined)
  // Add a CSS rule via JS if not present
  (function ensureFadeCss() {
    const id = '__fade-css';
    if (document.getElementById(id)) return;
    const style = document.createElement('style');
    style.id = id;
    style.textContent = `
      .content-container { transition: opacity .22s ease; }
      .content-container.fading-out { opacity: 0; }
    `;
    document.head.appendChild(style);
  })();

  // Search UI helpers (used by search.js)
  window.showSearchUI = function (show) {
    const ui = document.getElementById('search-ui');
    if (!ui) return;
    if (show) { ui.classList.remove('hidden'); ui.setAttribute('aria-hidden','false'); const input = document.getElementById('search-input'); if (input) input.focus(); }
    else { ui.classList.add('hidden'); ui.setAttribute('aria-hidden','true'); }
  };

})();
