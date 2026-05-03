(function () {
  const STORAGE_KEY = 'webmail-theme';
  const THEMES = {
    'lainrocks-old': '/static/css/themes/lainrocks-old.css',
    'lainrocks':     '/static/css/themes/lainrocks.css',
    'yotsuba':       '/static/css/themes/yotsuba.css',
  };
  const DEFAULT = 'lainrocks-old';

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
    const sel = document.getElementById('theme-select');
    if (!sel) return;
    sel.value = localStorage.getItem(STORAGE_KEY) || DEFAULT;
    sel.addEventListener('change', function () {
      applyTheme(this.value);
    });
  });
})();