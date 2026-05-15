// parus/web/js/lens-chrono.js — линза хронологической таблицы

import { api, esc } from './utils.js';

export function initChronoLens(ctx) {
  const { projectId, frameIndex, getPairs } = ctx;
  const btn       = document.getElementById('chronoToggleBtn');
  const container = document.getElementById('chronoContainer');
  if (!btn || !container) return;
  const chronoValue = (getPairs().find(p => p.Left === 'chrono') || {}).Right || '';
  if (!chronoValue) return;
  btn.addEventListener('click', () => {
    if (!container.classList.contains('hidden')) {
      container.classList.add('hidden');
      btn.classList.remove('active');
      return;}
    container.classList.remove('hidden');
    btn.classList.add('active');
    if (!container.dataset.loaded) {
      renderChrono(container);
      container.dataset.loaded = '1';}});
// 
async function renderChrono(container) {
    container.innerHTML = '<div style="padding:16px;color:var(--ink3)">load...</div>';
    try {
      const rows = await api.getChrono(projectId, frameIndex);
      if (!rows || rows.length === 0) {
        container.innerHTML = '<div style="padding:16px;color:var(--ink3)">not frames</div>';
        return;}
      let titleMap = {};
      try {
        const allFrames = await api.getFrames(projectId);
        allFrames.forEach(f => { titleMap[f.index] = f.title; });
      } catch(e) {}
      function renderCell(val) {
        if (!val) return '';
        const extLink = val.match(/^\[\[(.+?)\|(https?:\/\/.+?)\]\]$/);
        if (extLink) return `<a href="${esc(extLink[2])}" target="_blank" rel="noopener">${esc(extLink[1])}</a>`;
        if (val.startsWith('http://') || val.startsWith('https://')) return `<a href="${esc(val)}" target="_blank" rel="noopener">${esc(val)}</a>`;
        if (titleMap[val]) return `<a href="frame.html?id=${esc(projectId)}&frame=${esc(val)}">${esc(titleMap[val])}</a>`;
        return esc(val);}
      const extraKeys = rows[0] ? Object.keys(rows[0].extras) : [];
      let html = `<div style="overflow-x:auto;padding:8px">
      <table style="width:100%;border-collapse:collapse;font-size:0.9rem">
      <thead><tr style="border-bottom:2px solid var(--border)">
      <th style="padding:8px;text-align:left">фрейм</th>
      <th style="padding:8px;text-align:left">${esc(rows[0]?.dateKey || 'date')}</th>`;
      extraKeys.forEach(col => { html += `<th style="padding:8px;text-align:left">${esc(col)}</th>`; });
      html += `</tr></thead><tbody>`;
      rows.forEach(row => {
        html += `<tr style="border-bottom:1px solid var(--border)">
        <td style="padding:8px"><a href="frame.html?id=${esc(projectId)}&frame=${esc(row.frameIndex)}"
          style="font-weight:bold;color:var(--ink);text-decoration:none">${esc(row.frameTitle)}</a></td>
        <td style="padding:8px">${renderCell(row.dateVal)}</td>`;
        extraKeys.forEach(col => { html += `<td style="padding:8px">${renderCell(row.extras[col] || '')}</td>`; });
        html += `</tr>`;});
      html += `</tbody></table></div>`;
      container.innerHTML = html;
    } catch(e) {
      container.innerHTML = '<div style="padding:16px;color:var(--accent2)">error load</div>';}} }