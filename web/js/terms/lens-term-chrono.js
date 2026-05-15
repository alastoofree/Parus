// parus/web/js/terms/lens-term-chrono.js — линза хроно для терминологической базы

import { esc } from '../utils.js';

export function initTermChronoLens(ctx) {
  const { frameIndex, getPairs } = ctx;
  const btn       = document.getElementById('chronoToggleBtn');
  const container = document.getElementById('chronoContainer');
  if (!btn || !container) return;
  const chronoVal  = (getPairs().find(p => p.Left === 'chrono')   || {}).Right || '';
  const iChronoVal = (getPairs().find(p => p.Left === 'i-chrono') || {}).Right || '';
  if (!chronoVal && !iChronoVal) return;
  btn.style.display = '';
  const color = (getPairs().find(p => p.Left === 'color') || {}).Right || '';
  if (color) {btn.style.borderColor = color;}
  btn.style.opacity = '0.4';
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
      renderChrono(container);
      container.dataset.loaded = '1';}});
// 
  async function renderChrono(container) {
    container.innerHTML = '<div style="padding:16px;color:var(--ink3)">load...</div>';
    try {
      const r    = await fetch('/api/terms/chrono?path=' + encodeURIComponent(frameIndex));
      const rows = await r.json();
      if (!rows || rows.length === 0) {
        container.innerHTML = '<div style="padding:16px;color:var(--ink3)">not data</div>';
        return;}
//         
      function renderCell(val) {
        if (!val) return '';
        const extLink = val.match(/^\[(.+?)\]\((https?:\/\/.+?)\)$/);
        if (extLink) return `<a href="${esc(extLink[2])}" target="_blank" rel="noopener">${esc(extLink[1])}</a>`;
        if (val.startsWith('http://') || val.startsWith('https://')) {
          return `<a href="${esc(val)}" target="_blank" rel="noopener">${esc(val)}</a>`;}
        if (val.startsWith('<a ')) return val;
        return esc(val);}
      const extraKeys = rows[0] ? Object.keys(rows[0].extras) : [];
      let html = `<div style="overflow-x:auto;padding:8px">
        <table style="width:100%;border-collapse:collapse;font-size:0.9rem">
        <thead><tr style="border-bottom:2px solid var(--border)">
        <th style="padding:8px;text-align:left">date</th>
        <th style="padding:8px;text-align:left">терм</th>`;
      extraKeys.forEach(col => {html += `<th style="padding:8px;text-align:left">${esc(col)}</th>`;});
      html += `</tr></thead><tbody>`;
      rows.forEach(row => {
        html += `<tr style="border-bottom:1px solid var(--border)">
          <td style="padding:8px">${renderCell(row.dateVal)}</td>
          <td style="padding:8px"><a href="term.html?path=${encodeURIComponent(row.termPath)}"
            style="font-weight:bold;color:${row.termColor || 'var(--ink)'};text-decoration:none">${esc(row.termName)}</a></td>`;
        extraKeys.forEach(col => {html += `<td style="padding:8px">${renderCell(row.extras[col] || '')}</td>`;});
        html += `</tr>`;});
      html += `</tbody></table></div>`;
      container.innerHTML = html;
    } catch(e) {
      container.innerHTML = '<div style="padding:16px;color:var(--accent2)">error load</div>';}} }