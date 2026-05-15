// /parus/web/js/portal.js — PARADIGMA: список тем

import { api, esc } from './utils.js';

let projects   = [];
let selectedId = null;
 
async function loadProjectMeta(p) {
  try {
    const data  = await api.getProject(p.id);
    const pairs = data.pairs || [];
    const counts = (pairs.find(x => x.Left === 'counts') || {}).Right || '';
    const c     = counts.trim().split(/\s+/);
    p.nFrames   = c[0] || '0';
    p.nTags     = c[1] || '0';
  } catch(e) {
    p.nFrames = '?';
    p.nTags   = '?';}
  const exts = ['png','svg','ico','jpg','jpeg'];
  for (const ext of exts) {
    const url = '/img/' + p.id + '/favicon.' + ext + '?v=' + Date.now();
    const r   = await fetch(url, { method: 'HEAD' });
    if (r.ok) { p.favicon = url; break; }}
  return p; }
// 
function render() {
  const list = document.getElementById('basesList');
  if (projects.length === 0) {
    list.innerHTML = '<div class="empty">not themes</div>';
  } else {
    list.innerHTML = projects.map(p => `
      <div class="base-item ${p.id === selectedId ? 'active' : ''}" data-id="${p.id}" title="${p.id === 'meta-base' ? 'F: ' + (p.nFrames||'?') : 'F(T): ' + (p.nFrames||'?') + ' (' + (p.nTags||'?') + ')'}">
      
        <div class="base-left">
          ${p.favicon
            ? `<img class="base-favicon" src="${p.favicon}"
                onerror="this.style.display='none'">`
            : '<span class="base-favicon-placeholder"></span>'}
          <span class="base-name">${p.title || '—'}</span>
        </div>
        <span class="base-counts" title="${p.id === 'meta-base' ? 'F: ' + (p.nFrames||'?') : 'F(T): ' + (p.nFrames||'?') + ' (' + (p.nTags||'?') + ')'}"></span>
      </div>
    `).join('');} }
// 
document.getElementById('basesList').addEventListener('click', e => {
const item = e.target.closest('.base-item');
if (item) window.location.href = 'project.html?id=' + item.dataset.id;});
document.getElementById('searchInput').addEventListener('keydown', async e => {
    if (e.key !== 'Enter') return;
    const q = e.target.value.trim();
    if (!q) {render(); return;}
    const r = await fetch(`/api/global/fullsearch?q=${encodeURIComponent(q)}`);
    const results = await r.json();
    const list = document.getElementById('basesList');
    if (!results.length) {list.innerHTML = '<div class="empty">ничего не найдено</div>'; return;}
    list.innerHTML = results.map(item => `
        <div class="base-item" data-id="${item.baseId}" data-frame="${item.frameIndex}">
            <div class="base-left">
                <span class="base-index">${esc(item.baseTitle)}</span>
                <span class="base-name">${esc(item.frameTitle)}</span>
            </div>
            ${item.tagTitle ? `<div class="search-excerpt">${esc(item.tagTitle)}</div>` : ''}
        </div>`).join('');
    list.querySelectorAll('.base-item').forEach(item => {
        item.addEventListener('click', () => {
            window.location.href = 'frame.html?id=' + item.dataset.id + '&frame=' + item.dataset.frame;});});});
// Create
document.getElementById('btnCreate').addEventListener('click', () => {
document.getElementById('panelLabel').textContent   = 'New theme';
document.getElementById('inputTitle').value         = '';
document.getElementById('panelCreate').classList.add('open');
document.getElementById('inputTitle').focus();});
// 
function closeCreate() {
document.getElementById('panelCreate').classList.remove('open'); }
//
document.getElementById('btnCloseCreate').addEventListener('click', closeCreate);
const btnCancelCreate = document.getElementById('btnCancelCreate');
if (btnCancelCreate) btnCancelCreate.addEventListener('click', closeCreate);
document.getElementById('panelCreate').addEventListener('click', e => {
if (e.target === e.currentTarget) closeCreate();});
// 
document.getElementById('btnSaveCreate').addEventListener('click', async () => {
const title        = document.getElementById('inputTitle').value.trim() || 'Not title';
const fileOrderLen = document.getElementById('inputFileOrder').value.trim().length || 2;
const p = await api.createProject(title, fileOrderLen);
const enriched = await loadProjectMeta(p);
projects.push(enriched);
selectedId = p.id;
closeCreate();
render();});
// Init
async function init() {
  const raw = await api.getProjects();
  projects  = await Promise.all(raw.map(p => loadProjectMeta(p)));
  projects.sort((a, b) => a.id === 'meta-base' ? -1 : b.id === 'meta-base' ? 1 : 0);
  render(); }
init();