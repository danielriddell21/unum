'use strict';

const input = document.getElementById('hashInput');
const btn = document.getElementById('deriveBtn');
const results = document.getElementById('results');
const resultBody = document.getElementById('resultBody');
const emptyHint = document.getElementById('empty');
const historyList = document.getElementById('historyList');

async function derive(text) {
  if (!text.trim()) return;
  const res = await fetch('/api/derive?input=' + encodeURIComponent(text));
  if (!res.ok) return;
  const data = await res.json();
  renderResult(data);
  await loadHistory(text);
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

async function loadHistory(activeInput) {
  const res = await fetch('/api/history');
  if (!res.ok) return;
  const entries = await res.json();

  if (!entries || entries.length === 0) {
    historyList.innerHTML = '<li class="no-history">no history yet</li>';
    return;
  }

  // HistoryEntry JSON tags are lowercase: { input, time }
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
