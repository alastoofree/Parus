// web/js/lens-rank.js — линза ранга

import { api, buildCharOrder, compareByASort } from './utils.js';

export function initRankLens(ctx) {
  const { projectId, frameIndex, getPairs, getTag, getASort } = ctx;
  const btn       = document.getElementById('rankToggleBtn');
  const container = document.getElementById('rankContainer');
  const tagValue = getTag();
  const rank     = parseInt(tagValue);
  if (!tagValue || rank <= 0) return;
  if (!btn || !container) return;
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
  // Кэш титулов и тегов
  let frameCache = []; // [{index, title, isTag}]
  let titleCache = {}; // index → title для быстрого доступа
  async function buildCache() {
    try {
      frameCache = await api.getFrames(projectId);
      frameCache.forEach(f => { titleCache[f.index] = f.title; });
    } catch(e) {
      frameCache = [];
      titleCache = {};}}
// 
  function resolveTitle(index) {
    const f = frameCache.find(f => f.index === index);
    return f ? f.title : index;}
// 
  function isTag(index) {
    const f = frameCache.find(f => f.index === index);
    return f ? f.isTag : false;}
// 
  function getISort(index) {
    // Читаем i-sort из frameCache — но там нет i-sort
    // i-sort есть только в полном фрейме, здесь используем
    // только то что уже загружено в метафрейме
    // Для корневого фрейма берём из getPairs()
    if (index === frameIndex) {
      const p = getPairs().find(p => p.Left === 'i-sort');
      return p ? p.Right : '';}
    return '';}
  // Рендер дерева
  async function renderRank(container) {
    container.innerHTML = '<div class="rank-loading">загрузка...</div>';
    await buildCache();
const result = await api.getChildren(projectId, frameIndex, getISort(frameIndex), getASort());
const children = result.map(c => c.index);
result.forEach(c => { titleCache[c.index] = c.title; });
    if (children.length === 0) {
      container.innerHTML = '<div class="rank-empty">нет дочерних фреймов</div>';
      return;}
    container.innerHTML = '';
    const tree = document.createElement('div');
    tree.className = 'rank-tree';
    for (const idx of children) {
	    tree.appendChild(await makeNode(idx, 1, rank));}
    container.appendChild(tree);}
  ctx.renderRankInto = (el) => renderRank(el);
// 
async function makeNode(index, currentDepth, maxDepth) {
    const hasChildren = currentDepth < maxDepth;
    const nodeEl      = document.createElement('div');
    nodeEl.className  = 'rank-node';
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
    marker.textContent = isTag(index) ? '⬡' : '▪';
    const link = document.createElement('a');
    link.className   = 'rank-link';
    link.href        = 'frame.html?id=' + projectId + '&frame=' + index;
    link.textContent = resolveTitle(index);
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
                const result = await api.getChildren(projectId, index, '', getASort());
                result.forEach(c => { titleCache[c.index] = c.title; });
                for (const c of result) {
                    childList.appendChild(await makeNode(c.index, currentDepth + 1, maxDepth));}
                childList.dataset.loaded = '1';}
            childList.classList.remove('hidden');
            arrow.textContent = '▼';
        } else {
            childList.classList.add('hidden');
            arrow.textContent = '▶';}});
      nodeEl.appendChild(childList);}
    return nodeEl;} }