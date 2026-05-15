// web/js/lens-tags.js — линза тегов

import { api, esc, buildCharOrder, compareByASort, gaussOrder } from './utils.js';
export function initTagsLens(ctx) {
  const { projectId, frameIndex, getPairs, getMetaPairs, getAliasBefore, onSave } = ctx;
  let tagsCache  = [];
  let allFrames  = [];
  let globalTags = [];

  document.getElementById('tagsToggleBtn').addEventListener('click', async () => {
    const container = document.getElementById('tagsContainer');
    const btn       = document.getElementById('tagsToggleBtn');
    if (!container.classList.contains('hidden')) {
      container.classList.add('hidden');
      btn.textContent = '⬡';
      btn.classList.remove('active');
      return; }
    container.classList.remove('hidden');
    btn.textContent = '⬡';
    btn.classList.add('active');
    await loadTagsPanel(container);});
// 
  async function loadTagsPanel(container) {
    try { tagsCache  = await api.getTags(projectId); }
    catch(e) { tagsCache = []; }
    const aSort = ctx.getASort ? ctx.getASort() : '';
    if (aSort) {
      const order = buildCharOrder(aSort);
      tagsCache.sort((a, b) => compareByASort(a.title || '', b.title || '', order)); }
    try { allFrames  = await api.getFrames(projectId); }
    catch(e) { allFrames = []; }
    try { globalTags = await api.getGlobalTags(); }
    catch(e) { globalTags = []; }
    const tagsPair       = getPairs().find(p => p.Left === 'tags');
    const currentIndices = tagsPair
      ? tagsPair.Right.trim().split(/\s+/).filter(Boolean)
      : [];
    currentIndices.sort((a, b) => {
      const ta = (tagsCache.find(t => t.index === a) || {}).title || a;
      const tb = (tagsCache.find(t => t.index === b) || {}).title || b;
      const aSort = ctx.getASort ? ctx.getASort() : '';
      if (!aSort) return 0;
    return compareByASort(ta, tb, buildCharOrder(aSort));});
    renderTagsPanel(container, currentIndices); }
//
  function resolveTitle(index) {
    const inTags = tagsCache.find(t => t.index === index);
    if (inTags) return inTags.title;
    const inAll  = allFrames.find(f => f.index === index);
    if (inAll)   return inAll.title;
    return index; }
//
  function getContentOfTag(tagIndex) {
    const slot = getMetaPairs().find(p => p.Left === 'tags');
    if (!slot) return [];
    const blocks = slot.Right.match(/\[([^\]]*)\]/g) || [];
    const pos = gaussOrder(tagIndex);
    if (pos >= blocks.length) return [];
    return blocks[pos].slice(1,-1).trim().split(/\s+/).filter(Boolean); }
//
  function renderTagsPanel(container, currentIndices) {
    container.innerHTML = `
      <div class="tags-panel">
        <div class="tags-panel-pills" id="tagsPills"></div>
        <div class="tags-panel-input-wrap">
          <input class="tags-panel-input" id="tagsInput"
            type="text" placeholder="add tag..."
            autocomplete="off" spellcheck="false">
          <div class="tags-autocomplete" id="tagsAutocomplete"></div>
        </div>
      </div>`;
    const pillsEl  = document.getElementById('tagsPills');
    const input    = document.getElementById('tagsInput');
    const dropdown = document.getElementById('tagsAutocomplete');
    function makePill(index) {
      const isGlobal  = index.startsWith('meta-base:');
      const realIndex = isGlobal ? index.replace('meta-base:', '') : index;
      const tag   = isGlobal
        ? globalTags.find(t => t.index === realIndex)
        : tagsCache.find(t => t.index === realIndex);
      const color = tag?.color || '';
      const icon  = tag?.icon  || '';
      const title = tag ? tag.title : realIndex;
      const pill = document.createElement('span');
      pill.className     = 'tag-pill' + (isGlobal ? ' tag-pill-global' : '');
      pill.dataset.index = index;
      if (color) pill.style.borderColor = color;
      const iconHtml = icon
        ? `<img class="tag-pill-icon" src="/img/${projectId}/${icon}"
            onerror="this.style.display='none'">`
        : (isGlobal ? '<span class="tag-pill-global-marker">⬡</span>' : '');
      pill.innerHTML =
          iconHtml
        + `<span class="tag-pill-title">${esc(title)}</span>`
        + `<span class="tag-pill-index">${esc(realIndex)}</span>`
        + `<span class="tag-pill-expand" title="frames for tag">▾</span>`
        + `<span class="tag-pill-remove">×</span>`;
      pill.querySelector('.tag-pill-remove')
          .addEventListener('click', e => {
            e.stopPropagation();
            pill.remove();
            autoSave();});
      pill.querySelector('.tag-pill-title')
          .addEventListener('click', () => {
            if (isGlobal) {
              window.location.href =
                'frame.html?id=00000-META-BASE&frame=' + realIndex;
            } else {
              window.location.href =
                'frame.html?id=' + projectId + '&frame=' + realIndex;}});
      pill.querySelector('.tag-pill-expand')
          .addEventListener('click', e => {
            e.stopPropagation();
            const existing = pill.querySelector('.tag-pill-list');
            if (existing) { existing.remove(); return; }
            document.querySelectorAll('.tag-pill-list')
                    .forEach(el => el.remove());
            const members = isGlobal
              ? []
              : getContentOfTag(realIndex);
            const list = document.createElement('div');
            list.className = 'tag-pill-list';
            if (members.length === 0) {
              list.innerHTML =
                '<span class="tag-pill-list-empty">not list</span>';
            } else {
              const aSort = ctx.getASort ? ctx.getASort() : '';
              if (aSort) {
                const order = buildCharOrder(aSort);
                members.sort((a, b) => compareByASort(
                  resolveTitle(a), resolveTitle(b), order));}
              members.forEach(memberIdx => {
                const item = document.createElement('a');
                item.className = 'tag-pill-list-item';
                item.href =
                  'frame.html?id=' + projectId + '&frame=' + memberIdx;
                const isMemberTag =
                  allFrames.find(f => f.index === memberIdx)?.isTag || false;
                const marker = isMemberTag ? '⬡' : '▪';
                item.innerHTML =
                    `<span class="tag-pill-list-marker">${marker}</span>`
                  + `<span class="tag-pill-list-title">`
                  + `${esc(resolveTitle(memberIdx))}</span>`
                  + `<span class="tag-pill-list-index">`
                  + `${esc(memberIdx)}</span>`;
                list.appendChild(item);});}
            pill.appendChild(list);});
      return pill;}
      // 
    function renderPills(indices) {
      pillsEl.innerHTML = '';
      indices.forEach(idx => pillsEl.appendChild(makePill(idx))); }
      // 
    function addPill(index) {
      const used = new Set(
        [...pillsEl.querySelectorAll('.tag-pill')].map(p => p.dataset.index));
      if (!used.has(index)) pillsEl.appendChild(makePill(index));
        autoSave(); }
    renderPills(currentIndices);
      // 
  async function autoSave() {
    const indices = [...pillsEl.querySelectorAll('.tag-pill')]
      .map(p => p.dataset.index);
    const pairs = getPairs();
    for (let p of pairs) {
      if (p.Left === 'tags')     { p.Right = indices.join(' '); }
      if (p.Left === 'modified') { p.Right = new Date().toISOString().slice(0,10); }}
    await api.saveFrame(projectId, frameIndex, pairs, getAliasBefore());
    const fresh = await api.getFrame(projectId, frameIndex);
    onSave(fresh.pairs || []);}
    // Автодополнение
    input.addEventListener('input', () => {
      const val = input.value.toLowerCase().trim();
      dropdown.innerHTML = '';
      if (!val) { dropdown.style.display = 'none'; return; }
      const used = new Set(
        [...pillsEl.querySelectorAll('.tag-pill')].map(p => p.dataset.index));
      const matches = tagsCache.filter(t =>
        t.title.toLowerCase().includes(val) && !used.has(t.index));
      dropdown.style.display = 'block';
      if (matches.length > 0) {
        matches.slice(0, 26).forEach(tag => {
          const item = document.createElement('div');
          item.className = 'tags-autocomplete-item';
          item.innerHTML =
              `<span>${esc(tag.title)}</span>`
            + `<span class="tags-autocomplete-index">${esc(tag.index)}</span>`;
          item.onmousedown = e => {
            e.preventDefault();
            addPill(tag.index);
            input.value = '';
            dropdown.style.display = 'none';};
          dropdown.appendChild(item);});}
      // Глобальные теги
      const globalMatches = globalTags.filter(t =>
        t.title.toLowerCase().includes(val) &&
        !tagsCache.find(m => m.title.toLowerCase() === t.title.toLowerCase()) &&
        !used.has('meta-base:' + t.index) &&
        ![...pillsEl.querySelectorAll('.tag-pill')]
            .find(p => {
            const pillIndex = p.dataset.index;
            const pillTitle = pillIndex.startsWith('meta-base:')
                ? (globalTags.find(g => g.index === pillIndex.replace('meta-base:',''))?.title || '')
                : resolveTitle(pillIndex);
            return pillTitle.toLowerCase() === t.title.toLowerCase();}));
      if (globalMatches.length > 0) {
        const sep = document.createElement('div');
        sep.className   = 'tags-autocomplete-sep';
        sep.textContent = 'global';
        dropdown.appendChild(sep);
        globalMatches.slice(0, 10).forEach(tag => {
          const item = document.createElement('div');
          item.className = 'tags-autocomplete-item tags-autocomplete-global';
          item.innerHTML =
              `<span>⬡ ${esc(tag.title)}</span>`
            + `<span class="tags-autocomplete-index">${esc(tag.index)}</span>`;
          item.onmousedown = e => {
            e.preventDefault();
            addPill('meta-base:' + tag.index);
            input.value = '';
            dropdown.style.display = 'none';};
          dropdown.appendChild(item);});}
      const existingFrame = allFrames.find(f => f.title.toLowerCase() === val);
      if (existingFrame && existingFrame.isTag) return;
      const newItem = document.createElement('div');
      newItem.className = 'tags-autocomplete-item tags-autocomplete-new';
      if (existingFrame && !existingFrame.isTag) {
        newItem.innerHTML =
          `<span>сделать тегом «${esc(existingFrame.title)}»</span>`;
        newItem.onmousedown = async e => {
          e.preventDefault();
          input.value    = '';
          input.disabled = true;
          dropdown.style.display = 'none';
          try {
            const freshFrame = await api.getFrame(projectId, existingFrame.index);
            const newPairs   = freshFrame.pairs.map(p =>
              p.Left === 'rank' ? { ...p, Right: '0' } : p);
            await api.saveFrame(projectId, existingFrame.index, newPairs, []);
            existingFrame.isTag = true;
            tagsCache.push({index: existingFrame.index, title: existingFrame.title, color: '', icon: ''});
            addPill(existingFrame.index);
          } catch(err) { console.error('make tag:', err);}
          input.disabled = false;};
      } else {
        newItem.innerHTML =
          `<span>создать тег «${esc(input.value.trim())}»</span>`;
        newItem.onmousedown = async e => {
          e.preventDefault();
          const newTitle = input.value.trim();
          if (!newTitle) return;
          input.value    = '';
          input.disabled = true;
          dropdown.style.display = 'none';
          try {
            const r = await fetch('/api/projects/' + projectId + '/frames', {
              method: 'POST', headers: {'Content-Type': 'application/json'},
              body: JSON.stringify({title: newTitle})});
            const newFrame   = await r.json();
            const freshFrame = await api.getFrame(projectId, newFrame.index);
            const newPairs   = freshFrame.pairs.map(p =>
              p.Left === 'rank' ? {...p, Right: '0'} : p);
            await api.saveFrame(projectId, newFrame.index, newPairs, []);
            tagsCache.push({index: newFrame.index, title: newTitle, color: '', icon: ''});
            allFrames.push({index: newFrame.index, title: newTitle, isTag: true});
            addPill(newFrame.index);
          } catch(err) { console.error('create tag:', err);}
          input.disabled = false;};}
      dropdown.appendChild(newItem);});
        input.addEventListener('blur', () => {
      setTimeout(() => {dropdown.style.display = 'none';}, 150);});} }