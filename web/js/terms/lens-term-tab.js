// parus/web/js/terms/lens-term-tab.js — линза таба для терминологической базы

import { esc, renderFrameContent } from '../utils.js';

export function initTermTocLens(ctx) {
  const { getPairs, getToc, getISort, getASort, frameIndex } = ctx;
  const btn       = document.getElementById('tocToggleBtn');
  const container = document.getElementById('tocContainer');
  if (!btn || !container) return;

  btn.addEventListener('click', () => {
    if (!container.classList.contains('hidden')) {
      container.classList.add('hidden');
      btn.textContent = '↕';
      btn.classList.remove('active');
      btn.style.opacity = '0.4';
      return;}
    container.classList.remove('hidden');
    btn.textContent = '↕';
    btn.classList.add('active');
    btn.style.opacity = '1';
    if (!container.dataset.loaded) {
      renderToc(container, frameIndex, getToc(), getISort());
      container.dataset.loaded = '1';}});

  let titleCache = {};

  async function renderToc(container, tagPath, tocMode, iSort) {
    container.innerHTML = '<div class="toc-loading">load...</div>';
    try {
      const result = await ctx.api.getChildren(tagPath, iSort, getASort ? getASort() : '');
      result.forEach(c => { titleCache[c.index] = c.name || c.title; });
      if (result.length === 0) {
        container.innerHTML = '<div class="toc-empty">not frames</div>';
        return;}
      container.innerHTML = '';
      if (tocMode === 'hrznt') {
        renderHrznt(container, result, tagPath);
      } else {
        renderVrtcl(container, result, tagPath);}
    } catch(e) {
      container.innerHTML = '<div class="toc-error">error load</div>';}}

  function resolveTitle(path) {
    return titleCache[path] || path;}

  function renderHrznt(container, items, parentPath) {
    const wrap    = document.createElement('div');
    wrap.className = 'toc-hrznt';
    const header  = document.createElement('div');
    header.className = 'toc-hrznt-header';
    const content = document.createElement('div');
    content.className = 'toc-content';
    items.forEach((c, i) => {
      const b = document.createElement('button');
      b.className     = 'toc-btn-h' + (i === 0 ? ' active' : '');
      b.textContent   = c.name || c.title;
      b.dataset.index = c.index;
      b.addEventListener('click', () => {
        header.querySelectorAll('.toc-btn-h').forEach(x => x.classList.remove('active'));
        b.classList.add('active');
        loadTabContent(content, c.index);});
      header.appendChild(b);});
    wrap.appendChild(header);
    wrap.appendChild(content);
    container.appendChild(wrap);
    loadTabContent(content, items[0].index);}

  function renderVrtcl(container, items, parentPath) {
    const wrap    = document.createElement('div');
    wrap.className = 'toc-vrtcl';
    const sidebar = document.createElement('div');
    sidebar.className = 'toc-vrtcl-sidebar';
    const content = document.createElement('div');
    content.className = 'toc-content';
    items.forEach((c, i) => {
      const b = document.createElement('button');
      b.className     = 'toc-btn-v' + (i === 0 ? ' active' : '');
      b.textContent   = c.name || c.title;
      b.dataset.index = c.index;
      b.addEventListener('click', () => {
        sidebar.querySelectorAll('.toc-btn-v').forEach(x => x.classList.remove('active'));
        b.classList.add('active');
        loadTabContent(content, c.index);});
      sidebar.appendChild(b);});
    wrap.appendChild(sidebar);
    wrap.appendChild(content);
    container.appendChild(wrap);
    loadTabContent(content, items[0].index);}

  async function loadTabContent(contentEl, path) {
    if (contentEl.dataset.loadedId === path) return;
    contentEl.innerHTML = '<div class="toc-loading">load...</div>';
    try {
      const r        = await fetch('/api/terms/' + encodeURIComponent(path));
      const termData = await r.json();
      const termPairs = termData.pairs || [];
      const content   = (termPairs.find(p => p.Left === 'content') || {}).Right || '';
      const childToc  = (termPairs.find(p => p.Left === 'tab')     || {}).Right || '';
      const childSort = (termPairs.find(p => p.Left === 'i-sort')  || {}).Right || '';
      contentEl.innerHTML = '';
      contentEl.dataset.loadedId = path;
      if (content.trim()) {
        const cd = document.createElement('div');
        cd.className = 'toc-frame-content';
        cd.innerHTML = await renderFrameContent(content, 'markdown', '');
        contentEl.appendChild(cd);}
      if (childToc) {
        const childResult = await ctx.api.getChildren(path, childSort, getASort ? getASort() : '');
        if (childResult.length > 0) {
          childResult.forEach(c => { titleCache[c.index] = c.name || c.title; });
          const nestedWrap = document.createElement('div');
          nestedWrap.className = 'toc-nested';
          contentEl.appendChild(nestedWrap);
          if (childToc === 'hrznt') {
            renderHrznt(nestedWrap, childResult, path);
          } else {
            renderVrtcl(nestedWrap, childResult, path);}}}
      if (!content.trim() && !childToc) {
        contentEl.innerHTML = '<div class="toc-empty">not content</div>';}
    } catch(e) {
      contentEl.innerHTML = '<div class="toc-error">error load</div>';}} }