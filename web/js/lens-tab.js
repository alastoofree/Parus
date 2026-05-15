// web/js/lens-tab.js — линза таба

import { api, renderFrameContent, buildCharOrder, compareByASort } from './utils.js';

export function initTocLens(ctx) {
const {
    projectId, frameIndex,
    getPairs, getToc, getISort,
  } = ctx;

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
// 
  function resolveTitle(index) {
    return titleCache[index] || index;}
//
  let titleCache = {};
  async function buildTitleCache() {
    try {
      const frames = await api.getFrames(projectId);
      frames.forEach(f => { titleCache[f.index] = f.title; });
    } catch(e) { titleCache = {}; }}
// 
  async function renderToc(container, tagIndex, tocMode, iSort) {
    container.innerHTML = '<div class="toc-loading">load...</div>';
    await buildTitleCache();
    const result = await api.getChildren(projectId, tagIndex, iSort, ctx.getASort ? ctx.getASort() : '');
    const children = result.map(c => c.index);
    result.forEach(c => { titleCache[c.index] = c.title; });
    if (children.length === 0) {
      container.innerHTML = '<div class="toc-empty">not frames</div>';
      return;}
    container.innerHTML = '';
    if (tocMode === 'hrznt') {
      renderHrznt(container, children, tagIndex);
    } else {
      renderVrtcl(container, children, tagIndex);}}
  ctx.renderTocInto = (el, idx, mode, iSort) => renderToc(el, idx, mode, iSort);
// 
  function renderHrznt(container, children, parentIndex) {
    const wrap    = document.createElement('div');
    wrap.className = 'toc-hrznt';
    const header  = document.createElement('div');
    header.className = 'toc-hrznt-header';
    const content = document.createElement('div');
    content.className = 'toc-content';
    children.forEach((idx, i) => {
      const b = document.createElement('button');
      b.className   = 'toc-btn-h' + (i === 0 ? ' active' : '');
      b.textContent = resolveTitle(idx);
      b.dataset.index = idx;
      b.addEventListener('click', () => {
        header.querySelectorAll('.toc-btn-h')
              .forEach(x => x.classList.remove('active'));
        b.classList.add('active');
        loadTabContent(content, idx);});
      header.appendChild(b);});
    wrap.appendChild(header);
    wrap.appendChild(content);
    container.appendChild(wrap);
    loadTabContent(content, children[0]);}
// 
  function renderVrtcl(container, children, parentIndex) {
    const wrap    = document.createElement('div');
    wrap.className = 'toc-vrtcl';
    const sidebar = document.createElement('div');
    sidebar.className = 'toc-vrtcl-sidebar';
    const content = document.createElement('div');
    content.className = 'toc-content';
    children.forEach((idx, i) => {
      const b = document.createElement('button');
      b.className   = 'toc-btn-v' + (i === 0 ? ' active' : '');
      b.textContent = resolveTitle(idx);
      b.dataset.index = idx;
      b.addEventListener('click', () => {
        sidebar.querySelectorAll('.toc-btn-v')
               .forEach(x => x.classList.remove('active'));
        b.classList.add('active');
        loadTabContent(content, idx);});
      sidebar.appendChild(b);});
    wrap.appendChild(sidebar);
    wrap.appendChild(content);
    container.appendChild(wrap);
    loadTabContent(content, children[0]);}
// 
  async function loadTabContent(contentEl, idx) {
    if (contentEl.dataset.loadedId === idx) return;
    contentEl.innerHTML = '<div class="toc-loading">load...</div>';
    try {
      const frame      = await api.getFrame(projectId, idx);
      const framePairs = frame.pairs || [];
      const childToc   = (framePairs.find(p => p.Left === 'tab')    || {}).Right || '';
      const childISort = (framePairs.find(p => p.Left === 'i-sort') || {}).Right || '';
      const r    = await fetch('/api/projects/' + projectId + '/content/' + idx);
      const data = await r.json();
      contentEl.innerHTML = '';
      contentEl.dataset.loadedId = idx;
      if (data.content && data.content.trim()) {
        const cd = document.createElement('div');
        cd.className = 'toc-frame-content';
        cd.innerHTML = await renderFrameContent(data.content, data.contentType || 'text', projectId);
        contentEl.appendChild(cd);}
      if (childToc) {
        const childResult = await api.getChildren(projectId, idx, childISort, ctx.getASort ? ctx.getASort() : '');
        const childChildren = childResult.map(c => c.index);
        if (childChildren.length > 0) {
          const nestedWrap = document.createElement('div');
          nestedWrap.className = 'toc-nested';
          contentEl.appendChild(nestedWrap);
          if (childToc === 'hrznt') {
            renderHrznt(nestedWrap, childChildren, idx);
          } else {
            renderVrtcl(nestedWrap, childChildren, idx);}}}
      if (!data.content?.trim() && !childToc) {
        contentEl.innerHTML =
          '<div class="toc-empty">not content</div>';}
    } catch(e) {
      contentEl.innerHTML = '<div class="toc-error">error load</div>';}} }