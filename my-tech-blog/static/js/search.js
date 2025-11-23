// Simple client-side search using a prebuilt /search/index.json
// Lightweight tokenization and substring match. Designed for a few hundred posts.
(function () {
  'use strict';
  const baseUrl = window.BASE_URL || '';
  const indexUrl = baseUrl + '/search/index.json';
  let index = [];
  let ready = false;

  async function loadIndex() {
    try {
      const res = await fetch(indexUrl, {cache: 'reload'});
      index = await res.json();
      ready = true;
    } catch (err) {
      console.warn('Search index load failed', err);
      ready = false;
    }
  }

  function tokenize(s) {
    if (!s) return [];
    return s.toLowerCase().split(/[^a-z0-9]+/).filter(Boolean);
  }

  function scoreItem(item, terms) {
    // simple scoring: +2 for title matches, +1 for summary/body matches, +.5 per tag match
    let score = 0;
    const hay = (item.title + ' ' + (item.summary || '') + ' ' + (item.text || '')).toLowerCase();
    terms.forEach(t => {
      if (item.title.toLowerCase().includes(t)) score += 2;
      if (hay.includes(t)) score += 1;
      if ((item.tags || []).some(tag => tag.toLowerCase().includes(t))) score += 0.5;
    });
    return score;
  }

  function search(q) {
    if (!ready || !q) return [];
    const terms = tokenize(q);
    if (terms.length === 0) return [];
    const results = index.map(item => ({ item, score: scoreItem(item, terms) }))
      .filter(r => r.score > 0)
      .sort((a,b) => b.score - a.score)
      .slice(0, 30)
      .map(r => r.item);
    return results;
  }

  // Render results
  function renderResults(results) {
    const container = document.getElementById('search-results');
    container.innerHTML = '';
    if (results.length === 0) { container.innerHTML = '<div class="search-empty">No results</div>'; return; }
    const ul = document.createElement('ul');
    ul.className = 'search-results-list';
    results.forEach(r => {
      const li = document.createElement('li');
      li.className = 'search-result';
      li.innerHTML = `<a href="${r.url}"><h3>${escapeHtml(r.title)}</h3><p class="summary">${escapeHtml(r.summary || '')}</p></a>`;
      ul.appendChild(li);
    });
    container.appendChild(ul);
  }

  function escapeHtml(s) { return (s||'').replace(/[&<>"']/g, (m) => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[m])); }

  // Wire up UI
  function wire() {
    const toggle = document.getElementById('search-toggle');
    if (toggle) toggle.addEventListener('click', () => window.showSearchUI(true));

    const input = document.getElementById('search-input');
    if (!input) return;
    input.addEventListener('input', (e) => {
      const q = e.target.value;
      const results = search(q);
      renderResults(results);
    });

    const close = document.getElementById('search-close');
    if (close) close.addEventListener('click', () => window.showSearchUI(false));
  }

  // Init
  loadIndex().then(wire);
})();
