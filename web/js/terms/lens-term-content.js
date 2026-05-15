// parus/web/js/terms/lens-term-content.js — линза контента для терминологической базы

import { esc, renderFrameContent } from '../utils.js';
export function initTermContentLens(ctx) {
  const { frameIndex, getPairs, onSave } = ctx;
// 
  document.getElementById('contentToggleBtn').addEventListener('click', async () => {
    const container = document.getElementById('contentContainer');
    const btn       = document.getElementById('contentToggleBtn');
    if (!container.classList.contains('hidden')) {
      container.classList.add('hidden');
      btn.textContent = '≡';
      btn.classList.remove('active');
      return;}
    container.classList.remove('hidden');
    btn.textContent = '≡';
    btn.classList.add('active');
    await loadContentView(container);});
//     
  async function loadContentView(container) {
    container.innerHTML = '<div style="padding:16px;font-family:monospace;color:#666">load...</div>';
    try {
      const r    = await fetch('/api/terms/content?path=' + encodeURIComponent(frameIndex));
      const data = await r.json();
      showContentView(container, data.content || '', data.contentType || 'markdown', data.contentRaw || '');
    } catch(e) {
      container.innerHTML = '<div style="padding:16px;color:red">error load</div>';} }
// 
  async function showContentView(container, content, contentType, raw) {
    const hasContent    = content.trim().length > 0;
    const effectiveType = contentType || 'markdown';
    const mainHtml = hasContent ? await renderFrameContent(content, effectiveType, '') : '';
    let productHtml = '';
    try {
      const rp   = await fetch('/api/terms/product?path=' + encodeURIComponent(frameIndex));
      const prod = await rp.json();
      if (prod && Object.keys(prod).length > 0) {
        productHtml = '<hr style="margin:12px 0;border:none;border-top:1px solid var(--border)">';
        productHtml += '<div class="content-product" style="font-size:0.9rem">';
        for (const [key, val] of Object.entries(prod)) {
          productHtml += `<div style="margin-bottom:4px"><strong>${esc(key)}</strong>: ${esc(val)}</div>`;}
        productHtml += '</div>';}}
    catch(e) {}
    const rendered = (mainHtml + productHtml) || '<div class="content-empty">not content</div>';
    container.innerHTML = `
      <div class="content-panel">
        <div class="content-panel-toolbar" style="display:flex;justify-content:flex-end;gap:0.4rem">
          <button class="content-btn" id="contentJsonBtn" title="satellite json"
            style="width:28px;height:28px;border-radius:50%;background:#e8a838;border-color:#e8a838;color:white;font-size:0.75rem">{ }</button>
          <button class="content-btn" id="contentSvgBtn" title="satellite svg"
            style="width:28px;height:28px;border-radius:50%;background:#e8a838;border-color:#e8a838;color:white;font-size:0.75rem">◻</button>
          <button class="content-btn" id="contentEditBtn" title="redaction"
            style="width:28px;height:28px;border-radius:50%;background:#e8a838;border-color:#e8a838;color:white">✎</button>
        </div>
        <div class="content-panel-body">${rendered}</div>
      </div>`;
    const bodyEl = container.querySelector('.content-panel-body');
    if (bodyEl) await applyDict(bodyEl);
    document.getElementById('contentEditBtn').addEventListener('click', () => {
      showContentEdit(container, raw || content, contentType);}); 
    document.getElementById('contentJsonBtn').addEventListener('click', () => {
      showSatelliteEdit(container, 'json');});
    document.getElementById('contentSvgBtn').addEventListener('click', () => {
      showSatelliteEdit(container, 'svg');}); }
// 
  function showContentEdit(container, content, contentType) {
    container.innerHTML = `
      <div class="content-panel">
        <div class="content-panel-toolbar" style="display:flex;justify-content:flex-end;align-items:center;gap:0.5rem">
          <button class="content-btn" id="contentSaveBtn" title="record"
            style="width:28px;height:28px;border-radius:50%;background:#2a7a2a;border-color:#2a7a2a;color:white">✓</button>
        </div>
        <div class="content-panel-body">
          <textarea id="contentTextarea" class="content-textarea"
            spellcheck="false">${esc(content)}</textarea>
        </div>
      </div>`;
    document.getElementById('contentSaveBtn').addEventListener('click', async () => {
      const newContent = document.getElementById('contentTextarea').value;
      const newType    = 'markdown';
      const pairs      = getPairs();
      for (let p of pairs) {
        if (p.Left === 'content')      {p.Right = newContent;}}
      await fetch('/api/terms/' + encodeURIComponent(frameIndex), {
        method: 'PUT',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({pairs})});
      const fresh = await fetch('/api/terms/' + encodeURIComponent(frameIndex));
      const data  = await fresh.json();
      onSave(data.pairs || []);
      await loadContentView(container);}); } 
// 
  async function showSatelliteEdit(container, type) {
    const url = type === 'json'
      ? '/api/terms/satellite?path='      + encodeURIComponent(frameIndex)
      : '/api/terms/satellite-svg?path=' + encodeURIComponent(frameIndex);
    let current = '';
    try {
      const r    = await fetch(url);
      const data = await r.json();
      current = type === 'json'
        ? JSON.stringify(data, null, 2)
        : data.svg || '';}
    catch(e) {}
    container.innerHTML = `
      <div class="content-panel">
        <div class="content-panel-toolbar" style="display:flex;justify-content:flex-end;gap:0.4rem">
          <button class="content-btn" id="satSaveBtn" title="record"
            style="width:28px;height:28px;border-radius:50%;background:#2a7a2a;border-color:#2a7a2a;color:white">✓</button>
        </div>
        <div class="content-panel-body">
          <textarea id="satTextarea" class="content-textarea"
            spellcheck="false">${esc(current)}</textarea>
        </div>
      </div>`;
    document.getElementById('satSaveBtn').addEventListener('click', async () => {
      const val = document.getElementById('satTextarea').value;
      if (type === 'json') {
        let parsed;
        try {parsed = JSON.parse(val);}
        catch(e) {alert('invalid JSON'); return;}
        await fetch('/api/terms/satellite?path=' + encodeURIComponent(frameIndex), {
          method: 'PUT',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify(parsed)});
      } else {
        await fetch('/api/terms/satellite-svg?path=' + encodeURIComponent(frameIndex), {
          method: 'PUT',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({svg: val})});}
      await loadContentView(container);}); }
// 
  async function applyDict(bodyEl) {
    const pairs  = getPairs();
    const depth  = parseInt((pairs.find(p => p.Left === 'dict')   || {}).Right || '-1');
    const iDict  = (pairs.find(p => p.Left === 'i-dict') || {}).Right || '';
    if (depth < 0 || !iDict) return;
    const dictPaths = iDict.trim().split(/\s+/).filter(Boolean);
    if (!dictPaths.length) return;
    const dictEntries = [];
    for (const dpath of dictPaths) {
      try {
        const r    = await fetch('/api/terms/satellite?path=' + encodeURIComponent(dpath));
        const data = await r.json();
        for (const [key, val] of Object.entries(data)) {
          const cleanVal = String(val).replace(/\[([^\]]+)\]\([^)]+\)/g, '$1');
          dictEntries.push({keyTokens: key.trim().split(/\s+/), original: key, value: cleanVal});}}
      catch(e) {}}
    if (!dictEntries.length) return;
    const walker = document.createTreeWalker(bodyEl, NodeFilter.SHOW_TEXT);
    const nodes  = [];
    let node;
    while (node = walker.nextNode()) nodes.push(node);
    for (const textNode of nodes) {
      const parent = textNode.parentNode;
      if (!parent || parent.closest('code, pre, a')) continue;
      const result = matchDict(textNode.textContent, dictEntries, depth);
      if (!result) continue;
      const span = document.createElement('span');
      span.innerHTML = result;
      parent.replaceChild(span, textNode);} }
// 
  function tokensMatch(a, b, depth) {
    a = a.toLowerCase(); b = b.toLowerCase();
    if (depth === 0) return a === b;
    const minLen = Math.min(a.length, b.length);
    const prefixLen = Math.max(minLen - depth, 1);
    return a.slice(0, prefixLen) === b.slice(0, prefixLen); }
// 
  function phraseMatch(textTokens, keyTokens, depth) {
    if (textTokens.length !== keyTokens.length) return false;
    return textTokens.every((t, i) => tokensMatch(t, keyTokens[i], depth)); }
// 
  function matchDict(text, dictEntries, depth) {
    const words  = text.split(/(\s+)/);
    let result   = '', changed = false, wi = 0;
    while (wi < words.length) {
      if (/^\s+$/.test(words[wi])) {result += words[wi]; wi++; continue;}
      let matched = false;
      for (const {keyTokens, original, value} of dictEntries) {
        const len = keyTokens.length;
        const phrase = [], indices = [];
        let k = wi;
        while (phrase.length < len && k < words.length) {
          if (!/^\s+$/.test(words[k])) {phrase.push(words[k]); indices.push(k);}
          k++;}
        if (phrase.length < len) continue;
        if (phraseMatch(phrase, keyTokens, depth)) {
          result += `<span style="border-bottom:1px dotted var(--ink);cursor:help" title="${esc(original)}: ${esc(value)}">${esc(words.slice(wi, k).join(''))}</span>`;
          wi = k; matched = true; changed = true; break;}}
      if (!matched) {result += words[wi]; wi++;}}
    return changed ? result : null; } }