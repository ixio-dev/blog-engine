// Very small live reload client; preview server must open a websocket at /_livereload
(() => {
  if (!('WebSocket' in window)) return;
  const wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/_livereload';
  const ws = new WebSocket(wsUrl);
  ws.addEventListener('message', (ev) => {
    try {
      const msg = JSON.parse(ev.data);
      if (msg.type === 'reload') {
        if (msg.path) console.info('Changed:', msg.path);
        location.reload();
      } else if (msg.type === 'rebuild') {
        console.info('Rebuilt:', msg.path || '');
        // optionally do more graceful updates
        location.reload();
      }
    } catch (e) { location.reload(); }
  });
})();
