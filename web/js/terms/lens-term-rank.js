// parus/web/js/terms/lens-term-rank.js — линза ранга для терминологической базы

import { esc, buildCharOrder, compareByASort } from '../utils.js';

export function initTermRankLens(ctx) {
  const { getPairs, getASort, frameIndex } = ctx;
  const btn       = document.getElementById('rankToggleBtn');
  const container = document.getElementById('rankContainer');
  if (!btn || !container) return;
  const rank = (getPairs().find(p => p.Left === 'rank') || {}).Right || '';
  if (!rank) return;
  btn.style.display = '';
  btn.addEventListener('click', () => {
    if (!container.classList.contains('hidden')) {
      container.classList.add('hidden');
      btn.classList.remove('active');
      btn.style.opacity = '0.4';
      return;}
    container.classList.remove('hidden');
    btn.classList.add('active');
    btn.style.opacity = '1';
    if (!container.dataset.loaded) {
      renderRank(container);
      container.dataset.loaded = '1';}});

  let titleCache = {};

  async function renderRank(container) {
    container.innerHTML = '<div class="rank-loading">загрузка...</div>';
    try {
      const result = await ctx.api.getChildren(frameIndex, '', getASort ? getASort() : '');
      if (!result || result.length === 0) {
        container.innerHTML = '<div class="rank-empty">нет дочерних термов</div>';
        return;}
      result.forEach(c => { titleCache[c.index] = c.name || c.title; });
      container.innerHTML = '';
      const tree = document.createElement('div');
      tree.className = 'rank-tree';
      const maxDepth = parseInt(rank) || 1;
      for (const c of result) {
        tree.appendChild(await makeNode(c.index, c.name || c.title, 1, maxDepth));}
      container.appendChild(tree);
    } catch(e) {
      container.innerHTML = '<div class="rank-error">ошибка загрузки</div>';}}

  async function makeNode(path, name, currentDepth, maxDepth) {
    const hasChildren = currentDepth < maxDepth;
    const nodeEl = document.createElement('div');
    nodeEl.className = 'rank-node';
    const row = document.createElement('div');
    row.className = 'rank-row';
    const arrow = document.createElement('span');
    arrow.className = 'rank-arrow';
    if (hasChildren) {
      arrow.textContent = '▶';
      arrow.classList.add('rank-arrow-has-children');
    } else {
      arrow.textContent = ' ';}
    const marker = document.createElement('span');
    marker.className   = 'rank-marker';
    marker.textContent = '▪';
    const link = document.createElement('a');
    link.className   = 'rank-link';
    link.href        = 'term.html?path=' + encodeURIComponent(path);
    link.textContent = name;
    row.appendChild(arrow);
    row.appendChild(marker);
    row.appendChild(link);
    nodeEl.appendChild(row);
    if (hasChildren) {
      const childList = document.createElement('div');
      childList.className = 'rank-children hidden';
      arrow.addEventListener('click', async e => {
        e.preventDefault();
        const isHidden = childList.classList.contains('hidden');
        if (isHidden) {
          if (!childList.dataset.loaded) {
            const result = await ctx.api.getChildren(path, '', getASort ? getASort() : '');
            result.forEach(c => { titleCache[c.index] = c.name || c.title; });
            for (const c of result) {
              childList.appendChild(await makeNode(c.index, c.name || c.title, currentDepth + 1, maxDepth));}
            childList.dataset.loaded = '1';}
          childList.classList.remove('hidden');
          arrow.textContent = '▼';
        } else {
          childList.classList.add('hidden');
          arrow.textContent = '▶';}});
      nodeEl.appendChild(childList);}
    return nodeEl;} }