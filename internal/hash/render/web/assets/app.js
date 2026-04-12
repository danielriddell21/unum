'use strict';

// ── Dark / Light mode ────────────────────────────────────────────────────────

const _cfg    = window.UNUM_CONFIG || {};
const _LS_KEY = 'unum-mode';

const DARK_THEMES = {
  cyber: {
    '--bg':'#0D0D0D', '--bg-panel':'#111111', '--bg-hover':'#1A1A2E',
    '--border':'#1E1E1E', '--border-active':'#00D4FF',
    '--text':'#C0C0C0', '--muted':'#3A3A3A',
  },
  matrix: {
    '--bg':'#0D0D0D', '--bg-panel':'#0A1A0A', '--bg-hover':'#001A00',
    '--border':'#003300', '--border-active':'#00FF41',
    '--text':'#00CC22', '--muted':'#005500',
  },
  dracula: {
    '--bg':'#282A36', '--bg-panel':'#21222C', '--bg-hover':'#44475A',
    '--border':'#3D4050', '--border-active':'#BD93F9',
    '--text':'#F8F8F2', '--muted':'#6272A4',
  },
  nord: {
    '--bg':'#2E3440', '--bg-panel':'#272C36', '--bg-hover':'#3B4252',
    '--border':'#3B4252', '--border-active':'#88C0D0',
    '--text':'#ECEFF4', '--muted':'#4C566A',
  },
};

const LIGHT_THEMES = {
  clean: {
    '--bg':'#F5F7FA', '--bg-panel':'#EAECF0', '--bg-hover':'#DDE3EE',
    '--border':'#C8D0DC', '--border-active':'#007ACC',
    '--text':'#1A2332', '--muted':'#95A5A6',
  },
  solarized: {
    '--bg':'#FDF6E3', '--bg-panel':'#EEE8D5', '--bg-hover':'#E0D9C5',
    '--border':'#D0C9B5', '--border-active':'#268BD2',
    '--text':'#657B83', '--muted':'#93A1A1',
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

const input = document.getElementById('hashInput');
const btn = document.getElementById('deriveBtn');
const results = document.getElementById('results');
const resultBody = document.getElementById('resultBody');
const emptyHint = document.getElementById('empty');
const historyList = document.getElementById('historyList');

const MAX_HISTORY = 50;
const STORAGE_KEY = 'unum-hash-history';

async function derive(text) {
  if (!text.trim()) return;
  const res = await fetch('/api/derive?input=' + encodeURIComponent(text));
  if (!res.ok) return;
  const data = await res.json();
  renderResult(data);
  appendHistory(text);
  loadHistory(text);
}

function appendHistory(val) {
  let entries = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
  entries = entries.filter(e => e.input !== val);
  entries.unshift({ input: val, time: new Date().toISOString() });
  if (entries.length > MAX_HISTORY) entries = entries.slice(0, MAX_HISTORY);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(entries));
}

function renderResult(data) {
  emptyHint.style.display = 'none';
  results.classList.remove('hidden');

  const rows = [
    { label: 'input',  value: data.Input },
    { label: 'port',   value: String(data.Port) },
    { label: 'uuid',   value: data.UUID },
    { label: 'color',  value: data.Color, isColor: true },
    { label: 'short',  value: data.Short },
    { label: 'emoji',  value: data.Emoji },
    { label: 'phrase', value: data.Phrase },
  ];

  resultBody.innerHTML = rows.map(r => {
    let valCell;
    if (r.isColor) {
      valCell = `<td class="value"><span class="color-swatch" style="background:${r.value}"></span>${r.value}</td>`;
    } else {
      valCell = `<td class="value">${escHtml(r.value)}</td>`;
    }
    return `<tr><td class="label">${r.label}</td>${valCell}</tr>`;
  }).join('');
}

function loadHistory(activeInput) {
  const entries = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');

  if (!entries || entries.length === 0) {
    historyList.innerHTML = '<li class="no-history">no history yet</li>';
    return;
  }

  historyList.innerHTML = entries.map(e => {
    const cls = (activeInput !== undefined && e.input === activeInput) ? ' class="active"' : '';
    return `<li${cls} data-input="${escAttr(e.input)}" title="${escAttr(e.input)}">${escHtml(e.input)}</li>`;
  }).join('');

  historyList.querySelectorAll('li[data-input]').forEach(li => {
    li.addEventListener('click', () => {
      input.value = li.dataset.input;
      derive(li.dataset.input);
    });
  });
}

function escHtml(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

function escAttr(s) {
  return String(s).replace(/"/g, '&quot;');
}

// Derive on Enter key
input.addEventListener('keydown', e => {
  if (e.key === 'Enter') derive(input.value);
});

// Derive button click
btn.addEventListener('click', () => derive(input.value));

// Load existing history on page open
loadHistory();

// Apply dark/light mode and add toggle button
applyMode(resolveMode());
const _hdr = document.getElementById('header');
if (_hdr) _hdr.appendChild(renderModeToggle());
