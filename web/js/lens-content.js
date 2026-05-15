// web/js/lens-content.js — линза контента

import { api, esc, toGauss, mdToHtml, renderFrameContent } from './utils.js';

export function initContentLens(ctx) {
  const { projectId, frameIndex, getPairs, getAliasBefore, onSave } = ctx;
  document.getElementById('contentToggleBtn').addEventListener('click', async () => {
    const container = document.getElementById('contentContainer');
    const btn       = document.getElementById('contentToggleBtn');
    if (!container.classList.contains('hidden')) {
      // закрываем
      container.classList.add('hidden');
      btn.textContent = '≡';
      btn.classList.remove('active');
      return;}
    // открываем
    container.classList.remove('hidden');
    btn.textContent = '≡';
    btn.classList.add('active');
    await loadContentView(container);});
// 
  async function loadContentView(container) {
    container.innerHTML =
      '<div style="padding:16px;font-family:monospace;color:#666">load...</div>';
    try {
      const r    = await fetch('/api/projects/' + projectId + '/content/' + frameIndex);
      const data = await r.json();
      showContentView(
        container,
        data.content     || '',
        data.contentType || 'markdown',
        data.contentRaw  || '',
        data.product     || '');
    } catch(e) {
      container.innerHTML = '<div style="padding:16px;color:red">error load</div>';}}
// 
  async function showContentView(container, content, contentType, raw, product) {
    const productHtml = product ? await renderFrameContent(product, 'markdown', projectId) : '';
    const hasContent = content.trim().length > 0;
    const effectiveType = contentType || 'markdown';
    const mainHtml = hasContent
        ? await renderFrameContent(content, effectiveType, projectId)
        : '';
    const rendered = productHtml
        ? (mainHtml ? mainHtml + '<hr class="dag-divider">' : '') + '<div class="content-product">' + productHtml + '</div>'
        : mainHtml || '<div class="content-empty">not content</div>';
    container.innerHTML = `
      <div class="content-panel">
        <div class="content-panel-toolbar" style="display:flex;justify-content:flex-end">
        <button class="content-btn" id="contentEditBtn" title="redaction" style="width:28px;height:28px;border-radius:50%;background:#e8a838;border-color:#e8a838;color:white">✎</button>
        </div>
        <div class="content-panel-body">${rendered}</div>
      </div>`;
    container.querySelectorAll('.svg-placeholder').forEach(el => {
        const svgStr = decodeURIComponent(el.dataset.svg);
        const parser = new DOMParser();
        const doc = parser.parseFromString(svgStr, 'image/svg+xml');
        el.replaceWith(doc.documentElement);});
    const bodyEl = container.querySelector('.content-panel-body');
    if (bodyEl) await applyDict(bodyEl);
    document.getElementById('contentEditBtn').addEventListener('click', () => {
      showContentEdit(container, raw || content, contentType);});}
// 
  function showContentEdit(container, content, contentType) {
    container.innerHTML = `
      <div class="content-panel">
        <div class="content-panel-toolbar" style="display:flex;justify-content:flex-end;align-items:center;gap:0.5rem">
        <select id="contentTypeSelect" title="Content type" style="font-family:'JetBrains Mono',monospace;font-size:0.72rem;border:1px solid var(--border);border-radius:3px;padding:2px 6px;background:var(--bg);color:var(--ink)">
        <option value="markdown" ${contentType==='markdown' || contentType==='' || contentType==='text' ? 'selected':''}>markdown</option>
        <option value="html" ${contentType==='html' ? 'selected':''}>html</option>
        <option value="svg"  ${contentType==='svg'  ? 'selected':''}>svg</option>
        <option value="json" ${contentType==='json' ? 'selected':''}>json</option>
    </select>
        <button class="content-btn" id="contentSaveBtn" title="record" style="width:28px;height:28px;border-radius:50%;background:#2a7a2a;border-color:#2a7a2a;color:white">✓</button>
        </div>
        <div class="content-panel-body">
          <textarea id="contentTextarea" class="content-textarea"
            spellcheck="false">${esc(content)}</textarea>
        </div>
      </div>`;
    document.getElementById('contentSaveBtn').addEventListener('click', async () => {
      const newContent = document.getElementById('contentTextarea').value;
      const newType    = document.getElementById('contentTypeSelect').value;
      const pairs      = getPairs();
      let foundContent = false, foundType = false;
      for (let p of pairs) {
        if (p.Left === 'content')      { p.Right = newContent; foundContent = true; }
        if (p.Left === 'content-type') { p.Right = newType;    foundType    = true; }}
      if (!foundContent) pairs.push({
        TypeIndex: toGauss(pairs.length, pairs[0]?.TypeIndex.length || 2),
        Left: 'content', Right: newContent,});
      if (!foundType) pairs.push({
        TypeIndex: toGauss(pairs.length, pairs[0]?.TypeIndex.length || 2),
        Left: 'content-type', Right: newType,});
      await api.saveFrame(projectId, frameIndex, pairs, getAliasBefore());
      const fresh = await api.getFrame(projectId, frameIndex);
      onSave(fresh.pairs || []);
      await loadContentView(container);});
    document.getElementById('contentCancelBtn').addEventListener('click', async () => {
      await loadContentView(container);});}
// 
  async function applyDict(bodyEl) {
    const pairs   = getPairs();
    const depth   = parseInt((pairs.find(p => p.Left === 'dict')   || {}).Right || '-1');
    const iDict   = (pairs.find(p => p.Left === 'i-dict') || {}).Right || '';
    if (depth < 0 || !iDict) return;
    const dictIds = iDict.trim().split(/\s+/).filter(Boolean);
    if (!dictIds.length) return;
    // Строим карту: усечённый_ключ → {original, value}
    const dictEntries = [];
    for (const did of dictIds) {
      try {
        const f       = await api.getFrame(projectId, did);
        const content = (f.pairs.find(p => p.Left === 'content') || {}).Right || '{}';
        const data    = JSON.parse(content);
        for (const [key, val] of Object.entries(data)) {
          dictEntries.push({
            keyTokens: key.trim().split(/\s+/),
            original:  key,
            value:     String(val)});}
      } catch(e) {}}
    if (!dictEntries.length) return;
    // Обходим текстовые узлы
    const walker = document.createTreeWalker(bodyEl, NodeFilter.SHOW_TEXT);
    const nodes  = [];
    let node;
    while (node = walker.nextNode()) nodes.push(node);
    for (const textNode of nodes) {
      const parent = textNode.parentNode;
      if (!parent || parent.closest('code, pre, a')) continue;
      const text    = textNode.textContent;
      const result = matchDict(text, dictEntries, depth);
      if (!result) continue;
      const span = document.createElement('span');
      span.innerHTML = result;
      parent.replaceChild(span, textNode);}}
// 
    function tokensMatch(a, b, depth) {
    a = a.toLowerCase();
    b = b.toLowerCase();
    if (depth === 0) return a === b;
    const minLen = Math.min(a.length, b.length);
    const prefixLen = Math.max(minLen - depth, 1);
    return a.slice(0, prefixLen) === b.slice(0, prefixLen);}
// 
  function phraseMatch(textTokens, keyTokens, depth) {
    if (textTokens.length !== keyTokens.length) return false;
    return textTokens.every((t, i) => tokensMatch(t, keyTokens[i], depth)); }
//  
  function matchDict(text, dictEntries, depth) {
    const words  = text.split(/(\s+)/);
    let result   = '';
    let changed  = false;
    let wi       = 0;
    while (wi < words.length) {
      if (/^\s+$/.test(words[wi])) { result += words[wi]; wi++; continue; }
      let matched = false;
      for (const {keyTokens, original, value} of dictEntries) {
        const len = keyTokens.length;
        const phrase = [];
        const indices = [];
        let k = wi;
        while (phrase.length < len && k < words.length) {
          if (!/^\s+$/.test(words[k])) { phrase.push(words[k]); indices.push(k); }
          k++;}
        if (phrase.length < len) continue;
        if (phraseMatch(phrase, keyTokens, depth)) {
          const raw = words.slice(wi, k).join('');
          result += `<span style="border-bottom:1px dotted var(--ink);cursor:help" title="${esc(original)}: ${esc(value)}">${esc(raw)}</span>`;
          wi      = k;
          matched = true;
          changed = true;
          break;}}
      if (!matched) { result += words[wi]; wi++; }}
    return changed ? result : null;} }