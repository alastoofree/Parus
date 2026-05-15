// parus/web/js/frame.js — ядро: слоты, сохранение, init

import { api, esc, toGauss } from './utils.js';
import { ParseqGraph } from './graph.js';
import { initTagsLens }      from './lens-tags.js';
import { initContentLens }   from './lens-content.js';
import { initRankLens } from './lens-rank.js';
import { initTocLens } from './lens-tab.js';
import { initChronoLens } from './lens-chrono.js';

const params     = new URLSearchParams(location.search);
const projectId  = params.get('id');
const frameIndex = params.get('frame');
if (!projectId || !frameIndex) window.location.href = 'portal.html';
const READONLY_KEYS = [
  'id','modified','content','content-type','tags','links','includes'];
const READONLY_LEFT_KEYS = [
  'title', 'alias', 'icon','color', 'rank','tab', 'i-sort', 
  'west', 'nord', 'graph','list', 'match-nord','match-south', 
  'chrono', 'i-chrono', 'dict', 'i-dict', 'show-product', 'product' ];
let pairs        = [];
let savedPairs   = [];
let slotOrderLen = 2;
let aliasBefore  = [];

function render() {
    const infoKeys = new Set(['id','modified','title','alias','content','content-type',
                              'tags','links','includes','icon','color', 'rank','tab', 'i-sort',
                              'west','nord','graph','list','match-nord','match-south',
                              'chrono','i-chrono','dict','i-dict','show-product','product']);
    const sorted = [
      ...pairs.filter(p => !infoKeys.has(p.Left)),
      ...pairs.filter(p =>  infoKeys.has(p.Left)),];
  document.getElementById('slotsList').innerHTML = sorted.map((p) => {
    const i = pairs.indexOf(p);
    const ro     = READONLY_KEYS.includes(p.Left);
    const roLeft = ro || READONLY_LEFT_KEYS.includes(p.Left);
    let rowClass = 'slot-row';
    if (ro) rowClass += ' readonly';
    if (['links','tags','alias','includes'].includes(p.Left)) rowClass += ' slot-row-link';
    return `
      <div class="${rowClass}">
        <span class="slot-row-index">${esc(p.TypeIndex)}</span>
        <div class="slot-cell key ${roLeft ? 'readonly' : ''}">
          <input type="text" value="${esc(p.Left)}" data-i="${i}" data-side="Left"
            ${roLeft ? 'readonly tabindex="-1"' : ''} placeholder="терм"
            autocomplete="off" spellcheck="false">
        </div>
        <div class="slot-cell ${ro ? 'readonly' : ''}">
          <input type="text" value="${esc(p.Right)}" data-i="${i}" data-side="Right"
            ${ro ? 'readonly tabindex="-1"' : ''} placeholder="${linkPlaceholder(p.Left)}"
            autocomplete="off" spellcheck="false">
        </div>
      </div>`;
  }).join('');}
// 
function linkPlaceholder(left) {
switch(left) {
    case 'alias':    return 'гаусс-индексы синонимов';
    case 'tags':     return 'гаусс-индексы тегов';
    case 'links':    return 'гаусс-индексы ссылок';
    case 'includes': return 'гаусс-индексы включений';
default:         return 'значение';}}
// Обработчики слотов
document.getElementById('slotsList').addEventListener('input', e => {
  const input = e.target.closest('input');
  if (!input) return;
  pairs[parseInt(input.dataset.i)][input.dataset.side] = input.value;});
document.getElementById('btnAddSlot').addEventListener('click', () => {
  const userSlots = pairs.filter(p => p.TypeIndex.length === slotOrderLen);
  pairs.push({ TypeIndex: toGauss(userSlots.length, slotOrderLen), Left: '', Right: '' });
  render();
  const inputs = document.getElementById('slotsList').querySelectorAll('input[data-side="Left"]');
  const last = inputs[inputs.length - 1];
  if (last) last.focus();});
document.getElementById('btnSave').addEventListener('click', async () => {
  const today = new Date().toISOString().slice(0, 10);
  for (let p of pairs) {
    if (p.Left === 'modified') { p.Right = today; break; }}
  await api.saveFrame(projectId, frameIndex, pairs, aliasBefore);
  const ap = pairs.find(p => p.Left === 'alias');
  aliasBefore = ap ? ap.Right.trim().split(/\s+/).filter(Boolean) : [];
  savedPairs = JSON.parse(JSON.stringify(pairs));
  const btn = document.getElementById('btnSave');
  btn.style.background = '#2a7a2a';
  btn.style.borderColor = '#2a7a2a';
  setTimeout(() => { 
    btn.style.background = '#e8a838'; 
    btn.style.borderColor = '#e8a838';}, 1500);
  render();});
const btnDiscard = document.getElementById('btnDiscard');
if (btnDiscard) btnDiscard.addEventListener('click', () => {
  pairs = JSON.parse(JSON.stringify(savedPairs));
  render();});
document.addEventListener('cliInput' in document ? 'keydown' : 'keydown', () => {});
// Навигация из графа
document.addEventListener('parseq-navigate', e => {
  const { projectId: navProject, frameIndex: navFrame } = e.detail;
  // Узел-база в метаграфе имеет префикс "base:"
  if (navFrame && navFrame.startsWith('base:')) {
    const baseId = navFrame.slice(5);
    window.location.href = 'project.html?id=' + baseId; return;}
  window.location.href =
    'frame.html?id=' + (navProject || projectId) + '&frame=' + navFrame;});
document.getElementById('cliInput').addEventListener('keydown', async e => {
if (e.key !== 'Enter') return;
const query = e.target.value.trim();
e.target.value = '';
if (!query) return;
const results = await api.searchFrame(projectId, frameIndex, query);
    const container = document.getElementById('searchResults');
if (!results || results.length === 0) {
    container.innerHTML = '<div class="search-empty">не найдено</div>';
    container.classList.remove('hidden');
    return;}
container.innerHTML = results.map(r =>
    `<div class="search-result">
        <a href="frame.html?id=${projectId}&frame=${r.index}">${r.title}</a>
        <div class="search-excerpt">${r.excerpt}</div>
    </div>`).join('');
container.classList.remove('hidden');});
// Инициализация
async function init() {
  let metaPairs = [];
  try {
    const p = await api.getProject(projectId);
    const titlePair = (p.pairs || []).find(pair => pair.Left === 'title');
    document.getElementById('breadcrumbProject').textContent =
        projectId === '00000-META-BASE' ? 'META-BASE' : (titlePair ? titlePair.Right : '—');
    document.getElementById('breadcrumbProject').href =
      'project.html?id=' + projectId;
    metaPairs = p.pairs || [];
    if (projectId === '00000-META-BASE') {
        const metaContentOf = metaPairs.find(p => p.Left === 'meta-content-of');
        if (metaContentOf) {
            metaPairs = metaPairs.filter(p => p.Left !== 'content-of-tag');
            metaPairs.push({ Left: 'content-of-tag', Right: metaContentOf.Right });}}
  } catch(e) {
    document.getElementById('breadcrumbProject').textContent = projectId === '00000-META-BASE' ? 'META-BASE' : '—';}
  const favicon = document.getElementById('favicon');
  if (favicon) {
    const exts = ['png','svg','ico','jpg','jpeg'];
    for (const ext of exts) {
      const url = '/img/' + projectId + '/favicon.' + ext + '?v=' + Date.now();
      const r = await fetch(url, { method: 'HEAD' });
      if (r.ok) { favicon.href = url; break; }}}
  try {
    const frame = await api.getFrame(projectId, frameIndex);
    pairs = frame.pairs || [];
    const metaGraphParam = (metaPairs.find(p => p.Left === 'graph') || {}).Right || '';
    const metaListParam  = (metaPairs.find(p => p.Left === 'list')  || {}).Right || '';
    const graphParam = (pairs.find(p => p.Left === 'graph') || {}).Right || metaGraphParam;
    const listParam  = (pairs.find(p => p.Left === 'list')  || {}).Right || metaListParam;
    const westSpec = (metaPairs.find(p => p.Left === 'west') || {}).Right || '';
    const nordSpec = (metaPairs.find(p => p.Left === 'nord') || {}).Right || '';
    graph = new ParseqGraph('dagContainer', projectId, graphParam, listParam, westSpec, nordSpec);
    savedPairs = JSON.parse(JSON.stringify(pairs));
    const ap = pairs.find(p => p.Left === 'alias');
    aliasBefore = ap ? ap.Right.trim().split(/\s+/).filter(Boolean) : [];
    slotOrderLen = frameIndex.length;
    const ulid = (pairs.find(p => p.Left === 'id') || {}).Right || '';
    const ulidEl = document.getElementById('frameUlid');
    if (ulidEl) ulidEl.textContent = ulid;
    document.getElementById('breadcrumbFrame').textContent =
      (pairs.find(p => p.Left === 'title') || {}).Right || frameIndex;
  } catch(e) {
    document.getElementById('breadcrumbFrame').textContent = frameIndex;}
  render();
  //     
  const tagPair  = pairs.find(p => p.Left === 'rank');
  const tagValue = tagPair ? tagPair.Right.trim() : '';
  const rankNum  = parseInt(tagValue);
  if (tagValue && rankNum > 0) {
    const rankBtn = document.getElementById('rankToggleBtn');
    const title   = (pairs.find(p => p.Left === 'title') || {}).Right || frameIndex;
    const icon    = (pairs.find(p => p.Left === 'icon')  || {}).Right || '';
    const color   = (pairs.find(p => p.Left === 'color') || {}).Right || '';
    rankBtn.textContent = icon && !icon.includes('.') ? '' : '◈';
    rankBtn.title = title;
    rankBtn.style.display = '';
    rankBtn.style.opacity = '0.4';
    if (color) {
      rankBtn.style.background   = color;
      rankBtn.style.borderColor  = color;
      rankBtn.style.color        = 'white';}
    if (icon && icon.includes('.')) {
      rankBtn.innerHTML = `<img src="/img/${projectId}/${icon}" style="width:18px;height:18px;object-fit:contain">`;}}
  // Показываем кнопку tab если toc заполнен
  const tocPair  = pairs.find(p => p.Left === 'tab');
  const tocValue = tocPair ? tocPair.Right.trim() : '';
      if (tocValue) {
        const tocBtn = document.getElementById('tocToggleBtn');
        const icon   = (pairs.find(p => p.Left === 'icon')  || {}).Right || '';
        const color  = (pairs.find(p => p.Left === 'color') || {}).Right || '';
        tocBtn.style.display = '';
        tocBtn.style.opacity = '0.4';
        if (color) {
          tocBtn.style.background  = color;
          tocBtn.style.borderColor = color;
          tocBtn.style.color       = 'white';}
        if (icon && icon.includes('.')) {
          tocBtn.innerHTML = `<img src="/img/${projectId}/${icon}" style="width:18px;height:18px;object-fit:contain">`;}}
  // 
  const chronoPair  = pairs.find(p => p.Left === 'chrono');
  const chronoValue = chronoPair ? chronoPair.Right.trim() : '';
  if (chronoValue) {
    document.getElementById('chronoToggleBtn').style.display = '';}
  // Линзы — передаём контекст и колбэк синхронизации
  const ctx = {
    projectId,
    frameIndex,
    getPairs:       () => pairs,
    getMetaPairs:   () => metaPairs,
    getAliasBefore: () => aliasBefore,
    getToc:    () => (pairs.find(p => p.Left === 'tab')   || {}).Right || '',
    getISort:  () => (pairs.find(p => p.Left === 'i-sort') || {}).Right || '',
    getASort: () => {
    const p = metaPairs.find(p => p.Left === 'a-sort');
    return p ? p.Right : '';},
    getTag: () => (pairs.find(p => p.Left === 'rank') || {}).Right || '', 
    onSave: (freshPairs) => {
      pairs      = freshPairs;
      savedPairs = JSON.parse(JSON.stringify(freshPairs));
      const ap   = pairs.find(p => p.Left === 'alias');
      aliasBefore = ap ? ap.Right.trim().split(/\s+/).filter(Boolean) : [];
      render();},};
  initTagsLens(ctx);
  initContentLens(ctx);
  initRankLens(ctx);
  initTocLens(ctx);
  initChronoLens(ctx);    
  const slotsBtn = document.getElementById('slotsToggleBtn');
  const slotsContainer = document.getElementById('slotsContainer');
  slotsBtn.addEventListener('click', () => {
    if (slotsContainer.classList.contains('hidden')) {
      slotsContainer.classList.remove('hidden');
      slotsBtn.textContent = '⊞';
      slotsBtn.classList.add('active');
    } else {
      slotsContainer.classList.add('hidden');
      slotsBtn.textContent = '⊞';
      slotsBtn.classList.remove('active');}});
  // Автооткрытие линз из o-lens метафрейма
  const oLensPair = metaPairs.find(p => p.Left === 'o-lens');
  const oLens     = oLensPair ? oLensPair.Right.trim() : '';
//  if (oLens) {
if (oLens && !params.get('open')) {    
    const lensMap = {
    'graph':   'dagToggleBtn',
    'tags':    'tagsToggleBtn',
    'content': 'contentToggleBtn',
    'rank':    'rankToggleBtn',
    'tab':     'tocToggleBtn',
    'chrono':  'chronoToggleBtn',
    'slots':   'slotsToggleBtn',};
    setTimeout(() => {
      oLens.split(/\s+/).forEach(name => {
        const btnId = lensMap[name];
        if (!btnId) return;
        const btn = document.getElementById(btnId);
        if (btn && btn.style.display !== 'none') {
          btn.click();}});}, 100);}
const openLens = params.get('open');
if (openLens) {
  setTimeout(() => {
    const btnId = openLens === 'tab' ? 'tocToggleBtn' : 'rankToggleBtn';
    const cId   = openLens === 'tab' ? 'tocContainer' : 'rankContainer';
    const btn   = document.getElementById(btnId);
    const c     = document.getElementById(cId);
    if (btn && btn.style.display !== 'none' && c && c.classList.contains('hidden')) btn.click();}, 400);}
// Создать новый фрейм
document.getElementById('btnNewFrame').addEventListener('click', () => {
document.getElementById('inputFrameTitle').value = '';
document.getElementById('panelCreateFrame').classList.add('open');
document.getElementById('inputFrameTitle').focus();});
document.getElementById('btnCloseCreateFrame').addEventListener('click', () => {
document.getElementById('panelCreateFrame').classList.remove('open');});
document.getElementById('btnSaveCreateFrame').addEventListener('click', async () => {
const title = document.getElementById('inputFrameTitle').value.trim();
if (!title) return;
document.getElementById('panelCreateFrame').classList.remove('open');
const f = await api.createFrame(projectId, title);
if (f.index) window.location.href = 'frame.html?id=' + projectId + '&frame=' + f.index;});
// Клонировать фрейм
document.getElementById('btnCloneFrame').addEventListener('click', async () => {
const f = await api.createFrame(projectId, '');
if (!f.index) return;
const fresh = await api.getFrame(projectId, f.index);
const newPairs = fresh.pairs.map(p => {
const src = pairs.find(s => s.Left === p.Left);
if (!src) return p;
if (['id','title','modified','url'].includes(p.Left)) return p;
return { ...p, Right: src.Right };});
await api.saveFrame(projectId, f.index, newPairs, []);
window.location.href = 'frame.html?id=' + projectId + '&frame=' + f.index;}); }
// --- Graph ---
//const graph = new ParseqGraph('dagContainer', projectId);
let graph = null;
const dagBtn = document.getElementById('dagToggleBtn');
const dagContainer = document.getElementById('dagContainer');
let graphLoaded = false;
dagBtn.addEventListener('click', () => {
	const isHidden = dagContainer.classList.contains('hidden');
	if (isHidden) {
		dagContainer.classList.remove('hidden');
		dagBtn.classList.add('active');
        if (!graphLoaded && graph) { graphLoaded = true; graph.loadAndRender(frameIndex); }
	} else {
		dagContainer.classList.add('hidden');
		dagBtn.classList.remove('active'); } });
init();