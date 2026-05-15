// parus/web/js/project.js — PARADIGMA: концепты темы

import { api, esc, loadSvgIcon, buildCharOrder, compareByASort, gaussOrder, toast } from './utils.js';
import { initTocLens } from './lens-tab.js';
import { initRankLens } from './lens-rank.js';
const params    = new URLSearchParams(location.search);
const projectId = params.get('id');
if (!projectId) window.location.href = 'portal.html';
let frames        = [];
let projectTitle  = '';
let projectASort  = '';
let projectOLens  = '';
let projectRanks  = '';
let projectTabs   = '';
let projectWest   = '';
let projectNord   = '';
let projectGraph  = '';
let projectList   = '';
let projectChrono = '';
let projectDeepNord  = '';
let projectDeepSouth = '';
// 
function sortFrames(arr) {
	if (projectASort) {
		const order = buildCharOrder(projectASort);
		arr.sort((a, b) => compareByASort(a.title || '', b.title || '', order));
	} else {
		arr.sort((a, b) => gaussOrder(a.index) - gaussOrder(b.index)); }
	return arr; }
// 
function render() {
	const list = document.getElementById('framesList');
	if (frames.length === 0) { list.innerHTML = '<div class="empty">not concepts</div>'; return; }
	list.innerHTML = frames.map(f => `
		<div class="base-item" data-id="${f.index}">
			<div class="base-left">
				<span class="base-index">${f.index}</span>
				${f.icon
					? f.icon.includes('.')
						? `<img class="base-item-icon" src="/img/${projectId}/${f.icon}" onerror="this.style.display='none'">`
						: `<span class="base-item-svg" data-svg-index="${f.icon}" style="width:16px;height:16px;display:inline-flex"></span>`
					: ''}
				<span class="base-name">${f.title || '—'}</span>
			</div>
			${f.isTag && projectId !== 'meta-base' ? '<span class="base-item-tag">⬡</span>' : ''}
		</div>`).join('');
	document.querySelectorAll('.base-item-svg').forEach(async el => {
		const svg = await loadSvgIcon(projectId, el.dataset.svgIndex);
		if (svg) el.innerHTML = svg; }); }
document.getElementById('framesList').addEventListener('click', e => {
	const item = e.target.closest('.base-item');
	if (item) window.location.href = 'frame.html?id=' + projectId + '&frame=' + item.dataset.id; });
// Create frame
document.getElementById('btnCreateFrame').addEventListener('click', () => {
	document.getElementById('inputFrameTitle').value = '';
	document.getElementById('panelCreateFrame').classList.add('open');
	document.getElementById('inputFrameTitle').focus(); });
// 
function closeCreateFrame() { document.getElementById('panelCreateFrame').classList.remove('open'); }
// 
document.getElementById('btnCloseCreateFrame').addEventListener('click', closeCreateFrame);
const btnCancelCreateFrame = document.getElementById('btnCancelCreateFrame');
if (btnCancelCreateFrame) btnCancelCreateFrame.addEventListener('click', closeCreateFrame);
document.getElementById('panelCreateFrame').addEventListener('click', e => {
	if (e.target === e.currentTarget) closeCreateFrame(); });
document.getElementById('btnSaveCreateFrame').addEventListener('click', async () => {
	const title = document.getElementById('inputFrameTitle').value.trim();
	const f = await api.createFrame(projectId, title);
	if (!f.index) { toast(f.message || 'Не удалось создать концепт', true); return; }
	frames = await api.getFrames(projectId);
	sortFrames(frames);
	closeCreateFrame();
	render(); });
// Settings
function openSettings() {
	document.getElementById('settingsId').textContent  = projectId;
	document.getElementById('settingsTitle').value     = projectTitle;
	document.getElementById('settingsASort').value     = projectASort;
	document.getElementById('settingsOLens').value     = projectOLens;
	document.getElementById('settingsRanks').value     = projectRanks;
	document.getElementById('settingsTabs').value      = projectTabs;
	document.getElementById('settingsWest').value      = projectWest;
	document.getElementById('settingsNord').value      = projectNord;
    document.getElementById('settingsGraph').value     = projectGraph;
    document.getElementById('settingsList').value      = projectList;
    document.getElementById('settingsChrono').value    = projectChrono;
    document.getElementById('settingsDeepNord').value  = projectDeepNord;
    document.getElementById('settingsDeepSouth').value = projectDeepSouth;
	document.getElementById('panelSettings').classList.add('open');
	document.getElementById('settingsTitle').focus(); }
// 
function closeSettings() { document.getElementById('panelSettings').classList.remove('open'); }
// 
document.getElementById('btnSettings').addEventListener('click', openSettings);
document.getElementById('btnCloseSettings').addEventListener('click', closeSettings);
const btnCancelSettings = document.getElementById('btnCancelSettings');
if (btnCancelSettings) btnCancelSettings.addEventListener('click', closeSettings);
document.getElementById('panelSettings').addEventListener('click', e => {
	if (e.target === e.currentTarget) closeSettings(); });
// 
document.getElementById('btnSaveSettings').addEventListener('click', async () => {
	const title  = document.getElementById('settingsTitle').value.trim() || projectTitle;
	const aSort  = document.getElementById('settingsASort').value.trim();
	const oLens  = document.getElementById('settingsOLens').value.trim();
	const ranks  = document.getElementById('settingsRanks').value.trim();
	const tabs   = document.getElementById('settingsTabs').value.trim();
	const west   = document.getElementById('settingsWest').value.trim();
	const nord   = document.getElementById('settingsNord').value.trim();
    const graph  = document.getElementById('settingsGraph').value.trim();
    const list   = document.getElementById('settingsList').value.trim();
    const chrono = document.getElementById('settingsChrono').value.trim();
    const deepNord  = document.getElementById('settingsDeepNord').value.trim();
    const deepSouth = document.getElementById('settingsDeepSouth').value.trim();
    await api.updateProject(projectId, { title, aSort, oLens, ranks, tabs, west, nord, graph, list, chrono, deepNord, deepSouth });
	projectTitle     = title;
	projectRanks     = ranks;
	projectTabs      = tabs;
	projectASort     = aSort;
	projectOLens     = oLens;
	projectWest      = west;
	projectNord      = nord;
	projectGraph     = graph;
	projectList      = list;
	projectChrono    = chrono;
    projectDeepNord  = deepNord;
    projectDeepSouth = deepSouth;
	document.getElementById('breadcrumbProject').textContent = title;
	closeSettings();
	renderLensButtons(); });
// Rebuild
document.getElementById('btnRebuild').addEventListener('click', async () => {
    const btn = document.getElementById('btnRebuild');
    btn.style.opacity = '0.4';
    btn.disabled = true;
    try {
        await fetch('/api/projects/' + projectId + '/rebuild', {method: 'POST'});
        frames = await api.getFrames(projectId);
        sortFrames(frames);
        render();
        btn.textContent = '✓';
        toast('Rebuild завершён');
        setTimeout(() => {btn.textContent = '↺';}, 3000);
    } catch(e) {
        btn.textContent = '✗';
        toast('Ошибка: ' + e.message, true);
        setTimeout(() => {btn.textContent = '↺';}, 3000);
    } finally {
        btn.style.opacity = '1';
        btn.disabled = false;}});
// Lens buttons
async function openLens(idx, lensType, btn) {
	const container = document.getElementById('baseLensContainer');
	const key = lensType + ':' + idx;
	if (container.dataset.active === key) {
		container.innerHTML = '';
		container.dataset.active = '';
		btn.style.background = '#6b6760';
		return; }
	document.querySelectorAll('.btn-lens-shortcut').forEach(b => b.style.background = '#6b6760');
	btn.style.background = '#2a5caa';
	container.dataset.active = key;
	if (lensType === 'tab') {
		window.location.href = 'frame.html?id=' + projectId + '&frame=' + idx + '&open=tab';
	} else {
		window.location.href = 'frame.html?id=' + projectId + '&frame=' + idx + '&open=rank'; } }
// 
function renderLensButtons() {
	document.querySelectorAll('.btn-lens-shortcut').forEach(b => b.remove());
	const header   = document.querySelector('header');
	const tabList  = projectTabs.trim().split(/\s+/).filter(Boolean);
	const rankList = projectRanks.trim().split(/\s+/).filter(Boolean);
//     
function applyFrameStyle(btn, idx) {
	api.getFrame(projectId, idx).then(frame => {
		const color = (frame.pairs || []).find(p => p.Left === 'color')?.Right || '';
		const icon  = (frame.pairs || []).find(p => p.Left === 'icon')?.Right  || '';
		if (color) { btn.style.background = color; btn.style.borderColor = color; }
		if (icon && icon.includes('.')) btn.innerHTML = `<img src="/img/${projectId}/${icon}" style="width:18px;height:18px;object-fit:contain">`; }); }
	tabList.forEach(idx => {
		const f   = frames.find(f => f.index === idx);
		const btn = document.createElement('button');
		btn.className = 'btn-icon btn-lens-shortcut';
		btn.title = (f ? f.title : idx) + ' : tab';
		btn.textContent = '≡';
		btn.addEventListener('click', () => openLens(idx, 'tab', btn));
		btn.style.cssText = 'width:32px;height:32px;border-radius:50%;background:#6b6760;border-color:#6b6760;color:white;font-size:0.85rem;margin-left:0.5rem';
//		header.appendChild(btn);
        const breadcrumbs = document.querySelector('header .breadcrumbs');
        breadcrumbs.insertBefore(btn, document.getElementById('btnCreateFrame'));
		applyFrameStyle(btn, idx); });
	rankList.forEach(idx => {
		const f   = frames.find(f => f.index === idx);
		const btn = document.createElement('button');
		btn.className = 'btn-icon btn-lens-shortcut';
		btn.title = (f ? f.title : idx) + ' : rank';
		btn.textContent = '◈';
		btn.addEventListener('click', () => openLens(idx, 'rank', btn));
		btn.style.cssText = 'width:32px;height:32px;border-radius:50%;background:#6b6760;border-color:#6b6760;color:white;font-size:0.85rem;margin-left:0.5rem';
//		header.appendChild(btn);
        const breadcrumbs = document.querySelector('header .breadcrumbs');
        breadcrumbs.insertBefore(btn, document.getElementById('btnCreateFrame'));
		applyFrameStyle(btn, idx); }); }
// Search
document.getElementById('cliInput').addEventListener('input', e => {
	const q = e.target.value.trim().toLowerCase();
	if (!q) { render(); return; }
	const terms    = q.split('&').map(t => t.trim()).filter(Boolean);
	const filtered = frames.filter(f => terms.every(t => (f.title || '').toLowerCase().includes(t)));
	const list     = document.getElementById('framesList');
	if (!filtered.length) { list.innerHTML = '<div class="empty">ничего не найдено</div>'; return; }
	list.innerHTML = filtered.map(f => `
		<div class="base-item" data-id="${f.index}">
			<div class="base-left">
				<span class="base-index">${f.index}</span>
				<span class="base-name">${highlight(f.title || '—', terms)}</span>
			</div>
			${f.isTag && projectId !== 'meta-base' ? '<span class="base-item-tag">⬡</span>' : ''}
		</div>`).join('');
	list.querySelectorAll('.base-item').forEach(item => {
		item.addEventListener('click', () => {
			window.location.href = 'frame.html?id=' + projectId + '&frame=' + item.dataset.id; }); }); });
// Import/Export/Convert/Union/Search
document.getElementById('cliInput').addEventListener('keydown', async e => {
	if (e.key !== 'Enter') return;
	const val = e.target.value.trim();
	e.target.value = '';
	render();
	if (!val) return;
	const parts = val.split(/\s+/);
	const cmd   = parts[0].toLowerCase();
	if (cmd === 'import' || cmd === 'export') {
		const baseId  = parts[1];
		const indices = parts.slice(2);
		if (!baseId || !indices.length) return;
		const endpoint = cmd === 'import'
			? `/api/projects/${projectId}/import`
			: `/api/projects/${projectId}/export`;
		const body = cmd === 'import'
			? {from: baseId, frames: indices}
			: {to: baseId, frames: indices};
		const r = await fetch(endpoint, {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify(body)});
		const d = await r.json();
		const n = d.imported ?? d.exported ?? 0;
		toast(`${cmd}: ${n} frames`);
		frames = await api.getFrames(projectId);
		sortFrames(frames);
		render();
	} else if (cmd === 'convert') {
		const pageUrl = parts[1];
		if (!pageUrl) return;
		toast('Конвертация...');
		const r = await fetch(`/api/projects/${projectId}/convert`, {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify({url: pageUrl})});
		const d = await r.json();
		if (d.message) {toast('Ошибка: ' + d.message, true); return;}
		toast(`convert: ${d.imported} frames`);
		frames = await api.getFrames(projectId);
		sortFrames(frames);
		render();
    } else if (cmd === 'union') {
        const idx = parts[1];
        if (!idx) return;
        toast('Объединение...');
        const r = await fetch(`/api/projects/${projectId}/union`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({index: idx})});
        const d = await r.json();
        if (d.message) {toast('Ошибка: ' + d.message, true); return;}
        toast('union: готово');
        frames = await api.getFrames(projectId);
        sortFrames(frames);
        render();
    } else {
        const sr = await fetch(`/api/projects/${projectId}/search?q=${encodeURIComponent(val)}`);
        const results = await sr.json();
        const list = document.getElementById('framesList');
        if (!results.length) {list.innerHTML = '<div class="empty">ничего не найдено</div>'; return;}
        list.innerHTML = results.map(f => `
            <div class="base-item" data-id="${f.index}">
                <div class="base-left">
                    <span class="base-index">${f.index}</span>
                    <span class="base-name">${highlight(f.title || '—', [val])}</span>
                </div>
                ${f.excerpt ? `<div class="search-excerpt">${highlight(f.excerpt, [val])}</div>` : ''}
            </div>`).join('');
        list.querySelectorAll('.base-item').forEach(item => {
            item.addEventListener('click', () => {
                window.location.href = 'frame.html?id=' + projectId + '&frame=' + item.dataset.id;});});}});
// 
function highlight(text, terms) {
	let result = esc(text);
	terms.forEach(t => {
		const re = new RegExp(t.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi');
		result = result.replace(re, m => `<mark>${m}</mark>`); });
	return result; }
// Init
async function init() {
	const p             = await api.getProject(projectId);
	const titlePair     = (p.pairs || []).find(pair => pair.Left === 'title');
	const aSortPair     = (p.pairs || []).find(pair => pair.Left === 'a-sort');
	const oLensPair     = (p.pairs || []).find(pair => pair.Left === 'o-lens');
	const ranksPair     = (p.pairs || []).find(pair => pair.Left === 'ranks');
	const tabsPair      = (p.pairs || []).find(pair => pair.Left === 'tabs');
	const westPair      = (p.pairs || []).find(pair => pair.Left === 'west');
	const nordPair      = (p.pairs || []).find(pair => pair.Left === 'nord');
    const graphPair     = (p.pairs || []).find(pair => pair.Left === 'graph');
	const listPair      = (p.pairs || []).find(pair => pair.Left === 'list');
	const chronoPair    = (p.pairs || []).find(pair => pair.Left === 'chrono');
	const countsPair    = (p.pairs || []).find(pair => pair.Left === 'counts');
    const deepNordPair  = (p.pairs || []).find(pair => pair.Left === 'deep-nord');
    const deepSouthPair = (p.pairs || []).find(pair => pair.Left === 'deep-south');
	projectTitle  = (titlePair && titlePair.Right) ? titlePair.Right : (projectId === '00000-META-BASE' ? 'META-BASE' : '—');
	projectASort     = aSortPair  ? aSortPair.Right  : '';
	projectOLens     = oLensPair  ? oLensPair.Right  : '';
	projectRanks     = ranksPair  ? ranksPair.Right  : '';
	projectTabs      = tabsPair   ? tabsPair.Right   : '';
	projectWest      = westPair   ? westPair.Right   : '';
	projectNord      = nordPair   ? nordPair.Right   : '';
    projectGraph     = graphPair  ? graphPair.Right  : '';
	projectList      = listPair   ? listPair.Right   : '';
    projectChrono    = chronoPair ? chronoPair.Right : '';
    projectDeepNord  = deepNordPair  ? deepNordPair.Right  : '';
    projectDeepSouth = deepSouthPair ? deepSouthPair.Right : '';
	document.getElementById('breadcrumbProject').textContent = projectTitle;
	if (countsPair) {
		const counts  = countsPair.Right.trim().split(/\s+/);
		const nFrames = counts[0] || '0';
		const nTags   = counts[1] || '0';
		document.getElementById('projectCounts').textContent =
			projectId === 'meta-base' ? 'F: ' + nFrames : 'F(T): ' + nFrames + ' (' + nTags + ')'; }
	const favicon = document.getElementById('favicon');
	const exts = ['png','svg','ico','jpg','jpeg'];
	for (const ext of exts) {
        const url = '/img/' + projectId + '/favicon.' + ext + '?v=' + Date.now();
		const r   = await fetch(url, { method: 'HEAD' });
		if (r.ok) { favicon.href = url; break; } }
	frames = await api.getFrames(projectId);
	sortFrames(frames);
	render();
	renderLensButtons(); 
    if (projectId === '00000-META-BASE') {
        const nordRow = document.getElementById('settingsNordRow');
        if (nordRow) nordRow.style.display = 'none'; }}
init();