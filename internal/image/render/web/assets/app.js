'use strict';

const uploadPanel = document.getElementById('upload-panel');
const layout = document.getElementById('layout');
const dropArea = document.getElementById('drop-area');
const pasteFilename = document.getElementById('paste-filename');
const browseBtn = document.getElementById('browse-btn');
const fileInput = document.getElementById('file-input');
const upError = document.getElementById('up-error');

const beforeImg = document.getElementById('beforeImg');
const afterImg = document.getElementById('afterImg');
const beforeMeta = document.getElementById('beforeMeta');
const afterMeta = document.getElementById('afterMeta');
const quality = document.getElementById('quality');
const qualityOut = document.getElementById('qualityOut');
const scale = document.getElementById('scale');
const scaleOut = document.getElementById('scaleOut');
const format = document.getElementById('format');
const statSaving = document.getElementById('statSaving');
const statSize = document.getElementById('statSize');
const statDims = document.getElementById('statDims');
const download = document.getElementById('download');
const ladderBody = document.getElementById('ladderBody');
const statusInfo = document.getElementById('status-info');
const statusHints = document.getElementById('status-hints');

const HINT_UPLOAD = 'drop, paste, or browse an image';
const HINT_LOADED = 'drag the sliders · click a ladder row · download to save';

let current = null;
let pending = 0;

function humanBytes(n) {
  if (n < 1000) return n + ' B';
  const units = ['kB', 'MB', 'GB'];
  let v = n, i = -1;
  while (v >= 1000 && i < units.length - 1) { v /= 1000; i++; }
  if (v >= 999.5 && i < units.length - 1) { v /= 1000; i++; }
  return (v < 10 ? v.toFixed(1) : Math.round(v)) + ' ' + units[i];
}

function pct(saving) {
  return (saving > 0 ? '−' : '+') + Math.abs(Math.round(saving * 100)) + '%';
}

function setHeaderFile(name) {
  document.getElementById('hdr-filename').textContent = name || '';
  document.getElementById('sep-file').style.display = name ? '' : 'none';
}

// ── Upload screen ─────────────────────────────────────────────────────────────

function showUploadPanel() {
  layout.style.display = 'none';
  uploadPanel.style.display = '';
  setHeaderFile('');
  statusInfo.textContent = '';
  statusHints.textContent = HINT_UPLOAD;
  pasteFilename.textContent = '';
  fileInput.value = '';
  upError.textContent = '';
  current = null;
}

function enterViewer() {
  uploadPanel.style.display = 'none';
  layout.style.display = '';
  statusHints.textContent = HINT_LOADED;
  addNewButton();
}

async function upload(file) {
  if (!file) return;
  upError.textContent = '';
  pasteFilename.textContent = file.name;
  statusInfo.textContent = 'decoding ' + file.name + '…';

  let res;
  try {
    res = await fetch('/api/upload?name=' + encodeURIComponent(file.name), {
      method: 'POST',
      body: file,
    });
  } catch (err) {
    upError.textContent = '✗ ' + err.message;
    statusInfo.textContent = '';
    return;
  }
  if (!res.ok) {
    upError.textContent = '✗ ' + (await res.text()).trim();
    statusInfo.textContent = '';
    return;
  }

  current = await res.json();
  // Ask the server for the bytes it just took, rather than making a same-origin
  // blob URL out of the File the user picked.
  beforeImg.src = '/api/original?id=' + encodeURIComponent(current.id);
  beforeMeta.textContent = current.format + ' · ' + current.width + ' × ' + current.height +
    ' · ' + humanBytes(current.bytes);
  setHeaderFile(current.name);
  enterViewer();

  refresh();
  loadLadder();
}

function params() {
  const p = new URLSearchParams({
    id: current.id,
    quality: quality.value,
    scale: scale.value + '%',
  });
  if (format.value) p.set('format', format.value);
  return p.toString();
}

async function refresh() {
  if (!current) return;
  const seq = ++pending;
  statusInfo.textContent = 'encoding…';

  const res = await fetch('/api/optimize?' + params());
  if (seq !== pending) return;
  if (!res.ok) {
    statusInfo.textContent = (await res.text()).trim();
    return;
  }

  const data = await res.json();
  afterImg.src = data.dataUrl;
  afterMeta.textContent = data.format + ' q' + data.quality + ' · ' + data.width + ' × ' +
    data.height + ' · ' + humanBytes(data.bytes);
  statSaving.textContent = pct(data.saving);
  statSaving.className = data.saving > 0 ? 'win' : 'loss';
  statSize.textContent = humanBytes(data.bytes);
  statDims.textContent = data.width + ' × ' + data.height;
  download.href = data.dataUrl;
  download.download = current.name.replace(/\.[^.]+$/, '') + '-small.' +
    (data.format === 'jpeg' ? 'jpg' : data.format);
  statusInfo.textContent = humanBytes(current.bytes) + ' → ' + humanBytes(data.bytes) +
    '  ' + pct(data.saving);
}

async function loadLadder() {
  if (!current) return;
  const res = await fetch('/api/analyze?' + params());
  if (!res.ok) return;

  const data = await res.json();
  ladderBody.replaceChildren();

  for (const step of data.steps || []) {
    ladderBody.appendChild(ladderRow(step));
  }
}

// Rows are built as DOM nodes rather than an HTML string, and each one closes
// over its own quality rather than stashing it in a data- attribute to read back
// out. Nothing is written as markup and nothing is read back from the DOM, so
// there is no way for a value to be reinterpreted as HTML on the next render.
function ladderRow(step) {
  const saving = current.bytes > 0 ? 1 - step.bytes / current.bytes : 0;

  const tr = document.createElement('tr');
  if (String(step.quality) === quality.value) tr.className = 'active';

  tr.appendChild(cell(step.quality));
  tr.appendChild(cell(humanBytes(step.bytes)));
  tr.appendChild(cell(pct(saving), saving > 0 ? 'win' : 'loss'));

  tr.addEventListener('click', () => {
    quality.value = step.quality;
    qualityOut.textContent = quality.value;
    refresh();
    loadLadder();
  });
  return tr;
}

function cell(text, className) {
  const td = document.createElement('td');
  td.textContent = text;
  if (className) td.className = className;
  return td;
}

// Clear back to the upload screen so another image can be loaded, matching the
// [ new ] affordance in the json, diff and diagram web UIs.
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

// ── Input ─────────────────────────────────────────────────────────────────────

browseBtn.addEventListener('click', () => fileInput.click());
dropArea.addEventListener('click', () => fileInput.click());
dropArea.addEventListener('keydown', e => {
  if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); fileInput.click(); }
});
fileInput.addEventListener('change', () => {
  if (fileInput.files.length) upload(fileInput.files[0]);
});

dropArea.addEventListener('dragover', e => {
  e.preventDefault();
  dropArea.classList.add('drag-over');
});
dropArea.addEventListener('dragleave', () => dropArea.classList.remove('drag-over'));

// Dropping anywhere in the window works too, so an image can be swapped without
// going back to the upload screen first.
['dragenter', 'dragover'].forEach(ev =>
  window.addEventListener(ev, e => e.preventDefault()));
window.addEventListener('drop', e => {
  e.preventDefault();
  dropArea.classList.remove('drag-over');
  const file = e.dataTransfer && e.dataTransfer.files[0];
  if (file) upload(file);
});

// An image on the clipboard arrives as a file on the paste event rather than as
// text, which is why the drop target is not a textarea.
window.addEventListener('paste', e => {
  const items = (e.clipboardData && e.clipboardData.files) || [];
  for (const file of items) {
    if (file.type.startsWith('image/')) { upload(file); return; }
  }
});

quality.addEventListener('input', () => { qualityOut.textContent = quality.value; refresh(); });
quality.addEventListener('change', loadLadder);
scale.addEventListener('input', () => { scaleOut.textContent = scale.value + '%'; refresh(); });
scale.addEventListener('change', loadLadder);
format.addEventListener('change', () => { refresh(); loadLadder(); });

// Boot: open straight into the optimizer when launched with a file, otherwise
// show the upload screen the same way the sibling tools do.
(async function boot() {
  showUploadPanel();
  const res = await fetch('/api/source');
  if (!res.ok) return;
  const blob = await res.blob();
  const name = res.headers.get('X-Unum-Name') || 'image';
  upload(new File([blob], name, { type: blob.type }));
})();
