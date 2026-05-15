// parus/web/term/js/term.js — страница термина

import { esc, toast, toGauss, renderFrameContent, buildCharOrder, compareByASort } from '../utils.js';
import { TermGraph } from './graph-terms.js';
import { initTermTagsLens } from './lens-term-tags.js';
import { initTermContentLens } from './lens-term-content.js';
import { initTermProcessLens } from './lens-term-process.js';
import { initTermChronoLens } from './lens-term-chrono.js';
import { initTermRankLens } from './lens-term-rank.js';
import { initTermTocLens } from './lens-term-tab.js';

const termApi = {
  getChildren: async (path, iSort = '', aSort = '') => {
    const r = await fetch(`/api/terms/children?path=${encodeURIComponent(path)}&isort=${encodeURIComponent(iSort)}&asort=${encodeURIComponent(aSort)}`);
    return r.json();},
  getTags: async () => {
    const r = await fetch('/api/terms/tags');
    return r.json();},
  getFrames: async () => {
    const r = await fetch('/api/terms');
    return (await r.json()).map(t => ({index: t.path, title: t.name, isTag: false}));} };
const params   = new URLSearchParams(location.search);
const termPath = params.get('path');
if (!termPath) window.location.href = 'terms.html';
const TERM_ASORT    = 'nmlrhgkcsztdbpfvwuoaeiyjqx нмлркхгжщчшсзцтдбпфвюуоёэеиыйяаьъ';
const termCharOrder = buildCharOrder(TERM_ASORT);
const READONLY_KEYS = ['uri', 'content', 'tags', 'links', 'includes', 'counts', 'themes'];
const READONLY_LEFT_KEYS = ['name', 'alias', 'uri', 'rank', 'tags', 'links', 'color', 'content', 'includes', 
  'i-sort', 'tab', 'dict', 'i-dict', 'deep', 'graph', 'list', 'west', 'nord', 'chrono', 'i-chrono', 
  'show-product', 'product', 'attribute', 'preposition', 'prefix', 'suffix', 'counts', 'themes', 
  'reserv-1', 'reserv-2', 'reserv-3', 'reserv-4', 'reserv-5', 'i-process', 'process', 'q-btns'];
const HIDDEN_KEYS = ['content', 'tags', 'links', 'includes', 'themes', 'product', 
                     'reserv-1', 'reserv-2', 'reserv-3', 'reserv-4', 'reserv-5'];
let pairs      = [];
let savedPairs = [];
let termGraph  = null;
let graphLoaded = false;
let cloneMode  = false;
let clonePairs = [];

function render() {
  document.getElementById('slotsList').innerHTML = pairs.map((p, i) => ({p, i})).filter(({p}) => !HIDDEN_KEYS.includes(p.Left)).map(({p, i}) => {
    const ro     = READONLY_KEYS.includes(p.Left);
    const roLeft = ro || READONLY_LEFT_KEYS.includes(p.Left);
    return `
      <div class="slot-row${ro ? ' readonly' : ''}">
        <span class="slot-row-index">${esc(p.TypeIndex)}</span>
        <div class="slot-cell key${roLeft ? ' readonly' : ''}">
          <input type="text" value="${esc(p.Left)}" data-i="${i}" data-side="Left"
            ${roLeft ? 'readonly tabindex="-1"' : ''}
            autocomplete="off" spellcheck="false" placeholder="slot">
        </div>
        <div class="slot-cell${ro ? ' readonly' : ''}">
          ${p.Left === 'counts'
            ? `<div style="font-family:monospace;font-size:0.8rem;padding:2px 4px;word-break:break-all">${esc(p.Right)}</div>`
            : `<input type="text" value="${esc(p.Right)}" data-i="${i}" data-side="Right"
            ${ro ? 'readonly tabindex="-1"' : ''}
            autocomplete="off" spellcheck="false" placeholder="value">`}
        </div>
      </div>`;}).join(''); }
// 
document.getElementById('slotsList').addEventListener('input', e => {
  const input = e.target.closest('input');
  if (!input) return;
  pairs[parseInt(input.dataset.i)][input.dataset.side] = input.value; });
// 
document.getElementById('btnAddSlot').addEventListener('click', () => {
  const userSlots = pairs.filter(p => p.TypeIndex.length === 4);
  pairs.push({TypeIndex: toGauss(userSlots.length, 4), Left: '', Right: ''});
  render();
  const inputs = document.getElementById('slotsList').querySelectorAll('input[data-side="Left"]');
  const last = inputs[inputs.length - 1];
  if (last) last.focus(); });
// 
document.getElementById('btnSave').addEventListener('click', async () => {
  if (cloneMode) {
    const name = (pairs.find(p => p.Left === 'name') || {}).Right || '';
    if (!name) {toast('введите имя', true); return;}
    const uri = (pairs.find(p => p.Left === 'uri') || {}).Right || '';
    const r = await fetch('/api/terms', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({name, sourcePath: termPath, targetUri: uri})});
    const d = await r.json();
    if (!d.path) {toast(d.message || 'error', true); return;}
    await fetch('/api/terms/' + encodeURIComponent(d.path), {
      method: 'PUT',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({pairs})});
    window.location.href = 'term.html?path=' + encodeURIComponent(d.path);
    return;}
  const r = await fetch('/api/terms/' + encodeURIComponent(termPath), {
    method: 'PUT',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({pairs})});
  if (!r.ok) {toast('error save', true); return;}
  savedPairs = JSON.parse(JSON.stringify(pairs));
  const btn = document.getElementById('btnSave');
  btn.style.background = '#2a7a2a';
  btn.style.borderColor = '#2a7a2a';
  setTimeout(() => {
    btn.style.background = '#e8a838';
    btn.style.borderColor = '#e8a838';}, 1500);
  render(); });
// 
document.getElementById('btnNewTerm').addEventListener('click', () => {
  document.getElementById('inputTermName').value = '';
  document.getElementById('panelCreateTerm').classList.add('open');
  document.getElementById('inputTermName').focus(); });
// 
document.getElementById('btnCloseCreateTerm').addEventListener('click', () => {
  document.getElementById('panelCreateTerm').classList.remove('open'); });
// 
document.getElementById('btnSaveCreateTerm').addEventListener('click', async () => {
  const name = document.getElementById('inputTermName').value.trim();
  if (!name) return;
  document.getElementById('panelCreateTerm').classList.remove('open');
  const r = await fetch('/api/terms', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({name})});
  const d = await r.json();
  if (d.message) {toast(d.message, true); return;}
  window.location.href = 'term.html?path=' + encodeURIComponent(d.path); });
// 
document.getElementById('btnCloneTerm').addEventListener('click', () => {
  if (cloneMode) {
    cloneMode = false;
    pairs = JSON.parse(JSON.stringify(savedPairs));
    document.getElementById('btnCloneTerm').style.background = '#2a5caa';
    document.getElementById('btnCloneTerm').style.borderColor = '#2a5caa';
    render();
    return;}
  cloneMode = true;
  clonePairs = JSON.parse(JSON.stringify(pairs));
  document.getElementById('btnCloneTerm').style.background = '#e8a838';
  document.getElementById('btnCloneTerm').style.borderColor = '#e8a838'; });
// 
function toggleLens(btnId, containerId) {
  const btn = document.getElementById(btnId);
  const container = document.getElementById(containerId);
  if (!btn || !container) return;
  btn.addEventListener('click', () => {
    const hidden = container.classList.contains('hidden');
    if (hidden) {
      container.classList.remove('hidden');
      btn.classList.add('active');
    } else {
      container.classList.add('hidden');
      btn.classList.remove('active'); }}); }
// 
toggleLens('slotsToggleBtn', 'slotsContainer');
// 
const dagBtn       = document.getElementById('dagToggleBtn');
const dagContainer = document.getElementById('dagContainer');
dagBtn.addEventListener('click', () => {
  const hidden = dagContainer.classList.contains('hidden');
  if (hidden) {
    dagContainer.classList.remove('hidden');
    dagBtn.classList.add('active');
    if (!graphLoaded && termGraph) {graphLoaded = true; termGraph.loadAndRender();}
  } else {
    dagContainer.classList.add('hidden');
    dagBtn.classList.remove('active');}});
const contentBtn       = document.getElementById('contentToggleBtn');
const contentContainer = document.getElementById('contentContainer');
// 
async function init() {
  const r = await fetch('/api/terms/' + encodeURIComponent(termPath));
  if (!r.ok) {window.location.href = 'terms.html'; return;}
  const data = await r.json();
  pairs = data.pairs || [];
  const color = (pairs.find(p => p.Left === 'color') || {}).Right || '';
  if (color) {
    const rankBtn = document.getElementById('rankToggleBtn');
    const tocBtn  = document.getElementById('tocToggleBtn');
    const processBtn = document.getElementById('processToggleBtn');
    if (processBtn) processBtn.style.borderColor = color;
    if (rankBtn) rankBtn.style.borderColor = color;
    if (tocBtn)  tocBtn.style.borderColor  = color;
    document.getElementById('breadcrumbTerm').style.color = color;}
    const lensParam = new URLSearchParams(location.search).get('lens');
    if (lensParam === 'process') {
        setTimeout(() => document.getElementById('processToggleBtn')?.click(), 150);}
    if (lensParam === 'rank') {
        setTimeout(() => document.getElementById('rankToggleBtn')?.click(), 150);}
    if (lensParam === 'tab') {
        setTimeout(() => document.getElementById('tocToggleBtn')?.click(), 150);}
    if (lensParam === 'chrono') {
        setTimeout(() => document.getElementById('chronoToggleBtn')?.click(), 150);}
  savedPairs = JSON.parse(JSON.stringify(pairs));
  const name = (pairs.find(p => p.Left === 'name') || {}).Right || termPath;
  document.getElementById('breadcrumbTerm').textContent = name;
  document.title = name + ' — TERMINUS';
  const tabPair = pairs.find(p => p.Left === 'tab');
  if (tabPair && tabPair.Right.trim()) {
    document.getElementById('tocToggleBtn').style.display = '';}
  render();
  const ctx = {
    projectId:      termPath,
    frameIndex:     termPath,
    getPairs:       () => pairs,
    getMetaPairs:   () => [],
    getAliasBefore: () => [],
    getToc:    () => (pairs.find(p => p.Left === 'tab')    || {}).Right || '',
    getISort:  () => (pairs.find(p => p.Left === 'i-sort') || {}).Right || '',
    getASort:  () => TERM_ASORT,
    getTag: () => (pairs.find(p => p.Left === 'rank') || {}).Right || '',
    onSave: (freshPairs) => {
      pairs      = freshPairs;
      savedPairs = JSON.parse(JSON.stringify(freshPairs));
      render();},
    api: termApi,};
  initTermTagsLens(ctx);
  initTermContentLens(ctx);
  initTermProcessLens(ctx);
  initTermChronoLens(ctx);
  initTermRankLens(ctx);
  initTermTocLens(ctx);
  const graphParam = (pairs.find(p => p.Left === 'graph') || {}).Right || '';
  const listParam  = (pairs.find(p => p.Left === 'list')  || {}).Right || '';
  termGraph = new TermGraph('dagContainer', termPath, graphParam, listParam);
  setTimeout(() => {
    dagBtn.click();
    document.getElementById('tagsToggleBtn').click();
    document.getElementById('contentToggleBtn').click(); }, 100); }
document.getElementById('cliInput').addEventListener('keydown', async e => {
  if (e.key !== 'Enter') return;
  const query = e.target.value.trim();
  e.target.value = '';
  if (!query) return;
  const r = await fetch('/api/terms/search?q=' + encodeURIComponent(query));
  const results = await r.json();
  const container = document.getElementById('searchResults');
  if (!results || results.length === 0) {
    container.innerHTML = '<div class="search-empty">not found</div>';
    container.classList.remove('hidden');
    return;}
  container.innerHTML = results.map(r => `
    <div class="search-result">
      <a href="term.html?path=${encodeURIComponent(r.path)}">${esc(r.name)}</a>
      ${r.excerpt ? `<div class="search-excerpt">${esc(r.excerpt)}</div>` : ''}
    </div>`).join('');
  container.classList.remove('hidden');});
init();