/* unum web UI — vanilla JS, no framework, no build step */

(function () {
  'use strict';

  let treeData = null;
  let selectedNode = null;
  let collapsed = new Set(); // node paths that are collapsed
  let searchQuery = '';
  let typegenMode = 'go'; // go | ts | jsonschema

  // ── Dark / Light mode ────────────────────────────────────────────────────────

  const _cfg    = window.UNUM_CONFIG || {};
  const _LS_KEY = 'unum-mode';

  const DARK_THEMES = {
    cyber: {
      '--bg':'#0D0D0D', '--bg-panel':'#111111', '--bg-hover':'#1A1A2E',
      '--border':'#1E1E1E', '--border-active':'#00D4FF', '--key':'#00D4FF',
      '--array-idx':'#FF6B6B', '--string':'#98C379', '--number':'#E5C07B',
      '--bool-true':'#56B6C2', '--bool-false':'#E06C75', '--null':'#5C6370',
      '--path':'#C678DD', '--search':'#FFD700', '--muted':'#3A3A3A',
      '--text':'#C0C0C0', '--hash':'#C678DD', '--stats':'#FFD700',
    },
    matrix: {
      '--bg':'#0D0D0D', '--bg-panel':'#0A1A0A', '--bg-hover':'#001A00',
      '--border':'#003300', '--border-active':'#00FF41', '--key':'#00FF41',
      '--array-idx':'#33FF33', '--string':'#00CC22', '--number':'#88FF44',
      '--bool-true':'#00FF41', '--bool-false':'#FF3300', '--null':'#005500',
      '--path':'#39FF14', '--search':'#FFFFFF', '--muted':'#005500',
      '--text':'#00CC22', '--hash':'#39FF14', '--stats':'#88FF44',
    },
    dracula: {
      '--bg':'#282A36', '--bg-panel':'#21222C', '--bg-hover':'#44475A',
      '--border':'#3D4050', '--border-active':'#BD93F9', '--key':'#BD93F9',
      '--array-idx':'#FF5555', '--string':'#50FA7B', '--number':'#F1FA8C',
      '--bool-true':'#8BE9FD', '--bool-false':'#FF5555', '--null':'#6272A4',
      '--path':'#FF79C6', '--search':'#F1FA8C', '--muted':'#6272A4',
      '--text':'#F8F8F2', '--hash':'#FF79C6', '--stats':'#F1FA8C',
    },
    nord: {
      '--bg':'#2E3440', '--bg-panel':'#272C36', '--bg-hover':'#3B4252',
      '--border':'#3B4252', '--border-active':'#88C0D0', '--key':'#88C0D0',
      '--array-idx':'#BF616A', '--string':'#A3BE8C', '--number':'#EBCB8B',
      '--bool-true':'#81A1C1', '--bool-false':'#BF616A', '--null':'#4C566A',
      '--path':'#B48EAD', '--search':'#EBCB8B', '--muted':'#4C566A',
      '--text':'#ECEFF4', '--hash':'#B48EAD', '--stats':'#EBCB8B',
    },
  };

  const LIGHT_THEMES = {
    clean: {
      '--bg':'#F5F7FA', '--bg-panel':'#EAECF0', '--bg-hover':'#DDE3EE',
      '--border':'#C8D0DC', '--border-active':'#007ACC', '--key':'#007ACC',
      '--array-idx':'#C0392B', '--string':'#27882B', '--number':'#B07D00',
      '--bool-true':'#2980B9', '--bool-false':'#C0392B', '--null':'#7F8C8D',
      '--path':'#8E44AD', '--search':'#E67E22', '--muted':'#95A5A6',
      '--text':'#1A2332', '--hash':'#8E44AD', '--stats':'#E67E22',
    },
    solarized: {
      '--bg':'#FDF6E3', '--bg-panel':'#EEE8D5', '--bg-hover':'#E0D9C5',
      '--border':'#D0C9B5', '--border-active':'#268BD2', '--key':'#268BD2',
      '--array-idx':'#DC322F', '--string':'#859900', '--number':'#B58900',
      '--bool-true':'#2AA198', '--bool-false':'#DC322F', '--null':'#93A1A1',
      '--path':'#D33682', '--search':'#CB4B16', '--muted':'#93A1A1',
      '--text':'#657B83', '--hash':'#D33682', '--stats':'#CB4B16',
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

  // ── Boot ────────────────────────────────────────────────────────────────────

  // fileParam / keyParam are set when navigating from the browser picker.
  const fileParam = new URLSearchParams(window.location.search).get('file');
  const keyParam  = new URLSearchParams(window.location.search).get('key');

  async function boot() {
    applyMode(resolveMode());

    try {
      let treeURL = '/api/tree';
      if (fileParam) treeURL = '/api/tree?file=' + encodeURIComponent(fileParam);
      else if (keyParam) treeURL = '/api/tree?key=' + encodeURIComponent(keyParam);
      const resp = await fetch(treeURL);
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
      `<span style="color:var(--key)">${treeData.nodeCount}</span> nodes · depth <span style="color:var(--key)">${treeData.maxDepth}</span> · <span style="color:var(--key)">${humanBytes(treeData.sizeBytes)}</span>`;
    document.getElementById('sep-meta').style.display = '';

    const header = document.getElementById('header');

    if (fileParam) {
      const link = document.createElement('a');
      link.href = '/';
      link.textContent = '[ change file ]';
      link.style.cssText = 'color:var(--muted);font-size:11px;text-decoration:none;transition:color 0.15s;';
      link.onmouseover = () => { link.style.color = 'var(--key)'; };
      link.onmouseout  = () => { link.style.color = 'var(--muted)'; };
      header.appendChild(link);
    }

    header.appendChild(renderModeToggle());
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

    // Auto-switch sidebar to Stats if numeric array
    if (node.kind === 'array' && node.stats && node.stats.numericCount > 0) {
      activateSidebarSection('stats');
      renderStatsSection(node);
    }
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
      { id: 'yaml',    label: 'YAML',             render: renderYAMLSection },
      { id: 'schema',  label: 'JSON SCHEMA',       render: renderSchemaSection },
      { id: 'merkle',  label: 'MERKLE ROOT',       render: renderMerkleSection },
      { id: 'stats',   label: 'STATISTICS',        render: renderStatsSection },
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

  function activateSidebarSection(id) {
    const section = document.getElementById('section-' + id);
    if (!section) return;
    const body = section.querySelector('.sidebar-section-body');
    if (body) body.classList.remove('collapsed');
  }

  function renderTypeGenSection(body, node) {
    // Sub-mode tabs
    body.innerHTML = '';
    const tabs = document.createElement('div');
    tabs.className = 'sub-tabs';
    for (const [mode, label] of [['go','Go'], ['ts','TypeScript'], ['jsonschema','JSON Schema']]) {
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

  function renderYAMLSection(body) {
    body.textContent = treeData.yaml || '';
  }

  function renderSchemaSection(body) {
    body.textContent = treeData.typegen ? (treeData.typegen['jsonschema'] || '') : '';
  }

  function renderMerkleSection(body) {
    const root = treeData.merkleRoot || '';
    body.innerHTML = root
      ? '<span class="syn-hash">root: ' + root + '</span>'
      : '<span style="color:var(--muted)">Merkle hashing not enabled.\nRun with: unum json &lt;file&gt; --web --merkle</span>';
  }

  function renderStatsSection(body, node) {
    const target = node && node.kind === 'array' && node.stats ? node : null;
    if (!target || !target.stats || target.stats.numericCount === 0) {
      body.textContent = 'Navigate to a numeric array to see statistics.';
      return;
    }
    const s = target.stats;
    body.innerHTML = [
      'Array: <span class="syn-hash">' + (target.path || '.') + '</span>',
      '',
      'Items:   ' + s.count + '  (numeric: ' + s.numericCount + ')',
      'Min:     <span class="syn-num">' + fmt(s.min) + '</span>',
      'Max:     <span class="syn-num">' + fmt(s.max) + '</span>',
      'Mean:    <span class="syn-num">' + fmt(s.mean) + '</span>',
      'Std Dev: <span class="syn-num">' + fmt(s.stddev) + '</span>',
      '',
      'p50:     ' + fmt(s.p50),
      'p95:     ' + fmt(s.p95),
      'p99:     ' + fmt(s.p99),
    ].join('\n');
  }

  // ── jq query ─────────────────────────────────────────────────────────────────

  function setupQuery() {
    const input = document.getElementById('query-input');
    const runBtn = document.getElementById('query-run');

    const run = async () => {
      const expr = input.value.trim();
      if (!expr) return;
      try {
        let queryURL = '/api/query';
        if (fileParam) queryURL = '/api/query?file=' + encodeURIComponent(fileParam);
        else if (keyParam) queryURL = '/api/query?key=' + encodeURIComponent(keyParam);
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
    });
  }

  // ── Toast ─────────────────────────────────────────────────────────────────────

  function showToast(msg) {
    const t = document.getElementById('toast');
    t.textContent = msg;
    t.style.display = 'block';
    setTimeout(() => { t.style.display = 'none'; }, 2000);
  }

  // ── Init ──────────────────────────────────────────────────────────────────────

  document.addEventListener('DOMContentLoaded', boot);
})();
