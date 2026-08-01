'use strict';

const uploadPanel = document.getElementById('upload-panel');
const layout = document.getElementById('layout');
const pasteArea = document.getElementById('paste-area');
const pasteFilename = document.getElementById('paste-filename');
const browseBtn = document.getElementById('browse-btn');
const fileInput = document.getElementById('file-input');
const upLang = document.getElementById('up-lang');
const upSubmit = document.getElementById('up-submit');
const upError = document.getElementById('up-error');

const source = document.getElementById('source');
const langSel = document.getElementById('lang');
const canvas = document.getElementById('canvas');
const errorBox = document.getElementById('error');
const statusLang = document.getElementById('status-lang');
const statusHints = document.getElementById('status-hints');
const hdrFilename = document.getElementById('hdr-filename');
const hdrMeta = document.getElementById('hdr-meta');
const sepFile = document.getElementById('sep-file');
const sepMeta = document.getElementById('sep-meta');

function setHeaderFile(name) {
  hdrFilename.textContent = name || '';
  sepFile.style.display = name ? '' : 'none';
}

// Built from DOM nodes rather than innerHTML so the header can never become an
// injection sink, with the numbers accented the way the sibling tools do.
function setHeaderMeta(w, h) {
  hdrMeta.replaceChildren();
  if (!w || !h) {
    sepMeta.style.display = 'none';
    return;
  }
  const accent = (v) => {
    const s = document.createElement('span');
    s.style.color = 'var(--border-active)';
    s.textContent = v;
    return s;
  };
  hdrMeta.append(accent(w), document.createTextNode(' × '), accent(h));
  sepMeta.style.display = '';
}

const cfg = window.DIAGRAM_CONFIG || { lang: 'd2', source: '' };

let timer = null;

function debounceRender() {
  clearTimeout(timer);
  timer = setTimeout(render, 250);
}

function currentTheme() {
  const c = window.UNUM_CONFIG || {};
  const mode = window._unumResolveMode?.() || 'dark';
  return mode === 'light' ? (c.lightTheme || 'clean') : (c.darkTheme || 'cyber');
}

async function render() {
  const lang = langSel.value;
  statusLang.textContent = '[ ' + lang + ' ]';
  if (!source.value.trim()) {
    // Nothing loaded yet (started with no file): leave the canvas empty rather
    // than rendering a blank diagram or surfacing a parse error.
    errorBox.classList.add('hidden');
    canvas.replaceChildren();
    setHeaderMeta(0, 0);
    return;
  }
  const res = await fetch('/api/render?lang=' + lang + '&format=svg&theme=' + encodeURIComponent(currentTheme()), {
    method: 'POST',
    body: source.value,
  });
  if (!res.ok) {
    const msg = await res.text();
    errorBox.textContent = msg;
    errorBox.classList.remove('hidden');
    return;
  }
  errorBox.classList.add('hidden');
  const svg = await res.text();
  // Parse the SVG into a document and append the node rather than assigning
  // innerHTML, so untrusted markup can never execute as script.
  const parsed = new DOMParser().parseFromString(svg, 'image/svg+xml');
  const el = parsed.querySelector('svg');
  if (!el) {
    canvas.replaceChildren();
    return;
  }
  const node = document.importNode(el, true);
  canvas.replaceChildren(node);
  const vb = node.viewBox?.baseVal;
  if (vb?.width) {
    // d2 SVGs carry only a viewBox; give them an intrinsic size so the
    // preview doesn't collapse, then CSS scales it to fit.
    if (!node.getAttribute('width')) {
      node.setAttribute('width', vb.width);
      node.setAttribute('height', vb.height);
    }
    setHeaderMeta(Math.round(vb.width), Math.round(vb.height));
  }
}

async function download(format) {
  const lang = langSel.value;
  const res = await fetch('/api/render?lang=' + lang + '&format=' + format + '&theme=' + encodeURIComponent(currentTheme()), {
    method: 'POST',
    body: source.value,
  });
  if (!res.ok) return;
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  const ext = format === 'drawio' ? 'drawio' : format;
  a.href = url;
  a.download = 'diagram.' + ext;
  a.click();
  URL.revokeObjectURL(url);
}

function langFromName(name) {
  const lower = name.toLowerCase();
  if (lower.endsWith('.mmd') || lower.endsWith('.mermaid')) return 'mermaid';
  if (lower.endsWith('.d2')) return 'd2';
  return '';
}

// ── Upload screen ─────────────────────────────────────────────────────────────

function showUploadPanel() {
  layout.style.display = 'none';
  uploadPanel.style.display = '';
  setHeaderFile('');
  setHeaderMeta(0, 0);
  statusLang.textContent = '';
  statusHints.textContent = 'drop, paste, or browse a .d2 / .mmd file';
  pasteArea.value = '';
  pasteArea._fileName = undefined;
  pasteArea._fileContent = undefined;
  pasteFilename.textContent = '';
  fileInput.value = '';
  upError.textContent = '';
  pasteArea.focus();
}

function enterEditor(src, lang, filename) {
  uploadPanel.style.display = 'none';
  layout.style.display = '';
  statusHints.textContent = 'edit source · live preview · download svg / png / drawio';
  source.value = src || '';
  if (lang === 'mermaid' || lang === 'd2') langSel.value = lang;
  setHeaderFile(filename || '');
  render();
  source.focus();
}

function loadUploadFile(file) {
  const reader = new FileReader();
  reader.onload = (e) => {
    pasteArea.value = e.target.result;
    pasteArea._fileName = file.name;
    pasteArea._fileContent = e.target.result;
    pasteFilename.textContent = file.name;
    const lang = langFromName(file.name);
    if (lang) upLang.value = lang;
  };
  reader.onerror = () => { upError.textContent = '✗ failed to read ' + file.name; };
  reader.readAsText(file);
}

browseBtn.addEventListener('click', () => fileInput.click());
fileInput.addEventListener('change', () => {
  const file = fileInput.files[0];
  if (file) loadUploadFile(file);
});
pasteArea.addEventListener('dragover', (e) => {
  e.preventDefault();
  pasteArea.classList.add('drag-over');
});
pasteArea.addEventListener('dragleave', () => pasteArea.classList.remove('drag-over'));
pasteArea.addEventListener('drop', (e) => {
  e.preventDefault();
  pasteArea.classList.remove('drag-over');
  const file = e.dataTransfer.files[0];
  if (file) loadUploadFile(file);
});
upSubmit.addEventListener('click', () => {
  upError.textContent = '';
  if (!pasteArea.value.trim()) { upError.textContent = '✗ nothing to render'; return; }
  const name = (pasteArea._fileContent !== undefined && pasteArea.value === pasteArea._fileContent && pasteArea._fileName)
    ? pasteArea._fileName : '';
  enterEditor(pasteArea.value, upLang.value, name);
});

// ── Editor ────────────────────────────────────────────────────────────────────

source.addEventListener('input', debounceRender);
langSel.addEventListener('change', render);
document.querySelectorAll('.downloads button').forEach(btn => {
  btn.addEventListener('click', () => download(btn.dataset.fmt));
});

// Clear back to the upload screen so another diagram can be loaded, matching the
// [ new ] affordance in the json and diff web UIs.
function addNewButton() {
  if (document.getElementById('new-btn')) return;
  const header = document.getElementById('header');
  const btn = document.createElement('button');
  btn.id = 'new-btn';
  btn.textContent = '[ new ]';
  btn.onclick = showUploadPanel;
  const toggle = document.getElementById('mode-toggle');
  if (toggle) {
    toggle.style.marginLeft = '0';
    header.insertBefore(btn, toggle);
  } else {
    header.appendChild(btn);
  }
}

// Re-render the diagram in the matching palette when the light/dark toggle flips.
document.addEventListener('DOMContentLoaded', () => {
  addNewButton();
  const toggle = document.getElementById('mode-toggle');
  if (toggle) toggle.addEventListener('click', () => {
    if (layout.style.display !== 'none') setTimeout(render, 0);
  });
});

// Boot: open straight into the editor when launched with a file, otherwise show
// the upload screen the same way the sibling tools do.
if ((cfg.source || '').trim()) {
  enterEditor(cfg.source, cfg.lang, cfg.file);
} else {
  showUploadPanel();
}
