'use strict';

const dropzone = document.getElementById('dropzone');
const fileInput = document.getElementById('fileInput');
const empty = document.getElementById('empty');
const panes = document.getElementById('panes');
const beforeImg = document.getElementById('beforeImg');
const afterImg = document.getElementById('afterImg');
const beforeMeta = document.getElementById('beforeMeta');
const afterMeta = document.getElementById('afterMeta');
const quality = document.getElementById('quality');
const qualityOut = document.getElementById('qualityOut');
const scale = document.getElementById('scale');
const scaleOut = document.getElementById('scaleOut');
const format = document.getElementById('format');
const stats = document.getElementById('stats');
const statSaving = document.getElementById('statSaving');
const statSize = document.getElementById('statSize');
const statDims = document.getElementById('statDims');
const download = document.getElementById('download');
const ladderBody = document.getElementById('ladderBody');
const statusText = document.getElementById('status-text');
const hdrInfo = document.getElementById('hdr-info');

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

async function upload(file) {
  statusText.textContent = 'decoding ' + file.name + '…';
  const res = await fetch('/api/upload?name=' + encodeURIComponent(file.name), {
    method: 'POST',
    body: file,
  });
  if (!res.ok) {
    statusText.textContent = 'could not read that file: ' + (await res.text()).trim();
    return;
  }

  current = await res.json();
  beforeImg.src = URL.createObjectURL(file);
  beforeMeta.textContent = current.format + ' · ' + current.width + ' × ' + current.height + ' · ' + humanBytes(current.bytes);
  hdrInfo.textContent = current.name;
  empty.classList.add('hidden');
  panes.classList.remove('hidden');
  stats.classList.remove('hidden');
  download.classList.remove('hidden');

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
  statusText.textContent = 'encoding…';

  const res = await fetch('/api/optimize?' + params());
  if (seq !== pending) return;
  if (!res.ok) {
    statusText.textContent = (await res.text()).trim();
    return;
  }

  const data = await res.json();
  afterImg.src = data.dataUrl;
  afterMeta.textContent = data.format + ' q' + data.quality + ' · ' + data.width + ' × ' + data.height + ' · ' + humanBytes(data.bytes);
  statSaving.textContent = pct(data.saving);
  statSaving.className = data.saving > 0 ? 'win' : 'loss';
  statSize.textContent = humanBytes(data.bytes);
  statDims.textContent = data.width + ' × ' + data.height;
  download.href = data.dataUrl;
  download.download = current.name.replace(/\.[^.]+$/, '') + '-small.' + (data.format === 'jpeg' ? 'jpg' : data.format);
  statusText.textContent = humanBytes(current.bytes) + ' → ' + humanBytes(data.bytes) + '  ' + pct(data.saving);
}

async function loadLadder() {
  if (!current) return;
  const res = await fetch('/api/analyze?' + params());
  if (!res.ok) return;

  const data = await res.json();
  ladderBody.innerHTML = (data.steps || []).map(step => {
    const saving = current.bytes > 0 ? 1 - step.bytes / current.bytes : 0;
    const active = String(step.quality) === quality.value ? ' class="active"' : '';
    return '<tr' + active + ' data-quality="' + step.quality + '">' +
      '<td>' + step.quality + '</td>' +
      '<td>' + humanBytes(step.bytes) + '</td>' +
      '<td class="' + (saving > 0 ? 'win' : 'loss') + '">' + pct(saving) + '</td></tr>';
  }).join('');

  ladderBody.querySelectorAll('tr[data-quality]').forEach(tr => {
    tr.addEventListener('click', () => {
      quality.value = tr.dataset.quality;
      qualityOut.textContent = quality.value;
      refresh();
      loadLadder();
    });
  });
}

quality.addEventListener('input', () => { qualityOut.textContent = quality.value; refresh(); });
quality.addEventListener('change', loadLadder);
scale.addEventListener('input', () => { scaleOut.textContent = scale.value + '%'; refresh(); });
scale.addEventListener('change', loadLadder);
format.addEventListener('change', () => { refresh(); loadLadder(); });

fileInput.addEventListener('change', () => {
  if (fileInput.files.length) upload(fileInput.files[0]);
});

['dragenter', 'dragover'].forEach(ev => dropzone.addEventListener(ev, e => {
  e.preventDefault();
  dropzone.classList.add('dragging');
}));
['dragleave', 'drop'].forEach(ev => dropzone.addEventListener(ev, e => {
  e.preventDefault();
  dropzone.classList.remove('dragging');
}));
dropzone.addEventListener('drop', e => {
  const file = e.dataTransfer.files[0];
  if (file) upload(file);
});

// If unum was launched with a file argument, load it straight away.
(async function loadInitial() {
  const res = await fetch('/api/source');
  if (!res.ok) return;
  const blob = await res.blob();
  const name = res.headers.get('X-Unum-Name') || 'image';
  upload(new File([blob], name, { type: blob.type }));
})();
