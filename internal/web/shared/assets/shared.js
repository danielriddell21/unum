/* unum shared theme + dark/light toggle — loaded after window.UNUM_CONFIG */
(function () {
  'use strict';

  const _cfg    = window.UNUM_CONFIG || {};
  const _LS_KEY = 'unum-mode';

  // Merged superset: every CSS variable used across all three tools.
  // Tool-specific vars (array-idx, diff-added, etc.) are a no-op on pages
  // that don't reference them in CSS.
  const DARK_THEMES = {
    cyber: {
      '--bg':'#0D0D0D', '--bg-panel':'#111111', '--bg-hover':'#1A1A2E',
      '--border':'#1E1E1E', '--border-active':'#00D4FF',
      '--text':'#C0C0C0', '--muted':'#3A3A3A',
      // json
      '--array-idx':'#FF6B6B', '--string':'#98C379', '--number':'#E5C07B',
      '--bool-true':'#56B6C2', '--bool-false':'#E06C75', '--null':'#5C6370',
      '--path':'#C678DD', '--search':'#FFD700', '--hash':'#C678DD', '--stats':'#FFD700',
      // diff
      '--bg-added':'#0d1a0d', '--bg-removed':'#1a0d0d',
      '--diff-added':'#98C379', '--diff-removed':'#E06C75', '--diff-hunk':'#00D4FF',
    },
    matrix: {
      '--bg':'#0D0D0D', '--bg-panel':'#0A1A0A', '--bg-hover':'#001A00',
      '--border':'#003300', '--border-active':'#00FF41',
      '--text':'#00CC22', '--muted':'#005500',
      '--array-idx':'#33FF33', '--string':'#00CC22', '--number':'#88FF44',
      '--bool-true':'#00FF41', '--bool-false':'#FF3300', '--null':'#005500',
      '--path':'#39FF14', '--search':'#FFFFFF', '--hash':'#39FF14', '--stats':'#88FF44',
      '--bg-added':'#001800', '--bg-removed':'#180000',
      '--diff-added':'#00FF41', '--diff-removed':'#FF3300', '--diff-hunk':'#39FF14',
    },
    dracula: {
      '--bg':'#282A36', '--bg-panel':'#21222C', '--bg-hover':'#44475A',
      '--border':'#3D4050', '--border-active':'#BD93F9',
      '--text':'#F8F8F2', '--muted':'#6272A4',
      '--array-idx':'#FF5555', '--string':'#50FA7B', '--number':'#F1FA8C',
      '--bool-true':'#8BE9FD', '--bool-false':'#FF5555', '--null':'#6272A4',
      '--path':'#FF79C6', '--search':'#F1FA8C', '--hash':'#FF79C6', '--stats':'#F1FA8C',
      '--bg-added':'#1e3128', '--bg-removed':'#3a1a1e',
      '--diff-added':'#50FA7B', '--diff-removed':'#FF5555', '--diff-hunk':'#BD93F9',
    },
    nord: {
      '--bg':'#2E3440', '--bg-panel':'#272C36', '--bg-hover':'#3B4252',
      '--border':'#3B4252', '--border-active':'#88C0D0',
      '--text':'#ECEFF4', '--muted':'#4C566A',
      '--array-idx':'#BF616A', '--string':'#A3BE8C', '--number':'#EBCB8B',
      '--bool-true':'#81A1C1', '--bool-false':'#BF616A', '--null':'#4C566A',
      '--path':'#B48EAD', '--search':'#EBCB8B', '--hash':'#B48EAD', '--stats':'#EBCB8B',
      '--bg-added':'#2c3a2c', '--bg-removed':'#3a2c2e',
      '--diff-added':'#A3BE8C', '--diff-removed':'#BF616A', '--diff-hunk':'#88C0D0',
    },
  };

  const LIGHT_THEMES = {
    clean: {
      '--bg':'#F5F7FA', '--bg-panel':'#EAECF0', '--bg-hover':'#DDE3EE',
      '--border':'#C8D0DC', '--border-active':'#007ACC',
      '--text':'#1A2332', '--muted':'#95A5A6',
      '--array-idx':'#C0392B', '--string':'#27882B', '--number':'#B07D00',
      '--bool-true':'#2980B9', '--bool-false':'#C0392B', '--null':'#7F8C8D',
      '--path':'#8E44AD', '--search':'#E67E22', '--hash':'#8E44AD', '--stats':'#E67E22',
      '--bg-added':'#E8F5E9', '--bg-removed':'#FFEBEE',
      '--diff-added':'#27882B', '--diff-removed':'#C0392B', '--diff-hunk':'#007ACC',
    },
    solarized: {
      '--bg':'#FDF6E3', '--bg-panel':'#EEE8D5', '--bg-hover':'#E0D9C5',
      '--border':'#D0C9B5', '--border-active':'#268BD2',
      '--text':'#657B83', '--muted':'#93A1A1',
      '--array-idx':'#DC322F', '--string':'#859900', '--number':'#B58900',
      '--bool-true':'#2AA198', '--bool-false':'#DC322F', '--null':'#93A1A1',
      '--path':'#D33682', '--search':'#CB4B16', '--hash':'#D33682', '--stats':'#CB4B16',
      '--bg-added':'#EBF5EB', '--bg-removed':'#FBE9E9',
      '--diff-added':'#859900', '--diff-removed':'#DC322F', '--diff-hunk':'#268BD2',
    },
  };

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
  });
})();
