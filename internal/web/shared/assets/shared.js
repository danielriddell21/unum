/* unum shared theme + dark/light toggle — loaded after window.UNUM_CONFIG */
(function () {
  'use strict';

  const _cfg    = window.UNUM_CONFIG || {};
  const _LS_KEY = 'unum-mode';

  // Theme maps are injected by the Go server via window.UNUM_CONFIG.themes.
  // Shape: { dark: { cyber: { '--bg': '#...', ... }, ... }, light: { ... } }
  const DARK_THEMES  = (_cfg.themes && _cfg.themes.dark)  || {};
  const LIGHT_THEMES = (_cfg.themes && _cfg.themes.light) || {};

  function resolveMode() {
    const saved = localStorage.getItem(_LS_KEY);
    return (saved === 'light') ? 'light' : 'dark';
  }

  function applyMode(mode) {
    const themes = mode === 'dark' ? DARK_THEMES : LIGHT_THEMES;
    const name   = mode === 'dark' ? (_cfg.darkTheme || 'cyber') : (_cfg.lightTheme || 'clean');
    const vars   = themes[name] || themes[Object.keys(themes)[0]];
    const root   = document.documentElement;
    for (const [k, v] of Object.entries(vars)) root.style.setProperty(k, v);
    localStorage.setItem(_LS_KEY, mode);
    const btn = document.getElementById('mode-toggle');
    if (btn) btn.textContent = mode === 'dark' ? '☀' : '☾';
  }

  function renderModeToggle() {
    const btn = document.createElement('button');
    btn.id = 'mode-toggle';
    btn.title = 'Toggle dark/light';
    btn.style.cssText = 'background:none;border:1px solid var(--border);color:var(--muted);' +
      'font-size:13px;padding:2px 7px;border-radius:3px;cursor:pointer;font-family:inherit;' +
      'transition:all 0.15s;margin-left:auto;';
    btn.textContent = resolveMode() === 'dark' ? '☀' : '☾';
    btn.onmouseover = () => { btn.style.borderColor = 'var(--border-active)'; btn.style.color = 'var(--border-active)'; };
    btn.onmouseout  = () => { btn.style.borderColor = 'var(--border)'; btn.style.color = 'var(--muted)'; };
    btn.onclick = () => applyMode(resolveMode() === 'dark' ? 'light' : 'dark');
    return btn;
  }

  // Expose for app.js files that call applyMode/resolveMode after boot
  window._unumApplyMode   = applyMode;
  window._unumResolveMode = resolveMode;
  window._unumModeToggle  = renderModeToggle;

  document.addEventListener('DOMContentLoaded', function () {
    applyMode(resolveMode());
    const hdr = document.getElementById('header');
    if (hdr) hdr.appendChild(renderModeToggle());
    if (_cfg.version) {
      var sb = document.getElementById('status-bar');
      if (sb) {
        var ver = document.createElement('span');
        ver.style.cssText = 'color:var(--muted);font-size:10px;margin-left:8px;';
        ver.textContent = _cfg.version;
        sb.appendChild(ver);
      }
    }
  });
})();
