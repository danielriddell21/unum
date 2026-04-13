/* unum web UI — vanilla JS, no framework, no build step */

(function () {
  'use strict';

  let treeData = null;
  let selectedNode = null;
  let collapsed = new Set(); // node paths that are collapsed
  let searchQuery = '';
  let typegenMode = 'go'; // go | ts

  // ── Boot ────────────────────────────────────────────────────────────────────

  const keyParam = new URLSearchParams(window.location.search).get('key');

  async function boot() {
    try {
      const treeURL = keyParam ? '/api/tree?key=' + encodeURIComponent(keyParam) : '/api/tree';
      const resp = await fetch(treeURL);
      if (resp.status === 204) { showUploadPanel(); return; }
      if (!resp.ok) throw new Error('Failed to load tree: ' + resp.status);
      treeData = await resp.json();
      renderHeader();
      renderTree();
      renderSidebar();
      setupSearch();
      setupQuery();
      setupKeyboard();
    } catch (e) {
      document.getElementById('tree-panel').textContent = '✗ ' + e.message;
    }
  }

  // ── Header ───────────────────────────────────────────────────────────────────

  function renderHeader() {
    document.getElementById('hdr-filename').textContent = treeData.filename;
    document.getElementById('sep-file').style.display = '';
    document.getElementById('hdr-meta').innerHTML =
      `<span style="color:var(--border-active)">${treeData.nodeCount}</span> nodes · depth <span style="color:var(--border-active)">${treeData.maxDepth}</span> · <span style="color:var(--border-active)">${humanBytes(treeData.sizeBytes)}</span>`;
    document.getElementById('sep-meta').style.display = '';

    if (!document.getElementById('new-btn')) {
      const header = document.getElementById('header');
      const btn = document.createElement('button');
      btn.id = 'new-btn';
      btn.textContent = '[ new ]';
      btn.style.cssText = 'background:none;border:none;color:var(--muted);font-size:11px;' +
        'cursor:pointer;font-family:inherit;transition:color 0.15s;padding:0;margin-left:auto;margin-right:6px;';
      btn.onmouseover = () => { btn.style.color = 'var(--border-active)'; };
      btn.onmouseout  = () => { btn.style.color = 'var(--muted)'; };
      btn.onclick = () => showUploadPanel();
      const toggle = document.getElementById('mode-toggle');
      if (toggle) { toggle.style.marginLeft = '0'; header.insertBefore(btn, toggle); }
      else header.appendChild(btn);
    }
  }

  function humanBytes(b) {
    if (b < 1024) return b + 'B';
    if (b < 1024 * 1024) return (b / 1024).toFixed(1) + 'KB';
    return (b / 1024 / 1024).toFixed(1) + 'MB';
  }

  // ── Tree rendering ───────────────────────────────────────────────────────────

  function renderTree() {
    const panel = document.getElementById('tree-panel');
    panel.innerHTML = '';
    renderNodeEl(treeData.tree, panel, 0, null);
  }

  function renderNodeEl(node, container, depth, parentKey) {
    const el = document.createElement('div');
    el.className = 'tree-node';
    el.dataset.path = node.path || '.';

    // Indent
    for (let i = 0; i < depth; i++) {
      const sp = document.createElement('span');
      sp.className = 'indent';
      el.appendChild(sp);
    }

    // Toggle button (for containers)
    const isContainer = node.kind === 'object' || node.kind === 'array';
    const toggleEl = document.createElement('span');
    toggleEl.className = 'toggle';
    if (isContainer && node.children && node.children.length > 0) {
      const isCollapsed = collapsed.has(node.path);
      toggleEl.textContent = isCollapsed ? '▶' : '▼';
      toggleEl.onclick = (e) => { e.stopPropagation(); toggleCollapse(node); };
    } else {
      toggleEl.textContent = ' ';
    }
    el.appendChild(toggleEl);

    // Key or index
    if (node.key !== undefined && node.key !== null && node.key !== '') {
      const keyEl = document.createElement('span');
      keyEl.className = 'key';
      keyEl.textContent = '"' + node.key + '"';
      if (searchQuery) highlightMatch(keyEl, node.key);
      el.appendChild(keyEl);
      const colon = document.createElement('span');
      colon.className = 'colon';
      colon.textContent = ': ';
      el.appendChild(colon);
    } else if (node.index !== undefined && node.index >= 0) {
      const idxEl = document.createElement('span');
      idxEl.className = 'array-idx';
      idxEl.textContent = '[' + node.index + ']';
      el.appendChild(idxEl);
      const sp = document.createElement('span');
      sp.className = 'colon';
      sp.textContent = ' ';
      el.appendChild(sp);
    }

    // Value
    const isCollapsed = collapsed.has(node.path);
    if (isContainer) {
      const bracketOpen = document.createElement('span');
      bracketOpen.className = 'bracket';
      bracketOpen.textContent = node.kind === 'object' ? '{' : '[';
      el.appendChild(bracketOpen);

      if (isCollapsed || !node.children || node.children.length === 0) {
        const count = document.createElement('span');
        count.className = 'child-count';
        count.textContent = node.children ? node.children.length : 0;
        el.appendChild(count);
        const bracketClose = document.createElement('span');
        bracketClose.className = 'bracket';
        bracketClose.textContent = node.kind === 'object' ? '}' : ']';
        el.appendChild(bracketClose);
      }
    } else {
      const valEl = document.createElement('span');
      valEl.className = valueClass(node.kind, node.raw);
      valEl.textContent = displayValue(node);
      if (searchQuery && node.kind === 'string') highlightMatch(valEl, node.displayValue || '');
      el.appendChild(valEl);
    }

    // Annotation strip
    const strip = buildAnnotationStrip(node);
    if (strip) el.appendChild(strip);

    el.onclick = () => selectNode(node, el);
    container.appendChild(el);

    // Render children
    if (isContainer && !isCollapsed && node.children && node.children.length > 0) {
      // Filter by search
      const visible = searchQuery
        ? node.children.filter(c => nodeMatchesSearch(c))
        : node.children;
      for (const child of visible) {
        renderNodeEl(child, container, depth + 1, node.key);
      }

      // Closing bracket
      const closeEl = document.createElement('div');
      closeEl.className = 'tree-node';
      for (let i = 0; i < depth; i++) {
        const sp = document.createElement('span');
        sp.className = 'indent';
        closeEl.appendChild(sp);
      }
      const sp = document.createElement('span');
      sp.className = 'toggle';
      sp.textContent = ' ';
      closeEl.appendChild(sp);
      const bracketClose = document.createElement('span');
      bracketClose.className = 'bracket';
      bracketClose.textContent = node.kind === 'object' ? '}' : ']';
      closeEl.appendChild(bracketClose);
      container.appendChild(closeEl);
    }
  }

  function toggleCollapse(node) {
    if (collapsed.has(node.path)) {
      collapsed.delete(node.path);
    } else {
      collapsed.add(node.path);
    }
    renderTree();
    updateStatus(selectedNode);
  }

  function selectNode(node, el) {
    document.querySelectorAll('.tree-node.selected').forEach(e => e.classList.remove('selected'));
    el.classList.add('selected');
    selectedNode = node;
    updateStatus(node);

  }

  function updateStatus(node) {
    const pathEl = document.getElementById('status-path');
    const typeEl = document.getElementById('status-type');
    if (!node) {
      pathEl.textContent = '.';
      typeEl.textContent = '';
      return;
    }
    pathEl.textContent = '[ ' + (node.path || '.') + ' ]';
    typeEl.textContent = node.kind.toUpperCase();
    pathEl.onclick = () => {
      navigator.clipboard.writeText(node.path || '.').then(() => showToast('path copied!'));
    };
    pathEl.style.cursor = 'pointer';
    pathEl.title = 'Click to copy path';
  }

  function valueClass(kind, raw) {
    if (kind === 'string') return 'value-string';
    if (kind === 'number') return 'value-number';
    if (kind === 'bool' && raw === 'true') return 'value-bool-true';
    if (kind === 'bool' && raw === 'false') return 'value-bool-false';
    if (kind === 'null') return 'value-null';
    return '';
  }

  function displayValue(node) {
    if (node.kind === 'string') {
      const v = node.displayValue || '';
      return v.length > 60 ? '"' + v.slice(0, 57) + '…"' : '"' + v + '"';
    }
    return node.raw || 'null';
  }

  function buildAnnotationStrip(node) {
    const parts = [];
    if (node.merkleHash) {
      parts.push('<span class="hash">#' + node.merkleHash.slice(0, 8) + '</span>');
    }
    if (node.stats && node.stats.numericCount > 0) {
      const s = node.stats;
      parts.push('<span class="stat">n=' + s.count + ' min=' + fmt(s.min) + ' max=' + fmt(s.max) + ' mean=' + fmt(s.mean) + '</span>');
    }
    if (!parts.length) return null;
    const strip = document.createElement('div');
    strip.className = 'annotation-strip';
    strip.innerHTML = parts.join('  ');
    return strip;
  }

  function fmt(v) {
    if (v === undefined || v === null) return '';
    return parseFloat(v.toPrecision(4)).toString();
  }

  // ── Search ───────────────────────────────────────────────────────────────────

  function setupSearch() {
    const input = document.getElementById('search-input');
    input.addEventListener('input', () => {
      searchQuery = input.value.trim().toLowerCase();
      renderTree();
    });
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') { input.value = ''; searchQuery = ''; renderTree(); }
    });
  }

  function nodeMatchesSearch(node) {
    const q = searchQuery;
    if (!q) return true;
    if ((node.key || '').toLowerCase().includes(q)) return true;
    if ((node.raw || '').toLowerCase().includes(q)) return true;
    if ((node.displayValue || '').toLowerCase().includes(q)) return true;
    if (node.children) return node.children.some(c => nodeMatchesSearch(c));
    return false;
  }

  function highlightMatch(el, text) {
    const q = searchQuery;
    if (!q || !text) return;
    const idx = text.toLowerCase().indexOf(q);
    if (idx < 0) return;
    const before = text.slice(0, idx);
    const match = text.slice(idx, idx + q.length);
    const after = text.slice(idx + q.length);
    el.innerHTML = escapeHTML(before) + '<span class="search-match">' + escapeHTML(match) + '</span>' + escapeHTML(after);
  }

  function escapeHTML(s) {
    return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
  }

  // ── Sidebar ──────────────────────────────────────────────────────────────────

  function renderSidebar() {
    const content = document.getElementById('sidebar-content');
    content.innerHTML = '';

    const sections = [
      { id: 'typegen', label: 'TYPE GENERATION', render: renderTypeGenSection },
      { id: 'schema',  label: 'JSON SCHEMA',      render: renderSchemaSection },
      { id: 'yaml',    label: 'YAML',             render: renderYAMLSection },
      { id: 'merkle',  label: 'MERKLE ROOT',       render: renderMerkleSection },
    ];

    for (const s of sections) {
      const section = document.createElement('div');
      section.className = 'sidebar-section';
      section.id = 'section-' + s.id;

      const header = document.createElement('div');
      header.className = 'sidebar-section-header';
      header.innerHTML = '<span class="arrow">▼</span><span class="label">' + s.label + '</span>';
      header.onclick = () => {
        const body = section.querySelector('.sidebar-section-body');
        const isCollapsed = body.classList.contains('collapsed');
        body.classList.toggle('collapsed', !isCollapsed);
        header.querySelector('.arrow').textContent = isCollapsed ? '▼' : '▶';
      };

      const body = document.createElement('div');
      body.className = 'sidebar-section-body';
      s.render(body, selectedNode);

      section.appendChild(header);
      section.appendChild(body);
      content.appendChild(section);
    }
  }

  function renderTypeGenSection(body, node) {
    // Sub-mode tabs
    body.innerHTML = '';
    const tabs = document.createElement('div');
    tabs.className = 'sub-tabs';
    for (const [mode, label] of [['go','Go'], ['ts','TypeScript']]) {
      const btn = document.createElement('button');
      btn.className = 'sub-tab' + (mode === typegenMode ? ' active' : '');
      btn.textContent = label;
      btn.onclick = () => { typegenMode = mode; renderTypeGenSection(body, node); };
      tabs.appendChild(btn);
    }
    body.appendChild(tabs);

    const code = document.createElement('div');
    code.textContent = treeData.typegen ? (treeData.typegen[typegenMode] || '') : '';
    body.appendChild(code);
  }

  function renderSchemaSection(body) {
    body.textContent = treeData.typegen ? (treeData.typegen['jsonschema'] || '') : '';
  }

  function renderYAMLSection(body) {
    body.textContent = treeData.yaml || '';
  }

  function renderMerkleSection(body) {
    const root = treeData.merkleRoot || '';
    body.innerHTML = root
      ? '<span class="syn-hash">root: ' + root + '</span>'
      : '<span style="color:var(--muted)">Merkle hashing not enabled.\nRun with: unum json &lt;file&gt; --web --merkle</span>';
  }

  // ── jq query ─────────────────────────────────────────────────────────────────

  function setupQuery() {
    const input = document.getElementById('query-input');
    const runBtn = document.getElementById('query-run');

    const run = async () => {
      const expr = input.value.trim();
      if (!expr) return;
      try {
        const queryURL = keyParam ? '/api/query?key=' + encodeURIComponent(keyParam) : '/api/query';
        const resp = await fetch(queryURL, {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({expr}),
        });
        const data = await resp.json();
        if (data.error) {
          showQueryResult('✗ ' + data.error);
        } else {
          showQueryResult(data.result);
        }
      } catch (e) {
        showQueryResult('✗ ' + e.message);
      }
    };

    runBtn.addEventListener('click', run);
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') run();
    });
  }

  function showQueryResult(text) {
    // Show in a small overlay or inject into the tree panel area
    let overlay = document.getElementById('query-result');
    if (!overlay) {
      overlay = document.createElement('div');
      overlay.id = 'query-result';
      overlay.style.cssText = [
        'position:fixed', 'bottom:32px', 'left:50%', 'transform:translateX(-50%)',
        'background:var(--bg-panel)', 'border:1px solid var(--path)',
        'padding:12px 16px', 'font-size:12px', 'max-width:600px', 'max-height:300px',
        'overflow:auto', 'white-space:pre', 'z-index:200', 'border-radius:4px',
        'box-shadow:0 0 20px rgba(198,120,221,0.2)',
      ].join(';');
      document.body.appendChild(overlay);
    }
    overlay.textContent = text;
    overlay.style.display = 'block';
    setTimeout(() => { overlay.style.display = 'none'; }, 10000);
  }

  // ── Keyboard shortcuts ────────────────────────────────────────────────────────

  function setupKeyboard() {
    document.addEventListener('keydown', (e) => {
      // Ignore when typing in inputs
      if (e.target.tagName === 'INPUT') return;

      if (e.key === '/') {
        e.preventDefault();
        document.getElementById('search-input').focus();
      }
      if (e.key === 'q') {
        showUploadPanel();
      }
    });
  }

  // ── Toast ─────────────────────────────────────────────────────────────────────

  function showToast(msg) {
    const t = document.getElementById('toast');
    t.textContent = msg;
    t.style.display = 'block';
    setTimeout(() => { t.style.display = 'none'; }, 2000);
  }

  // ── Upload panel ─────────────────────────────────────────────────────────────

  function showUploadPanel() {
    document.getElementById('hdr-filename').textContent = '';
    document.getElementById('sep-file').style.display = 'none';
    document.getElementById('hdr-meta').innerHTML = '';
    document.getElementById('sep-meta').style.display = 'none';
    document.getElementById('sidebar').style.display = 'none';
    document.getElementById('search-bar').style.display = 'none';
    document.getElementById('status-hints').textContent = 'drop or paste a .json file';
    document.getElementById('status-path').textContent = '.';
    document.getElementById('status-type').textContent = '';
    selectedNode = null;

    const panel = document.getElementById('tree-panel');
    panel.innerHTML = `
      <div id="upload-panel" style="display:flex;flex-direction:column;align-items:center;justify-content:center;height:100%;padding:32px;">
        <div id="json-drop-zone" style="border:2px dashed var(--border-active);border-radius:6px;padding:56px 80px;text-align:center;cursor:pointer;transition:background 0.15s;max-width:480px;width:100%;">
          <div style="color:var(--border-active);font-size:14px;font-weight:bold;margin-bottom:8px;">drop .json file here</div>
          <div style="color:var(--muted);font-size:12px;">or click to browse</div>
          <input id="json-file-input" type="file" accept=".json,application/json" style="display:none" />
        </div>
        <button id="json-paste-btn" style="display:block;margin:16px auto 0;background:transparent;border:1px solid var(--border-active);border-radius:4px;color:var(--border-active);cursor:pointer;font-family:inherit;font-size:12px;padding:6px 18px;transition:background 0.15s;">paste from clipboard</button>
        <div id="json-upload-error" style="color:#E06C75;font-size:12px;margin-top:12px;"></div>
      </div>
    `;

    const dropZone = document.getElementById('json-drop-zone');
    const fileInput = document.getElementById('json-file-input');
    const errEl = document.getElementById('json-upload-error');

    dropZone.addEventListener('click', () => fileInput.click());
    dropZone.addEventListener('dragover', (e) => { e.preventDefault(); dropZone.style.background = 'var(--bg-hover)'; });
    dropZone.addEventListener('dragleave', () => { dropZone.style.background = ''; });
    dropZone.addEventListener('drop', (e) => {
      e.preventDefault();
      dropZone.style.background = '';
      const file = e.dataTransfer.files[0];
      if (file) loadUploadFile(file, errEl);
    });
    fileInput.addEventListener('change', () => {
      const file = fileInput.files[0];
      if (file) loadUploadFile(file, errEl);
    });
    document.getElementById('json-paste-btn').addEventListener('click', () => pasteUpload(errEl));
  }

  function loadUploadFile(file, errEl) {
    const reader = new FileReader();
    reader.onload = async (e) => {
      try { await uploadAndLoad(e.target.result, file.name); }
      catch (err) { errEl.textContent = '✗ ' + err.message; }
    };
    reader.onerror = () => { errEl.textContent = '✗ failed to read file'; };
    reader.readAsText(file);
  }

  async function pasteUpload(errEl) {
    errEl.textContent = '';
    try {
      const text = await navigator.clipboard.readText();
      if (!text.trim()) { errEl.textContent = '✗ clipboard is empty'; return; }
      await uploadAndLoad(text, '(clipboard)');
    } catch (e) {
      errEl.textContent = '✗ ' + (e.name === 'NotAllowedError' ? 'clipboard access denied' : e.message);
    }
  }

  async function uploadAndLoad(content, filename) {
    const resp = await fetch('/api/upload', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ filename: filename, content: content }),
    });
    if (!resp.ok) {
      const msg = await resp.text();
      throw new Error(msg.trim() || resp.statusText);
    }
    const { key } = await resp.json();
    const treeResp = await fetch('/api/tree?key=' + encodeURIComponent(key));
    if (!treeResp.ok) throw new Error('Failed to load tree: ' + treeResp.status);
    treeData = await treeResp.json();
    document.getElementById('sidebar').style.display = '';
    document.getElementById('search-bar').style.display = '';
    renderHeader();
    renderTree();
    renderSidebar();
  }

  // ── Init ──────────────────────────────────────────────────────────────────────

  document.addEventListener('DOMContentLoaded', boot);
})();
