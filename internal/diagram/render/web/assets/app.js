'use strict';

const source = document.getElementById('source');
const langSel = document.getElementById('lang');
const canvas = document.getElementById('canvas');
const errorBox = document.getElementById('error');
const statusLang = document.getElementById('status-lang');
const statusSize = document.getElementById('status-size');

const cfg = window.RENDER_CONFIG || { lang: 'd2', source: '' };
source.value = cfg.source || '';
if (cfg.lang === 'mermaid' || cfg.lang === 'd2') langSel.value = cfg.lang;

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
    statusSize.textContent = Math.round(vb.width) + ' × ' + Math.round(vb.height);
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

source.addEventListener('input', debounceRender);
langSel.addEventListener('change', render);
document.querySelectorAll('.downloads button').forEach(btn => {
  btn.addEventListener('click', () => download(btn.dataset.fmt));
});

// Re-render the diagram in the matching palette when the light/dark toggle flips.
document.addEventListener('DOMContentLoaded', () => {
  const toggle = document.getElementById('mode-toggle');
  if (toggle) toggle.addEventListener('click', () => setTimeout(render, 0));
});

render();
