// parus/web/terms/js/terms.js — терминологическая база

import { esc, toast, buildCharOrder, compareByASort } from '../utils.js';

const TERM_ASORT = 'nmlrhgkcsztdbpfvwuoaeiyjqx нмлркхгжщчшсзцтдбпфвюуоёэеиыйяаьъ';
const termCharOrder = buildCharOrder(TERM_ASORT);
let terms = [];

async function loadTerms() {
  const r = await fetch('/api/terms');
  terms = await r.json();
  return terms; }
// 
function sortTerms(arr) {
  arr.sort((a, b) => compareByASort(a.name || '', b.name || '', termCharOrder));
  return arr; }
// 
function render() {
  const list = document.getElementById('termsList');
  if (terms.length === 0) {
    list.innerHTML = '<div class="empty">not terms</div>';
    return; }
  const nTags = terms.filter(t => t.rank && t.rank !== '').length;
  document.getElementById('termCounts').textContent = terms.length + ' (' + nTags + ')';
  list.innerHTML = terms.map(t => `
    <div class="base-item" data-path="${esc(t.path)}">
      <div class="base-left">
        <span class="base-favicon-svg" data-path="${esc(t.path)}"></span>
        <span class="base-name">${esc(t.name)}</span>
      </div>
      ${t.rank ? '<span class="base-item-tag">⬡</span>' : ''}
      ${t.uri ? `<span class="base-meta" style="font-size:0.65rem;color:var(--ink3)">${esc(t.uri.replace('KATER/', ''))}</span>` : ''}
    </div>`).join(''); }
// 
async function loadIcons() {
  document.querySelectorAll('.base-favicon-svg').forEach(async el => {
    const path = el.dataset.path;
    try {
      const r    = await fetch('/api/terms/satellite-svg?path=' + encodeURIComponent(path));
      const data = await r.json();
      if (data.svg) el.innerHTML = data.svg;
    } catch(e) {}});}
// 
function renderQBtns() {
  document.querySelectorAll('.btn-q-lens').forEach(b => b.remove());
//  const btns = terms.filter(t => t.rank && t.qBtns);
  const btns = terms.filter(t => t.qBtns);
  if (!btns.length) return;
  const headerDiv = document.querySelector('header div');
  btns.forEach(t => {
    const modes = t.qBtns.trim().split(/\s+/).filter(Boolean);
    modes.forEach(mode => {
      if (mode !== 'rank' && mode !== 'tab' && mode !== 'chrono' && mode !== 'process') return;
      const btn = document.createElement('button');
      btn.className = 'btn-icon btn-q-lens';
      btn.textContent = mode === 'tab' ? '↕' : mode === 'chrono' ? '⊕' : mode === 'process' ? '⊛' : '◈';
      btn.title = t.name + ' : ' + mode;
      btn.style.cssText = 'height:32px;border-radius:12px;padding:2px 12px;font-size:0.82rem;font-family:Spectral,serif;cursor:pointer;margin-left:0.5rem;border:1px solid var(--border);background:var(--bg2);color:var(--ink)';
      if (t.color) {btn.style.borderColor = t.color; btn.style.color = t.color;}
      btn.addEventListener('click', () => {
        window.location.href = 'term.html?path=' + encodeURIComponent(t.path) + '&lens=' + mode;});
      headerDiv.insertBefore(btn, document.getElementById('btnCreateTerm'));});}); }
// 
document.getElementById('termsList').addEventListener('click', e => {
  const item = e.target.closest('.base-item');
  if (item) window.location.href = 'term.html?path=' + encodeURIComponent(item.dataset.path); });
// Create
document.getElementById('btnCreateTerm').addEventListener('click', () => {
  document.getElementById('inputTermName').value = '';
  document.getElementById('panelCreateTerm').classList.add('open');
  document.getElementById('inputTermName').focus(); });
// 
function closeCreate() {
  document.getElementById('panelCreateTerm').classList.remove('open'); }
// 
document.getElementById('btnCloseCreateTerm').addEventListener('click', closeCreate);
document.getElementById('panelCreateTerm').addEventListener('click', e => {
  if (e.target === e.currentTarget) closeCreate(); });
// 
document.getElementById('btnSaveCreateTerm').addEventListener('click', async () => {
  const name = document.getElementById('inputTermName').value.trim();
  if (!name) return;
  const r = await fetch('/api/terms', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({name})});
  const d = await r.json();
  if (d.message) {toast(d.message, true); return;}
  closeCreate();
  window.location.href = 'term.html?path=' + encodeURIComponent(d.path); });
// Filter + Search
document.getElementById('cliInput').addEventListener('input', e => {
  const q = e.target.value.trim().toLowerCase();
  if (!q) {render(); return;}
  const filtered = terms.filter(t => (t.name || '').toLowerCase().includes(q));
  const list = document.getElementById('termsList');
  if (!filtered.length) {list.innerHTML = '<div class="empty">not found</div>'; return;}
  list.innerHTML = filtered.map(t => `
    <div class="base-item" data-path="${esc(t.path)}">
      <div class="base-left">
        <span class="base-favicon-svg" data-path="${esc(t.path)}"></span>
        <span class="base-name">${esc(t.name)}</span>
      </div>
      ${t.rank ? '<span class="base-item-tag">⬡</span>' : ''}
      ${t.uri ? `<span class="base-meta" style="font-size:0.65rem;color:var(--ink3)">${esc(t.uri.replace('KATER/', ''))}</span>` : ''}
    </div>`).join('');});
// 
document.getElementById('cliInput').addEventListener('keydown', async e => {
  if (e.key !== 'Enter') return;
  const q = e.target.value.trim();
  if (!q) return;
  const list = document.getElementById('termsList');
  try {
    const r = await fetch('/api/terms/search?q=' + encodeURIComponent(q));
    const results = await r.json();
    if (!results || results.length === 0) {list.innerHTML = '<div class="empty">not found</div>'; return;}
    list.innerHTML = results.map(t => `
  <div class="base-item" data-path="${esc(t.path)}">
    <div class="base-left" style="flex-direction:column;align-items:flex-start">
      <span class="base-name">${esc(t.name)}</span>
      ${t.excerpt ? `<div style="font-size:0.75rem;color:var(--ink3);white-space:normal;margin-top:2px">${esc(t.excerpt).replace(new RegExp(esc(q).replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi'), m => '<mark>' + m + '</mark>')}</div>` : ''}
    </div>
  </div>`).join('');
    list.querySelectorAll('.base-item').forEach(item => {
      item.addEventListener('click', () => {
        window.location.href = 'term.html?path=' + encodeURIComponent(item.dataset.path);});});
  } catch(e) {list.innerHTML = '<div class="empty">error</div>';}});
// Init
async function init() {
  await loadTerms();
  sortTerms(terms);
  render(); 
  loadIcons();
  renderQBtns(); }
init();