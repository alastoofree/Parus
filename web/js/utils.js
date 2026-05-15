// web/js/utils.js — общие утилиты и API

export const GAUSS = ['o','i','z','t','y','v','s','l','a','n','x','e','d','b','h','f','m','g'];
export function toGauss(n, len) {
	let r = '';
	for (let i = 0; i < len; i++) { r = GAUSS[n % 18] + r; n = Math.floor(n / 18); }
	return r; }
export function gaussOrder(s) {
	return s.split('').reduce((acc, c) => acc * 18 + 'oiztyvslanxedbhfmg'.indexOf(c), 0); }
export function esc(s) {
	return String(s || '').replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/</g,'&lt;'); }
export function mdToHtml(md, projectId = '') {
	const codeBlocks = [];
	md = md.replace(/```([\s\S]+?)```/g, (_, code) => {
		codeBlocks.push('<pre><code>' + code.replace(/^\n/,'').replace(/\n$/,'') + '</code></pre>');
		return `\x00CODE${codeBlocks.length - 1}\x00`;});
// 
const svgBlocks = [];
	md = md.replace(/<svg[\s\S]*?<\/svg>/gi, match => {
		svgBlocks.push(match);
		return `\x00SVG${svgBlocks.length - 1}\x00`;});
const inlineFmt = s => s
	.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_, alt, src) => {
		if (src.match(/\.(mp3|ogg|wav|m4a)$/i)) {
			const audioSrc = src.startsWith('http') ? src : `/img/${projectId}/${src}`;
			return `<audio controls style="width:100%"><source src="${audioSrc}"></audio>`;}
		if (src.includes('google.com/maps')) {
			return `<iframe src="${src}" style="width:100%;aspect-ratio:4/3;border:none" allowfullscreen></iframe>`;}
		if (src.includes('youtube.com') || src.includes('youtu.be')) {
			const m = src.match(/(?:v=|youtu\.be\/)([^&?]+)/);
			const id = m ? m[1] : '';
			return `<iframe src="https://www.youtube.com/embed/${id}" style="width:100%;aspect-ratio:16/9;border:none" allowfullscreen></iframe>`;}
		return src.startsWith('http')
			? `<img src="${src}" alt="${alt}" style="max-width:100%;height:auto">`
			: `<img src="/img/${projectId}/${src}" alt="${alt}" style="max-width:100%;height:auto">`;})
	.replace(/\[([^\]]+)\]\((https?:\/\/[^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>')
    .replace(/\[([^\]]+)\]\((frame\.html[^)]+)\)/g, '<a href="$2" target="_blank">$1</a>')
	.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
	.replace(/\*(.+?)\*/g,     '<em>$1</em>')
	.replace(/`(.+?)`/g,       '<code>$1</code>');
const lines = md.split('\n');
let html = '', inList = false, inTable = false;
for (let line of lines) {
	if (/^\|(.+)\|$/.test(line)) {
		if (inList)   { html += '</ul>';   inList  = false; }
		if (!inTable) { html += '<table style="border-collapse:collapse;width:100%">'; inTable = true; }
		if (/^\|[-| :]+\|$/.test(line)) { continue; }
		const cells = line.split('|').filter((_, i, a) => i > 0 && i < a.length - 1);
		html += '<tr>' + cells.map(c => `<td style="border:1px solid var(--border);padding:4px 8px">${inlineFmt(c.trim())}</td>`).join('') + '</tr>';
		continue;
	} else if (inTable) { html += '</table>'; inTable = false; }
	if (/^### (.+)$/.test(line)) {
		if (inList) { html += '</ul>'; inList = false; }
		html += '<h3>' + line.slice(4) + '</h3>';
	} else if (/^## (.+)$/.test(line)) {
		if (inList) { html += '</ul>'; inList = false; }
		html += '<h2>' + line.slice(3) + '</h2>';
	} else if (/^# (.+)$/.test(line)) {
		if (inList) { html += '</ul>'; inList = false; }
		html += '<h1>' + line.slice(2) + '</h1>';
	} else if (/^\- (.+)$/.test(line)) {
		if (!inList) { html += '<ul>'; inList = true; }
		html += '<li>' + inlineFmt(line.slice(2)) + '</li>';
	} else if (/^\d+\. (.+)$/.test(line)) {
		if (!inList) { html += '<ol>'; inList = true; }
		html += '<li>' + inlineFmt(line.replace(/^\d+\. /,'')) + '</li>';
	} else if (/^---+$/.test(line.trim())) {
		if (inList) { html += '</ul>'; inList = false; }
		html += '<hr>';
	} else if (line.trim() === '') {
		if (inList)  { html += '</ul>';   inList  = false; }
		if (inTable) { html += '</table>'; inTable = false; }
		html += '<br>';
	} else {
		if (inList) { html += '</ul>'; inList = false; }
		html += '<p>' + inlineFmt(line) + '</p>';}}
if (inList)  html += '</ul>';
if (inTable) html += '</table>';
html = html.replace(/\x00CODE(\d+)\x00/g, (_, i) => codeBlocks[parseInt(i)]);
html = html.replace(/\x00SVG(\d+)\x00/g, (_, i) => svgBlocks[parseInt(i)]);
return html; }
// renderFrameContent — универсальный рендер контента фрейма
// Резолвит трансклюзии, рендерит маркдаун с картинками, видео, картами
export async function renderFrameContent(content, contentType, projectId) {
	const parts = content.split(/(\{\{[^}]+\}\})/);
	let html = '';
	for (const part of parts) {
		const m = part.match(/^\{\{([^}]+)\}\}$/);
		if (m) {
			try {
				const r = await fetch('/api/projects/' + projectId + '/content/' + m[1]);
				const d = await r.json();
				if (d.contentType === 'svg') {
    html += `<span class="svg-placeholder" data-svg="${encodeURIComponent(d.contentRaw || d.content || '')}"></span>`;
				} else {
					html += mdToHtml(d.content || '', projectId);}
			} catch(e) {}
		} else if (part.trim()) {
            if (contentType === 'svg') {
                html += part;
            } else if (contentType === 'html') {
                html += part;
            } else {
                html += mdToHtml(part, projectId);}}}
	if (contentType === 'svg') {
		return '<div class="content-view content-svg">' + html + '</div>';}
	return '<div class="content-view content-markdown">' + html + '</div>';}
// 
export function buildCharOrder(aSort) {
	const order = { ' ': -1, '-': -1, ',': -1 };
	let pos = 0;
	for (const c of aSort.toLowerCase()) {
		if (c === ' ') continue;
		order[c] = pos++; }
	return order; }
// 
export function compareByASort(a, b, order) {
	const ra = [...a.toLowerCase()];
	const rb = [...b.toLowerCase()];
	const maxLen = Math.max(ra.length, rb.length);
	for (let i = 0; i < maxLen; i++) {
		if (i >= ra.length) return -1;
		if (i >= rb.length) return 1;
		const oa = order[ra[i]] !== undefined ? order[ra[i]] : 10000;
		const ob = order[rb[i]] !== undefined ? order[rb[i]] : 10000;
		if (oa !== ob) return oa - ob; }
	return 0; }
// 
export const api = {
	// Проекты
	async getProjects() {
		const r = await fetch('/api/projects');
		return r.json(); },
	async getProject(id) {
		const r = await fetch('/api/projects/' + id);
		return r.json(); },
	async createProject(title, fileOrderLen) {
		const r = await fetch('/api/projects', {
			method:  'POST',
			headers: { 'Content-Type': 'application/json' },
			body:    JSON.stringify({ title, fileOrderLen }), });
		return r.json(); },
    async updateProject(id, cfg) {
        const r = await fetch('/api/projects/' + id, {
            method:  'PUT',
            headers: { 'Content-Type': 'application/json' },
            body:    JSON.stringify(cfg), });
        return r.json(); },
    async getChildren(projectId, frameIndex, iSort = '', aSort = '') {
        const r = await fetch(`/api/projects/${projectId}/children/${frameIndex}?isort=${encodeURIComponent(iSort)}&asort=${encodeURIComponent(aSort)}`);
        return r.json(); },
    async getChrono(projectId, frameIndex) {
        const r = await fetch(`/api/projects/${projectId}/chrono/${frameIndex}`);
        return r.json(); },
	// Фреймы
	async getFrames(projectId) {
		const r = await fetch('/api/projects/' + projectId + '/frames');
		return r.json(); },
	async getFrame(projectId, frameIndex) {
		const r = await fetch('/api/projects/' + projectId + '/frames/' + frameIndex);
		return r.json(); },
    async searchFrame(projectId, frameIndex, query) {
    const r = await fetch(`/api/projects/${projectId}/search/${frameIndex}?q=${encodeURIComponent(query)}`);
    return r.json(); },
	async createFrame(projectId, title) {
		const r = await fetch('/api/projects/' + projectId + '/frames', {
			method:  'POST',
			headers: { 'Content-Type': 'application/json' },
			body:    JSON.stringify({ title }), });
		return r.json(); },
	async saveFrame(projectId, frameIndex, pairs, aliasBefore) {
		const r = await fetch('/api/projects/' + projectId + '/frames/' + frameIndex, {
			method:  'PUT',
			headers: { 'Content-Type': 'application/json' },
			body:    JSON.stringify({ pairs, aliasBefore }), });
		return r.json(); },
	// Теги и контент
	async getTags(projectId) {
		const r = await fetch('/api/projects/' + projectId + '/tags');
		return r.json(); },
//     
	async getGlobalTags() {
		const r    = await fetch('/api/global/tags');
		const data = await r.json();
		return (data || []).map(t => ({
			index: t.Index || t.index || '',
			title: t.Title || t.title || '',
			rank:  t.Rank  || t.rank  || '', })); }, };
// 
export async function loadSvgIcon(projectId, icon) {
	if (!icon || icon.includes('.')) return null;
	try {
		const r = await fetch('/api/projects/' + projectId + '/content/' + icon);
		const d = await r.json();
		if (d.contentType === 'svg' && d.contentRaw) return d.contentRaw;
	} catch(e) {}
	return null; }
// 
export function toast(msg, isErr) {
	const el = document.getElementById('toast');
	if (!el) return;
	el.textContent = msg;
	el.style.background = isErr ? 'var(--accent2)' : 'var(--ink)';
	el.classList.add('show');
	setTimeout(() => el.classList.remove('show'), 2600); }