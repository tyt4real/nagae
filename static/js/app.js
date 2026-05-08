(function () {
  const STORAGE_KEY = 'webmail-theme';
  const THEMES = {
    'lainrocks-old': '/static/css/themes/lainrocks-old.css',
    'lainrocks':     '/static/css/themes/lainrocks.css',
    'yotsuba':       '/static/css/themes/yotsuba.css',
  };
  const DEFAULT = 'lainrocks-old';

  // Inject the theme <link> as early as possible (before DOMContentLoaded)
  // to avoid a flash of the default colours.
  const linkEl = document.createElement('link');
  linkEl.rel = 'stylesheet';
  linkEl.id  = 'theme-stylesheet';
  document.head.appendChild(linkEl);

  function applyTheme(name) {
    const href = THEMES[name] || THEMES[DEFAULT];
    linkEl.href = href;
    localStorage.setItem(STORAGE_KEY, name);
    const sel = document.getElementById('theme-select');
    if (sel) sel.value = name;
  }

  const saved = localStorage.getItem(STORAGE_KEY) || DEFAULT;
  applyTheme(saved);

  document.addEventListener('DOMContentLoaded', function () {
    // Theme switcher.
    const sel = document.getElementById('theme-select');
    if (sel) {
      sel.value = localStorage.getItem(STORAGE_KEY) || DEFAULT;
      sel.addEventListener('change', function () { applyTheme(this.value); });
    }

    // Desktop notifications — only on inbox/search pages.
    if (!window.location.pathname.startsWith('/inbox') &&
        !window.location.pathname.startsWith('/')) return;

    let lastUnseen = null;

    function requestPermission() {
      if ('Notification' in window && Notification.permission === 'default') {
        Notification.requestPermission();
      }
    }

    function poll() {
      fetch('/notify/poll', { credentials: 'same-origin' })
        .then(function (r) { return r.ok ? r.json() : null; })
        .then(function (data) {
          if (!data) return;
          const unseen = data.unseen;

          if (lastUnseen !== null && unseen > lastUnseen &&
              'Notification' in window && Notification.permission === 'granted') {
            const delta = unseen - lastUnseen;
            new Notification('post/', {
              body: delta === 1
                ? 'You have 1 new message.'
                : 'You have ' + delta + ' new messages.',
              icon: '/static/img/icon.png',
              tag:  'webmail-new-mail',
            });
          }

          lastUnseen = unseen;

          // Update page title badge.
          if (unseen > 0) {
            document.title = '(' + unseen + ') post/';
          } else {
            document.title = document.title.replace(/^\(\d+\)\s*/, '');
          }
        })
        .catch(function () { /* ignore network errors during poll */ });
    }

    requestPermission();
    poll();
    setInterval(poll, 60000); // every 60 seconds
  });
})();