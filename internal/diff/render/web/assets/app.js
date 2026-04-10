/* unum diff web UI — vanilla JS, no framework, no build step */

(function () {
  'use strict';

  let diffData = null;
  let viewMode = 'unified'; // set properly in renderHeader once data arrives
  let searchQuery = '';

  // ── Themes ───────────────────────────────────────────────────────────────────

  const THEMES = {
    cyber: {
      '--bg':           '#0D0D0D',
      '--bg-panel':     '#111111',
      '--bg-hover':     '#1A1A2E',
      '--bg-added':     '#0d1a0d',
      '--bg-removed':   '#1a0d0d',
      '--border':       '#1E1E1E',
      '--border-active':'#00D4FF',
      '--text':         '#C0C0C0',
      '--muted':        '#3A3A3A',
      '--path':         '#C678DD',
      '--search':       '#FFD700',
      '--diff-added':   '#98C379',
      '--diff-removed': '#E06C75',
      '--diff-hunk':    '#00D4FF',
    },
    matrix: {
      '--bg':           '#0D0D0D',
      '--bg-panel':     '#0A1A0A',
      '--bg-hover':     '#001A00',
      '--bg-added':     '#001800',
      '--bg-removed':   '#180000',
      '--border':       '#003300',
      '--border-active':'#00FF41',
      '--text':         '#00CC22',
      '--muted':        '#005500',
      '--path':         '#39FF14',
      '--search':       '#FFFFFF',
      '--diff-added':   '#00FF41',
      '--diff-removed': '#FF3300',
      '--diff-hunk':    '#39FF14',
    },
    dracula: {
      '--bg':           '#282A36',
      '--bg-panel':     '#21222C',
      '--bg-hover':     '#44475A',
      '--bg-added':     '#1e3128',
      '--bg-removed':   '#3a1a1e',
      '--border':       '#3D4050',
      '--border-active':'#BD93F9',
      '--text':         '#F8F8F2',
      '--muted':        '#6272A4',
      '--path':         '#FF79C6',
      '--search':       '#F1FA8C',
      '--diff-added':   '#50FA7B',
      '--diff-removed': '#FF5555',
      '--diff-hunk':    '#BD93F9',
    },
    nord: {
      '--bg':           '#2E3440',
      '--bg-panel':     '#272C36',
      '--bg-hover':     '#3B4252',
      '--bg-added':     '#2c3a2c',
      '--bg-removed':   '#3a2c2e',
      '--border':       '#3B4252',
      '--border-active':'#88C0D0',
      '--text':         '#ECEFF4',
      '--muted':        '#4C566A',
      '--path':         '#B48EAD',
      '--search':       '#EBCB8B',
      '--diff-added':   '#A3BE8C',
      '--diff-removed': '#BF616A',
      '--diff-hunk':    '#88C0D0',
    },
  };

  function applyTheme(name) {
    const vars = THEMES[name] || THEMES.cyber;
    const root = document.documentElement;
    for (const [k, v] of Object.entries(vars)) {
      root.style.setProperty(k, v);
    }
    localStorage.setItem('unum-theme', name);
    document.querySelectorAll('.theme-btn').forEach(btn => {
      btn.classList.toggle('active', btn.dataset.theme === name);
    });
  }

  function renderThemePicker() {
    const picker = document.createElement('div');
    picker.id = 'theme-picker';
    picker.style.cssText = 'display:flex;align-items:center;gap:6px;margin-left:auto;';

    const label = document.createElement('span');
    label.textContent = 'theme:';
    label.style.cssText = 'color:var(--muted);font-size:11px;';
    picker.appendChild(label);

    const saved = localStorage.getItem('unum-theme') || 'cyber';
    for (const name of Object.keys(THEMES)) {
      const btn = document.createElement('button');
      btn.className = 'theme-btn' + (name === saved ? ' active' : '');
      btn.dataset.theme = name;
      btn.textContent = name;
      btn.style.cssText = [
        'background:none', 'border:1px solid var(--border)',
        'color:var(--muted)', 'font-size:10px', 'padding:2px 6px',
        'border-radius:3px', 'cursor:pointer', 'font-family:inherit',
        'transition:all 0.15s',
      ].join(';');
      btn.onmouseover = () => {
        btn.style.borderColor = 'var(--border-active)';
        btn.style.color = 'var(--border-active)';
      };
      btn.onmouseout = () => {
        btn.style.borderColor = btn.classList.contains('active') ? 'var(--border-active)' : 'var(--border)';
        btn.style.color = btn.classList.contains('active') ? 'var(--border-active)' : 'var(--muted)';
      };
      btn.onclick = () => applyTheme(name);
      picker.appendChild(btn);
    }

    return picker;
  }

  // ── Boot ─────────────────────────────────────────────────────────────────────

  async function boot() {
    applyTheme(localStorage.getItem('unum-theme') || 'cyber');
    try {
      const resp = await fetch('/api/diff');
      if (!resp.ok) throw new Error('Failed to load diff: ' + resp.status);
      diffData = await resp.json();
      renderHeader();
      renderDiff();
      setupSearch();
      setupKeyboard();
    } catch (e) {
      document.getElementById('diff-panel').textContent = '✗ ' + e.message;
    }
  }

  // ── Header ────────────────────────────────────────────────────────────────────

  function renderHeader() {
    document.getElementById('file-a').textContent = diffData.fileA;
    document.getElementById('file-b').textContent = diffData.fileB;
    document.getElementById('stat-added').textContent = '+' + diffData.added;
    document.getElementById('stat-removed').textContent = '-' + diffData.removed;
    if (diffData.modified > 0) {
      const modEl = document.createElement('span');
      modEl.id = 'stat-modified';
      modEl.className = 'stat-modified';
      modEl.textContent = '~' + diffData.modified;
      document.querySelector('.stats').appendChild(modEl);
    }
    document.getElementById('status-format').textContent =
      diffData.format !== 'text' ? '[' + diffData.format + ']' : '';

    document.getElementById('header').appendChild(renderThemePicker());

    // Set default view mode and wire up the toolbar toggle
    viewMode = diffData.tree ? 'semantic' : 'unified';
    renderViewToggle();
  }

  // ── View mode toggle ──────────────────────────────────────────────────────────

  function availableModes() {
    const hasSemantic = !!diffData.tree;
    const hasText = diffData.hunks && diffData.hunks.length > 0;
    if (hasSemantic && hasText) return ['semantic', 'unified', 'split'];
    if (hasText)                return ['unified', 'split'];
    return []; // tree-only (e.g. Terraform) — no toggle
  }

  function renderViewToggle() {
    const placeholder = document.getElementById('toggle-view');
    const modes = availableModes();

    if (modes.length === 0) {
      placeholder.style.display = 'none';
      return;
    }

    const group = document.createElement('div');
    group.className = 'view-modes';
    group.id = 'view-modes';

    for (const mode of modes) {
      const btn = document.createElement('button');
      btn.className = 'view-mode-btn' + (mode === viewMode ? ' active' : '');
      btn.dataset.mode = mode;
      btn.textContent = mode;
      btn.onclick = () => setViewMode(mode);
      group.appendChild(btn);
    }

    placeholder.replaceWith(group);
  }

  function setViewMode(mode) {
    viewMode = mode;
    document.querySelectorAll('.view-mode-btn').forEach(btn => {
      btn.classList.toggle('active', btn.dataset.mode === mode);
    });
    renderDiff();
  }

  // ── Diff rendering ────────────────────────────────────────────────────────────

  function renderDiff() {
    const panel = document.getElementById('diff-panel');
    panel.innerHTML = '';
    panel.classList.toggle('is-split', viewMode === 'split');

    if (viewMode === 'semantic' && diffData.tree) {
      renderTreeView(panel, diffData.tree);
      if (searchQuery) applyHighlight();
      return;
    }

    const hasHunks = diffData.hunks && diffData.hunks.length > 0;
    if (!hasHunks) {
      const empty = document.createElement('div');
      empty.className = 'empty-state';
      empty.textContent = '(no differences)';
      panel.appendChild(empty);
      return;
    }

    if (viewMode === 'split') {
      renderSplit(panel);
    } else {
      renderUnified(panel);
    }

    if (searchQuery) applyHighlight();
  }

  // ── Semantic tree view ────────────────────────────────────────────────────────

  let collapsedPaths = new Set();

  function renderTreeView(container, tree) {
    renderTreeNode(container, tree, 0);
  }

  function renderTreeNode(container, node, depth) {
    if (!node) return;
    const kind = node.kind;

    if (kind === 'unchanged' && node.children && node.children.length > 0) {
      // Container with mixed children — render as collapsible
      const label = node.key ? '"' + node.key + '"' : (node.index >= 0 ? '[' + node.index + ']' : '.');
      const isCollapsed = collapsedPaths.has(node.path);
      const toggle = document.createElement('div');
      toggle.className = 'tree-row tree-unchanged';
      toggle.dataset.path = node.path;
      toggle.style.paddingLeft = (depth * 16 + 8) + 'px';
      toggle.innerHTML =
        '<span class="tree-toggle">' + (isCollapsed ? '▶' : '▼') + '</span>' +
        '<span class="tree-key">' + escapeHTML(label) + '</span>';
      toggle.onclick = () => {
        if (collapsedPaths.has(node.path)) {
          collapsedPaths.delete(node.path);
        } else {
          collapsedPaths.add(node.path);
        }
        renderDiff();
      };
      container.appendChild(toggle);
      if (!isCollapsed) {
        for (const child of node.children) {
          renderTreeNode(container, child, depth + 1);
        }
      }
      return;
    }

    if (kind === 'unchanged') {
      // Unchanged leaf — skip (keep output focused on changes)
      return;
    }

    const el = document.createElement('div');
    el.className = 'tree-row tree-' + kind;
    el.dataset.content = (node.oldValue || '') + (node.newValue || '') + node.path;
    el.style.paddingLeft = (depth * 16 + 8) + 'px';

    const prefix = kind === 'added' ? '+' : (kind === 'removed' ? '-' : '~');
    const pathEl = '<span class="tree-path">' + escapeHTML(node.path) + '</span>';
    let valueEl = '';

    if (kind === 'modified') {
      valueEl = '<span class="tree-old">' + escapeHTML(node.oldValue) + '</span>' +
                '<span class="tree-arrow"> → </span>' +
                '<span class="tree-new">' + escapeHTML(node.newValue) + '</span>';
    } else if (kind === 'added') {
      valueEl = '<span class="tree-new">' + escapeHTML(node.newValue) + '</span>';
    } else {
      valueEl = '<span class="tree-old">' + escapeHTML(node.oldValue) + '</span>';
    }

    el.innerHTML = '<span class="tree-prefix">' + prefix + ' </span>' + pathEl + '  ' + valueEl;

    // If it has children (added/removed container), render them collapsed
    if (node.children && node.children.length > 0) {
      const isCollapsed = collapsedPaths.has(node.path);
      el.querySelector('.tree-prefix').textContent = prefix + (isCollapsed ? ' ▶ ' : ' ▼ ');
      el.onclick = (e) => {
        e.stopPropagation();
        if (collapsedPaths.has(node.path)) {
          collapsedPaths.delete(node.path);
        } else {
          collapsedPaths.add(node.path);
        }
        renderDiff();
      };
      container.appendChild(el);
      if (!isCollapsed) {
        for (const child of node.children) {
          renderTreeNode(container, child, depth + 1);
        }
      }
      return;
    }

    container.appendChild(el);
  }

  function renderUnified(container) {
    for (const hunk of diffData.hunks) {
      const hdrEl = document.createElement('div');
      hdrEl.className = 'diff-line hunk-header';
      hdrEl.textContent =
        '@@ -' + hunk.oldStart + ',' + hunk.oldCount +
        ' +' + hunk.newStart + ',' + hunk.newCount + ' @@';
      container.appendChild(hdrEl);

      for (const line of hunk.lines) {
        container.appendChild(buildUnifiedLine(line));
      }
    }
  }

  function buildUnifiedLine(line) {
    const el = document.createElement('div');
    el.className = 'diff-line diff-' + line.kind;
    el.dataset.content = line.content;

    const prefix = line.kind === 'added' ? '+' : (line.kind === 'removed' ? '-' : ' ');
    const oldStr = line.oldNum > 0 ? String(line.oldNum).padStart(4) : '    ';
    const newStr = line.newNum > 0 ? String(line.newNum).padStart(4) : '    ';

    const gutter = document.createElement('span');
    gutter.className = 'gutter';
    gutter.textContent = oldStr + ' ' + newStr + ' ';

    const content = document.createElement('span');
    content.className = 'line-content';
    content.textContent = prefix + line.content;

    el.appendChild(gutter);
    el.appendChild(content);
    return el;
  }

  function renderSplit(container) {
    const wrapper = document.createElement('div');
    wrapper.className = 'split-wrapper';

    const leftPanel = document.createElement('div');
    leftPanel.className = 'split-side split-left';
    leftPanel.id = 'split-left';

    const rightPanel = document.createElement('div');
    rightPanel.className = 'split-side split-right';
    rightPanel.id = 'split-right';

    for (const hunk of diffData.hunks) {
      const hdr =
        '@@ -' + hunk.oldStart + ',' + hunk.oldCount +
        ' +' + hunk.newStart + ',' + hunk.newCount + ' @@';
      appendHunkHeader(leftPanel, hdr);
      appendHunkHeader(rightPanel, hdr);

      let i = 0;
      while (i < hunk.lines.length) {
        const l = hunk.lines[i];
        if (l.kind === 'removed') {
          const removed = [];
          while (i < hunk.lines.length && hunk.lines[i].kind === 'removed') {
            removed.push(hunk.lines[i++]);
          }
          const added = [];
          while (i < hunk.lines.length && hunk.lines[i].kind === 'added') {
            added.push(hunk.lines[i++]);
          }
          const maxLen = Math.max(removed.length, added.length);
          for (let j = 0; j < maxLen; j++) {
            leftPanel.appendChild(j < removed.length
              ? buildSplitLine(removed[j].kind, removed[j].oldNum, removed[j].content)
              : buildEmptyLine());
            rightPanel.appendChild(j < added.length
              ? buildSplitLine(added[j].kind, added[j].newNum, added[j].content)
              : buildEmptyLine());
          }
        } else if (l.kind === 'added') {
          leftPanel.appendChild(buildEmptyLine());
          rightPanel.appendChild(buildSplitLine(l.kind, l.newNum, l.content));
          i++;
        } else {
          leftPanel.appendChild(buildSplitLine(l.kind, l.oldNum, l.content));
          rightPanel.appendChild(buildSplitLine(l.kind, l.newNum, l.content));
          i++;
        }
      }
    }

    wrapper.appendChild(leftPanel);
    wrapper.appendChild(rightPanel);
    container.appendChild(wrapper);

    syncScroll(leftPanel, rightPanel);
  }

  function appendHunkHeader(container, text) {
    const el = document.createElement('div');
    el.className = 'diff-line hunk-header';
    el.textContent = text;
    container.appendChild(el);
  }

  function buildSplitLine(kind, num, content) {
    const el = document.createElement('div');
    el.className = 'diff-line diff-' + kind;
    el.dataset.content = content;

    const prefix = kind === 'added' ? '+' : (kind === 'removed' ? '-' : ' ');
    const numStr = num > 0 ? String(num).padStart(4) : '    ';

    const gutter = document.createElement('span');
    gutter.className = 'gutter';
    gutter.textContent = numStr + ' ';

    const contentEl = document.createElement('span');
    contentEl.className = 'line-content';
    contentEl.textContent = prefix + content;

    el.appendChild(gutter);
    el.appendChild(contentEl);
    return el;
  }

  function buildEmptyLine() {
    const el = document.createElement('div');
    el.className = 'diff-line diff-empty';
    return el;
  }

  function syncScroll(left, right) {
    let syncing = false;
    left.addEventListener('scroll', () => {
      if (syncing) return;
      syncing = true;
      right.scrollTop = left.scrollTop;
      syncing = false;
    });
    right.addEventListener('scroll', () => {
      if (syncing) return;
      syncing = true;
      left.scrollTop = right.scrollTop;
      syncing = false;
    });
  }

  // ── Search ────────────────────────────────────────────────────────────────────

  function setupSearch() {
    const input = document.getElementById('search-input');
    input.addEventListener('input', () => {
      searchQuery = input.value.trim().toLowerCase();
      renderDiff();
    });
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        input.value = '';
        searchQuery = '';
        renderDiff();
      }
    });
  }

  function applyHighlight() {
    const q = searchQuery;
    if (!q) return;
    let firstMatch = null;

    document.querySelectorAll('[data-content]').forEach(el => {
      const raw = (el.dataset.content || '').toLowerCase();
      if (!raw.includes(q)) return;

      el.classList.add('line-match');
      if (!firstMatch) firstMatch = el;

      const contentEl = el.querySelector('.line-content');
      if (!contentEl) return;
      const text = contentEl.textContent;
      const idx = text.toLowerCase().indexOf(q);
      if (idx < 0) return;
      const before = escapeHTML(text.slice(0, idx));
      const match = escapeHTML(text.slice(idx, idx + q.length));
      const after = escapeHTML(text.slice(idx + q.length));
      contentEl.innerHTML = before + '<span class="search-hl">' + match + '</span>' + after;
    });

    if (firstMatch) {
      firstMatch.scrollIntoView({ block: 'center', behavior: 'smooth' });
    }
  }

  function escapeHTML(s) {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  // ── Keyboard ──────────────────────────────────────────────────────────────────

  function setupKeyboard() {
    document.addEventListener('keydown', (e) => {
      if (e.target.tagName === 'INPUT') return;
      if (e.key === '/') {
        e.preventDefault();
        document.getElementById('search-input').focus();
      }
      if (e.key === 'v') {
        const modes = availableModes();
        if (modes.length > 1) {
          const idx = modes.indexOf(viewMode);
          setViewMode(modes[(idx + 1) % modes.length]);
        }
      }
      if (e.key === 'j') {
        document.getElementById('diff-panel').scrollBy(0, 40);
      }
      if (e.key === 'k') {
        document.getElementById('diff-panel').scrollBy(0, -40);
      }
    });
  }

  // ── Init ──────────────────────────────────────────────────────────────────────

  document.addEventListener('DOMContentLoaded', boot);
})();
