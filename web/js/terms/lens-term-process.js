// parus/web/js/terms/lens-term-process.js — линза процесса для терминологической базы

import { esc } from '../utils.js';

export function initTermProcessLens(ctx) {
  const { frameIndex, getPairs } = ctx;
  const btn       = document.getElementById('processToggleBtn');
  const container = document.getElementById('processContainer');
  if (!btn || !container) return;
  const processVal = (getPairs().find(p => p.Left === 'process') || {}).Right || '';
  if (!processVal) return;
  const color = (getPairs().find(p => p.Left === 'color') || {}).Right || '';
  btn.style.display = '';
  btn.style.opacity = '0.4';
  if (color) {btn.style.borderColor = color;}
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
      renderProcess(container);
      container.dataset.loaded = '1';}});

  function showAliasTooltip(evt, aliases) {
    document.querySelectorAll('.dag-tooltip').forEach(el => el.remove());
    if (!aliases || aliases.length === 0) return;
    const tip = document.createElement('div');
    tip.className = 'dag-tooltip';
    const ul = document.createElement('ul');
    ul.className = 'dag-tooltip-list';
    aliases.forEach(a => {
      const li = document.createElement('li');
      const anchor = document.createElement('a');
      anchor.href = '#';
      anchor.textContent = a.name;
      anchor.onclick = e => {
        e.preventDefault();
        window.location.href = 'term.html?path=' + encodeURIComponent(a.path);
        tip.remove();};
      li.appendChild(anchor);
      ul.appendChild(li);});
    tip.appendChild(ul);
    tip.style.left = (evt.clientX + 12) + 'px';
    tip.style.top  = evt.clientY + 'px';
    document.body.appendChild(tip);
    const close = () => {tip.remove(); document.removeEventListener('click', close);};
    setTimeout(() => document.addEventListener('click', close), 0);}

  async function renderProcess(container) {
    container.innerHTML = '<div style="padding:16px;color:var(--ink3)">load...</div>';
    try {
      const r    = await fetch('/api/terms/process?path=' + encodeURIComponent(frameIndex));
      const rows = await r.json();
      if (!rows || rows.length === 0) {
        container.innerHTML = '<div style="padding:16px;color:var(--ink3)">not data</div>';
        return;}
      function renderTerm(name, path, aliases) {
        const hasAliases = aliases && aliases.length > 0;
        const nameHtml = path
          ? `<a href="term.html?path=${encodeURIComponent(path)}" style="text-decoration:none;color:inherit">${esc(name)}</a>`
          : esc(name);
        if (!hasAliases) return nameHtml;
        return nameHtml + ` <span class="dag-trigger-arrow active" style="cursor:pointer;font-size:0.7rem" data-alias-trigger>▶</span>`;}
      let html = `<div style="overflow-x:auto;padding:8px">
        <table style="width:100%;border-collapse:collapse;font-size:0.9rem">
        <thead><tr style="border-bottom:2px solid var(--border)">
        <th style="padding:8px;text-align:left">№</th>
        <th style="padding:8px;text-align:left">объект</th>
        <th style="padding:8px;text-align:left">препозиция</th>
        <th style="padding:8px;text-align:left">объект</th>
        </tr></thead><tbody>`;
      rows.forEach(row => {
        html += `<tr style="border-bottom:1px solid var(--border)">
          <td style="padding:8px;color:var(--ink3)">${esc(row.key)}</td>
          <td style="padding:8px" data-aliases='${JSON.stringify(row.aliasesA || [])}'>${renderTerm(row.termA, row.pathA, row.aliasesA)}</td>
          <td style="padding:8px" data-aliases='${JSON.stringify(row.aliasesP || [])}'>${renderTerm(row.prep,  row.pathP, row.aliasesP)}</td>
          <td style="padding:8px" data-aliases='${JSON.stringify(row.aliasesB || [])}'>${renderTerm(row.termB, row.pathB, row.aliasesB)}</td>
        </tr>`;});
      html += `</tbody></table></div>`;
      container.innerHTML = html;
      container.querySelectorAll('[data-alias-trigger]').forEach(trigger => {
        trigger.addEventListener('click', e => {
          e.preventDefault();
          e.stopPropagation();
          const td = trigger.closest('td');
          const aliases = JSON.parse(td.dataset.aliases || '[]');
          showAliasTooltip(e, aliases);});});
    } catch(e) {
      container.innerHTML = '<div style="padding:16px;color:var(--accent2)">error load</div>';}} }